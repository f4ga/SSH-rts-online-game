package network

import (
	"context"
	"fmt"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
	"ssh-arena-app/internal"
)

// SSHServer реализует internal.NetworkTransport через SSH.
type SSHServer struct {
	config      *ssh.ServerConfig
	listener    net.Listener
	hub         *Hub
	cmdChan     chan internal.ClientMessage
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	serverAddr  string
	playerAdder func(playerID string) // вызывается при подключении нового игрока
}

// NewSSHServer создаёт новый SSH-сервер.
func NewSSHServer(hostKey ssh.Signer, port int, playerAdder func(playerID string)) (*SSHServer, error) {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			// Принимаем любой ключ (для демо). В реальности нужно проверять.
			return &ssh.Permissions{
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(key),
				},
			}, nil
		},
		// Опционально разрешаем аутентификацию по паролю для тестов.
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			// Принимаем любой пароль (небезопасно). В продакшене отключить.
			return nil, nil
		},
	}
	config.AddHostKey(hostKey)

	ctx, cancel := context.WithCancel(context.Background())
	return &SSHServer{
		config:      config,
		hub:         NewHub(),
		cmdChan:     make(chan internal.ClientMessage, 100),
		ctx:         ctx,
		cancel:      cancel,
		serverAddr:  fmt.Sprintf(":%d", port),
		playerAdder: playerAdder,
	}, nil
}

// Start запускает сервер.
func (s *SSHServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.serverAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.serverAddr, err)
	}

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop останавливает сервер.
func (s *SSHServer) Stop() error {
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}
	s.hub.Close()
	close(s.cmdChan)
	s.wg.Wait()
	return nil
}

// Send отправляет сообщение конкретному клиенту.
func (s *SSHServer) Send(clientID string, msg []byte) error {
	return s.hub.Send(clientID, msg)
}

// Broadcast рассылает сообщение всем клиентам.
func (s *SSHServer) Broadcast(msg []byte) {
	s.hub.Broadcast(msg)
}

// Receive возвращает канал для получения команд от клиентов.
func (s *SSHServer) Receive() <-chan internal.ClientMessage {
	return s.cmdChan
}

// acceptLoop принимает входящие соединения.
func (s *SSHServer) acceptLoop() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}
		conn, err := s.listener.Accept()
		if err != nil {
			// Если контекст отменён, это ожидаемо.
			if s.ctx.Err() != nil {
				return
			}
			continue
		}
		go s.handleConn(conn)
	}
}

// handleConn обрабатывает SSH-соединение.
func (s *SSHServer) handleConn(conn net.Conn) {
	defer conn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		return
	}
	defer sshConn.Close()

	// Обрабатываем out-of-band запросы.
	go ssh.DiscardRequests(reqs)

	// Ожидаем каналы (сессии).
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(sshConn, channel, requests)
	}
}

// terminalState хранит состояние терминала клиента.
type terminalState struct {
	width  int
	height int
	// можно добавить флаг raw mode
}

// handleSession обрабатывает SSH-сессию.
func (s *SSHServer) handleSession(conn *ssh.ServerConn, channel ssh.Channel, requests <-chan *ssh.Request) {
	clientID := conn.RemoteAddr().String() // временный ID, лучше использовать fingerprint
	session := NewSession(clientID, channel)
	s.hub.Register(clientID, session)
	defer s.hub.Unregister(clientID)

	// Если задан playerAdder, добавляем игрока в игру.
	if s.playerAdder != nil {
		s.playerAdder(clientID)
	}

	// Отправляем приветственное сообщение.
	welcome := []byte("Welcome to SSH Arena! Use WASD to move, /help for commands.\r\n")
	channel.Write(welcome)

	var term terminalState

	// Обрабатываем запросы псевдотерминала.
	for req := range requests {
		switch req.Type {
		case "pty-req":
			// Сохраняем размер терминала (опционально)
			if len(req.Payload) >= 8 {
				term.width = int(req.Payload[0])<<24 | int(req.Payload[1])<<16 | int(req.Payload[2])<<8 | int(req.Payload[3])
				term.height = int(req.Payload[4])<<24 | int(req.Payload[5])<<16 | int(req.Payload[6])<<8 | int(req.Payload[7])
			}
			req.Reply(true, nil)
		case "shell":
			req.Reply(true, nil)
			// После shell больше не ожидаем запросов (но могут быть window-change)
			break
		case "window-change":
			// Обновляем размер терминала
			if len(req.Payload) >= 8 {
				term.width = int(req.Payload[0])<<24 | int(req.Payload[1])<<16 | int(req.Payload[2])<<8 | int(req.Payload[3])
				term.height = int(req.Payload[4])<<24 | int(req.Payload[5])<<16 | int(req.Payload[6])<<8 | int(req.Payload[7])
			}
			req.Reply(true, nil)
		default:
			// Игнорируем другие запросы
			req.Reply(false, nil)
		}
	}

	// Читаем ввод от клиента и отправляем в cmdChan.
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := channel.Read(buf)
			if err != nil {
				break
			}
			// Преобразуем ввод с учётом маппинга клавиш
			transformed := TransformInput(buf[:n])
			if transformed == "" {
				continue
			}
			clientMsg := internal.ClientMessage{
				ClientID: clientID,
				Data:     []byte(transformed),
			}
			select {
			case s.cmdChan <- clientMsg:
			default:
				// Если канал переполнен, игнорируем.
			}
		}
	}()

	// Ждём закрытия канала.
	<-s.ctx.Done()
}
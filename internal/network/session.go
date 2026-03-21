package network

import (
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Session представляет SSH-сессию клиента.
type Session struct {
	ID        string
	channel   ssh.Channel
	mu        sync.Mutex
	createdAt time.Time
	lastActive time.Time
}

// NewSession создаёт новую сессию.
func NewSession(id string, channel ssh.Channel) *Session {
	now := time.Now()
	return &Session{
		ID:         id,
		channel:    channel,
		createdAt:  now,
		lastActive: now,
	}
}

// Write отправляет данные клиенту.
func (s *Session) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActive = time.Now()
	return s.channel.Write(data)
}

// Close закрывает сессию.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.channel.Close()
}

// IsActive проверяет, активна ли сессия (по таймауту).
func (s *Session) IsActive(timeout time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.lastActive) < timeout
}
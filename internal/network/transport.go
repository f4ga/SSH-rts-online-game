package network

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"ssh-arena-app/internal"

	"golang.org/x/crypto/ssh"
)

// NetworkTransport возвращает реализацию NetworkTransport (SSH-сервер) без добавления игроков.
func NetworkTransport() (internal.NetworkTransport, error) {
	return NewSSHTransport(nil)
}

// NewSSHTransport создаёт SSH-сервер с опциональным коллбэком для добавления игроков.
func NewSSHTransport(playerAdder func(playerID string)) (internal.NetworkTransport, error) {
	// Пытаемся загрузить существующий ключ хоста из файла.
	var signer ssh.Signer
	keyPath := "ssh_host_key"
	if _, err := os.Stat(keyPath); err == nil {
		keyBytes, err := os.ReadFile(keyPath)
		if err == nil {
			signer, err = ssh.ParsePrivateKey(keyBytes)
			if err == nil {
				// Успешно загружен
			}
		}
	}
	// Если не удалось загрузить, генерируем новый.
	if signer == nil {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate host key: %w", err)
		}
		signer, err = ssh.NewSignerFromKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create signer: %w", err)
		}
		// Сохраняем ключ для будущих запусков (опционально)
		// (пропустим для простоты)
	}

	// Порт по умолчанию 2222.
	server, err := NewSSHServer(signer, 2222, playerAdder)
	if err != nil {
		return nil, err
	}
	return server, nil
}

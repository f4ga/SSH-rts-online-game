package network

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"ssh-arena-app/internal"

	"golang.org/x/crypto/ssh"
)

// NetworkTransport возвращает реализацию NetworkTransport (SSH-сервер).
func NetworkTransport() (internal.NetworkTransport, error) {
	// Генерируем временный RSA-ключ для хоста (в продакшене нужно загружать из файла).
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %w", err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// Порт по умолчанию 2222.
	server, err := NewSSHServer(signer, 2222)
	if err != nil {
		return nil, err
	}
	return server, nil
}

package network

import (
	"fmt"
	"strings"
	"ssh-arena-app/internal"
)

// ParseCommand преобразует сырые байты в команду.
func ParseCommand(data []byte) (internal.Command, error) {
	// Убираем пробелы и символы новой строки.
	line := strings.TrimSpace(string(data))
	if line == "" {
		return internal.Command{}, ErrEmptyCommand
	}

	// Если строка начинается с '/', это игровая команда.
	if strings.HasPrefix(line, "/") {
		return parseGameCommand(line)
	}

	// Иначе считаем это текстовым сообщением (чат).
	return internal.Command{
		Type: "chat",
		Payload: map[string]interface{}{
			"text": line,
		},
	}, nil
}

// parseGameCommand разбирает игровую команду вида /command arg1 arg2 ...
func parseGameCommand(line string) (internal.Command, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return internal.Command{}, ErrInvalidCommand
	}
	// Убираем ведущий '/'
	cmdType := parts[0][1:]
	args := parts[1:]

	payload := make(map[string]interface{})
	for i, arg := range args {
		key := fmt.Sprintf("arg%d", i)
		payload[key] = arg
	}
	payload["raw"] = strings.Join(args, " ")

	return internal.Command{
		Type:    cmdType,
		Payload: payload,
	}, nil
}

// ErrEmptyCommand ошибка пустой команды.
var ErrEmptyCommand = &ProtocolError{"empty command"}

// ErrInvalidCommand ошибка неверной команды.
var ErrInvalidCommand = &ProtocolError{"invalid command"}

// ProtocolError представляет ошибку протокола.
type ProtocolError struct {
	msg string
}

func (e *ProtocolError) Error() string {
	return e.msg
}
package ui

import (
	"bytes"
	"ssh-arena-app/internal"
)

// ParseInput преобразует сырые байты из терминала в команды.
// Поддерживает стрелки и функциональные клавиши через ANSI escape-последовательности.
func ParseInput(data []byte) []internal.Command {
	var commands []internal.Command

	// Разделяем по символам новой строки (обычно команды отправляются по Enter).
	lines := bytes.Split(data, []byte{'\n'})
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Проверяем escape-последовательности.
		if bytes.HasPrefix(line, []byte{0x1b}) {
			cmd := parseEscape(line)
			if cmd.Type != "" {
				commands = append(commands, cmd)
			}
			continue
		}

		// Обычный текст.
		commands = append(commands, internal.Command{
			Type: "raw",
			Payload: map[string]interface{}{
				"text": string(line),
			},
		})
	}
	return commands
}

// parseEscape разбирает ANSI escape-последовательность.
func parseEscape(data []byte) internal.Command {
	// Простейшие стрелки: \x1b[A, \x1b[B, \x1b[C, \x1b[D
	if len(data) >= 3 && data[0] == 0x1b && data[1] == '[' {
		switch data[2] {
		case 'A':
			return internal.Command{Type: "move_up", Payload: nil}
		case 'B':
			return internal.Command{Type: "move_down", Payload: nil}
		case 'C':
			return internal.Command{Type: "move_right", Payload: nil}
		case 'D':
			return internal.Command{Type: "move_left", Payload: nil}
		}
	}
	// Игнорируем другие последовательности.
	return internal.Command{}
}
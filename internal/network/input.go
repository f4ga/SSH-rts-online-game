package network

import (
	"strings"

	"ssh-arena-app/internal"
)

// KeyMapping определяет соответствие одиночных клавиш игровым командам.
var KeyMapping = map[byte]string{
	'w': "/move 0 -1",
	's': "/move 0 1",
	'a': "/move -1 0",
	'd': "/move 1 0",
	'e': "/interact",
	'f': "/attack",
	'q': "/cancel",
	' ': "/wait",
	// стрелки могут быть представлены escape-последовательностями, обрабатываются отдельно
}

// TransformInput преобразует ввод от клиента в строку команды.
// Если ввод — одиночный символ, присутствующий в KeyMapping, возвращает соответствующую команду.
// Иначе возвращает исходный ввод как строку.
func TransformInput(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	// Убираем пробельные символы в начале и конце
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return ""
	}
	// Если ввод состоит из одного символа (игнорируя возможный символ новой строки)
	// и этот символ есть в маппинге, преобразуем.
	if len(trimmed) == 1 {
		key := trimmed[0]
		if cmdStr, ok := KeyMapping[key]; ok {
			return cmdStr
		}
	}
	// Иначе возвращаем исходный ввод (как строку)
	return string(data)
}

// ParseInput преобразует ввод от клиента в команду.
// Использует TransformInput для маппинга клавиш, затем парсит через ParseCommand.
func ParseInput(data []byte) (internal.Command, error) {
	transformed := TransformInput(data)
	if transformed == "" {
		return internal.Command{}, ErrEmptyCommand
	}
	return ParseCommand([]byte(transformed))
}
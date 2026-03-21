package ui

import (
	"fmt"
	"strings"
	"time"
	"ssh-arena-app/internal"
)

// StatusRenderer генерирует строку статусной панели.
type StatusRenderer struct {
	width int
}

// NewStatusRenderer создаёт новый рендерер статуса.
func NewStatusRenderer(width int) *StatusRenderer {
	return &StatusRenderer{width: width}
}

// Render создаёт строку статуса для игрока.
func (sr *StatusRenderer) Render(player *internal.Player, resources map[string]int, citizens int, research string, notifications []string) string {
	// Верхняя линия.
	line := strings.Repeat("=", sr.width) + "\n"

	// Строка с ресурсами.
	resStr := "Resources: "
	for typ, amt := range resources {
		resStr += fmt.Sprintf("%s:%d ", typ, amt)
	}
	resStr = truncate(resStr, sr.width)

	// Строка с информацией игрока.
	playerStr := fmt.Sprintf("Player: %s | Credits: %d | Citizens: %d | Research: %s",
		player.Name, player.Credits, citizens, research)
	playerStr = truncate(playerStr, sr.width)

	// Строка времени.
	timeStr := fmt.Sprintf("Time: %s", time.Now().Format("15:04:05"))
	timeStr = truncate(timeStr, sr.width)

	// Уведомления.
	notifStr := ""
	if len(notifications) > 0 {
		notifStr = "Notifications: " + strings.Join(notifications, "; ")
		notifStr = truncate(notifStr, sr.width)
	}

	// Собираем всё.
	lines := []string{line, resStr, playerStr, timeStr}
	if notifStr != "" {
		lines = append(lines, notifStr)
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

// truncate обрезает строку до заданной ширины.
func truncate(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}
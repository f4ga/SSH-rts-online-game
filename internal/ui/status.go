package ui

import (
	"fmt"
	"strings"
	"time"

	"ssh-arena-app/internal"
)

// StatusRenderer генерирует строку статусной панели.
type StatusRenderer struct {
	width      int
	useColor   bool
	colorScheme *ColorScheme
}

// NewStatusRenderer создаёт новый рендерер статуса.
func NewStatusRenderer(width int) *StatusRenderer {
	return &StatusRenderer{
		width:      width,
		useColor:   true,
		colorScheme: DefaultColorScheme(),
	}
}

// Render создаёт строку статуса для игрока.
func (sr *StatusRenderer) Render(player *internal.Player, resources map[string]int, citizens int, research string, notifications []string) string {
	var out strings.Builder

	// Верхняя линия.
	out.WriteString(sr.separator("="))

	// Первая строка: имя игрока, кредиты, жители, время.
	timeStr := time.Now().Format("15:04:05")
	playerLine := fmt.Sprintf(" %s | Credits: %d | Citizens: %d | Time: %s ",
		player.Name, player.Credits, citizens, timeStr)
	playerLine = sr.padCenter(playerLine)
	out.WriteString(playerLine)
	out.WriteString("\n")

	// Вторая строка: ресурсы.
	resLine := "Resources: "
	if len(resources) == 0 {
		resLine += "none"
	} else {
		for typ, amt := range resources {
			icon := sr.resourceIcon(typ)
			resLine += fmt.Sprintf("%s %d  ", icon, amt)
		}
	}
	resLine = sr.padCenter(resLine)
	out.WriteString(resLine)
	out.WriteString("\n")

	// Третья строка: исследования.
	researchLine := "Research: "
	if research == "" {
		researchLine += "none"
	} else {
		researchLine += research
		// Можно добавить прогресс-бар, но пока просто текст.
	}
	researchLine = sr.padCenter(researchLine)
	out.WriteString(researchLine)
	out.WriteString("\n")

	// Уведомления.
	if len(notifications) > 0 {
		notifLine := "Notifications: " + strings.Join(notifications, " | ")
		notifLine = sr.padCenter(notifLine)
		out.WriteString(notifLine)
		out.WriteString("\n")
	}

	// Нижняя линия.
	out.WriteString(sr.separator("="))

	return out.String()
}

// separator возвращает разделитель на всю ширину.
func (sr *StatusRenderer) separator(char string) string {
	return strings.Repeat(char, sr.width) + "\n"
}

// padCenter центрирует текст, добавляя пробелы по бокам.
func (sr *StatusRenderer) padCenter(s string) string {
	if len(s) >= sr.width {
		return s[:sr.width]
	}
	padding := sr.width - len(s)
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// resourceIcon возвращает символ для типа ресурса.
func (sr *StatusRenderer) resourceIcon(resType string) string {
	icons := map[string]string{
		"wood":  "♣",
		"stone": "◼",
		"iron":  "⚙",
		"food":  "🍎",
		"gold":  "💰",
	}
	if icon, ok := icons[resType]; ok {
		return icon
	}
	return "?"
}

// SetUseColor включает/выключает цвет.
func (sr *StatusRenderer) SetUseColor(use bool) {
	sr.useColor = use
}
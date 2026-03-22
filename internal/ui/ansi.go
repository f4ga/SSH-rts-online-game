package ui

import (
	"fmt"
)

// Color представляет ANSI-цвет (3/4 бита).
type Color int

// Стандартные цвета (3/4 бита).
const (
	ColorBlack Color = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// BrightColor возвращает яркую версию цвета.
func BrightColor(c Color) Color {
	return c + 60
}

// BackgroundColor возвращает цвет фона.
func BackgroundColor(c Color) Color {
	return c + 10
}

// ANSI-последовательности.
const (
	escape      = "\x1b["
	reset       = escape + "0m"
	clearScreen = escape + "2J"
	home        = escape + "H"
	saveCursor  = escape + "s"
	restoreCursor = escape + "u"
)

// ClearScreen возвращает последовательность очистки экрана.
func ClearScreen() string {
	return clearScreen + home
}

// MoveCursor возвращает последовательность перемещения курсора.
func MoveCursor(row, col int) string {
	return fmt.Sprintf("%s%d;%dH", escape, row, col)
}

// SetColor возвращает последовательность установки цвета текста и фона (3/4 бита).
func SetColor(fg, bg Color) string {
	if bg == 0 {
		return fmt.Sprintf("%s%dm", escape, fg)
	}
	return fmt.Sprintf("%s%d;%dm", escape, fg, bg)
}

// SetTrueColorForeground возвращает последовательность установки цвета текста в true color (24-bit).
func SetTrueColorForeground(r, g, b int) string {
	return fmt.Sprintf("%s38;2;%d;%d;%dm", escape, r, g, b)
}

// SetTrueColorBackground возвращает последовательность установки цвета фона в true color (24-bit).
func SetTrueColorBackground(r, g, b int) string {
	return fmt.Sprintf("%s48;2;%d;%d;%dm", escape, r, g, b)
}

// SetTrueColor устанавливает и foreground, и background true color.
func SetTrueColor(fgR, fgG, fgB, bgR, bgG, bgB int) string {
	return SetTrueColorForeground(fgR, fgG, fgB) + SetTrueColorBackground(bgR, bgG, bgB)
}

// ResetColor возвращает последовательность сброса цвета.
func ResetColor() string {
	return reset
}

// ColorCode возвращает ANSI-код для цвета (3/4 бита).
func ColorCode(c Color) string {
	return fmt.Sprintf("%s%dm", escape, c)
}

// SaveCursor сохраняет позицию курсора.
func SaveCursor() string {
	return saveCursor
}

// RestoreCursor восстанавливает позицию курсора.
func RestoreCursor() string {
	return restoreCursor
}

// HideCursor скрывает курсор.
func HideCursor() string {
	return escape + "?25l"
}

// ShowCursor показывает курсор.
func ShowCursor() string {
	return escape + "?25h"
}
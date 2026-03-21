package ui

// Color представляет ANSI-цвет.
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
)

// ClearScreen возвращает последовательность очистки экрана.
func ClearScreen() string {
	return clearScreen + home
}

// MoveCursor возвращает последовательность перемещения курсора.
func MoveCursor(row, col int) string {
	return escape + string(rune(row)) + ";" + string(rune(col)) + "H"
}

// SetColor возвращает последовательность установки цвета текста и фона.
func SetColor(fg, bg Color) string {
	if bg == 0 {
		return escape + string(rune(fg)) + "m"
	}
	return escape + string(rune(fg)) + ";" + string(rune(bg)) + "m"
}

// ResetColor возвращает последовательность сброса цвета.
func ResetColor() string {
	return reset
}

// ColorCode возвращает ANSI-код для цвета.
func ColorCode(c Color) string {
	return escape + string(rune(c)) + "m"
}
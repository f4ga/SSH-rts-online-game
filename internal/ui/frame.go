package ui

import (
	"strings"
)

// FrameStyle определяет стиль рамки.
type FrameStyle struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
}

// Стили рамок.
var (
	// StyleThin - тонкая рамка (Unicode).
	StyleThin = FrameStyle{
		TopLeft:     '┌',
		TopRight:    '┐',
		BottomLeft:  '└',
		BottomRight: '┘',
		Horizontal:  '─',
		Vertical:    '│',
	}
	// StyleThick - толстая рамка (Unicode).
	StyleThick = FrameStyle{
		TopLeft:     '┏',
		TopRight:    '┓',
		BottomLeft:  '┗',
		BottomRight: '┛',
		Horizontal:  '━',
		Vertical:    '┃',
	}
	// StyleDouble - двойная рамка (Unicode).
	StyleDouble = FrameStyle{
		TopLeft:     '╔',
		TopRight:    '╗',
		BottomLeft:  '╚',
		BottomRight: '╝',
		Horizontal:  '═',
		Vertical:    '║',
	}
	// StyleASCII - рамка из ASCII символов.
	StyleASCII = FrameStyle{
		TopLeft:     '+',
		TopRight:    '+',
		BottomLeft:  '+',
		BottomRight: '+',
		Horizontal:  '-',
		Vertical:    '|',
	}
)

// Frame рисует рамку вокруг области.
type Frame struct {
	Width      int          // внутренняя ширина (без границ)
	Height     int          // внутренняя высота (без границ)
	Style      FrameStyle   // стиль рамки
	Title      string       // заголовок (если пусто, не отображается)
	TitleColor ColorRGB     // цвет заголовка
	BorderColor ColorRGB    // цвет рамки
	FillColor  ColorRGB     // цвет фона внутри (не используется, но можно)
}

// NewFrame создаёт новую рамку.
func NewFrame(width, height int) *Frame {
	return &Frame{
		Width:      width,
		Height:     height,
		Style:      StyleThin,
		Title:      "",
		TitleColor: ColorCitizen, // белый
		BorderColor: ColorGray,
		FillColor:  ColorDark,    // почти чёрный
	}
}

// SetStyle устанавливает стиль рамки.
func (f *Frame) SetStyle(style FrameStyle) {
	f.Style = style
}

// SetTitle устанавливает заголовок.
func (f *Frame) SetTitle(title string) {
	f.Title = title
}

// SetColors устанавливает цвета.
func (f *Frame) SetColors(border, title, fill ColorRGB) {
	f.BorderColor = border
	f.TitleColor = title
	f.FillColor = fill
}

// Render возвращает строку, представляющую рамку с содержимым.
// Параметр contentFunc вызывается для каждой внутренней клетки (x, y) и должна вернуть символ и цвета.
// Если contentFunc == nil, внутренность заполняется пробелами.
func (f *Frame) Render(contentFunc func(x, y int) (rune, ColorRGB, ColorRGB)) []string {
	// Внешние размеры с учётом рамки.
	outerWidth := f.Width + 2  // +2 для левой и правой границ
	outerHeight := f.Height + 2 // +2 для верхней и нижней границ

	// Создаём двумерный массив рун.
	grid := make([][]rune, outerHeight)
	for i := range grid {
		grid[i] = make([]rune, outerWidth)
	}

	// Заполняем фон пробелами.
	for y := 0; y < outerHeight; y++ {
		for x := 0; x < outerWidth; x++ {
			grid[y][x] = ' '
		}
	}

	// Рисуем углы.
	grid[0][0] = f.Style.TopLeft
	grid[0][outerWidth-1] = f.Style.TopRight
	grid[outerHeight-1][0] = f.Style.BottomLeft
	grid[outerHeight-1][outerWidth-1] = f.Style.BottomRight

	// Рисуем горизонтальные линии.
	for x := 1; x < outerWidth-1; x++ {
		grid[0][x] = f.Style.Horizontal
		grid[outerHeight-1][x] = f.Style.Horizontal
	}

	// Рисуем вертикальные линии.
	for y := 1; y < outerHeight-1; y++ {
		grid[y][0] = f.Style.Vertical
		grid[y][outerWidth-1] = f.Style.Vertical
	}

	// Добавляем заголовок, если есть.
	if f.Title != "" && outerWidth > 2 {
		titleLen := len(f.Title)
		// Ограничиваем длину заголовка, чтобы не вылезать за рамки.
		if titleLen > outerWidth-2 {
			titleLen = outerWidth - 2
		}
		start := (outerWidth - titleLen) / 2
		for i, ch := range f.Title[:titleLen] {
			grid[0][start+i] = ch
		}
	}

	// Заполняем внутренность.
	if contentFunc != nil {
		for y := 0; y < f.Height; y++ {
			for x := 0; x < f.Width; x++ {
				ch, _, _ := contentFunc(x, y)
				grid[y+1][x+1] = ch
			}
		}
	}

	// Преобразуем в строки с ANSI цветами.
	lines := make([]string, outerHeight)
	for y := 0; y < outerHeight; y++ {
		var sb strings.Builder
		for x := 0; x < outerWidth; x++ {
			ch := grid[y][x]
			// Определяем цвет: если это граница, используем BorderColor, иначе цвет из contentFunc.
			var fg ColorRGB
			var bg ColorRGB
			if y == 0 || y == outerHeight-1 || x == 0 || x == outerWidth-1 {
				fg = f.BorderColor
				bg = f.FillColor
			} else {
				if contentFunc != nil {
					_, fg, bg = contentFunc(x-1, y-1)
				} else {
					fg = ColorCitizen
					bg = f.FillColor
				}
			}
			// Применяем цвет.
			sb.WriteString(SetTrueColorForeground(fg[0], fg[1], fg[2]))
			sb.WriteString(SetTrueColorBackground(bg[0], bg[1], bg[2]))
			sb.WriteRune(ch)
		}
		sb.WriteString(ResetColor())
		lines[y] = sb.String()
	}
	return lines
}

// RenderSimple возвращает рамку с пустым содержимым в виде среза строк.
func (f *Frame) RenderSimple() []string {
	return f.Render(nil)
}
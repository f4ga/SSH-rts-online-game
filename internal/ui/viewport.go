package ui

import (
	// импорты не нужны для viewport
)

// Viewport определяет видимую область мира.
type Viewport struct {
	X, Y          int // верхний левый угол в мировых координатах
	Width, Height int // размеры в клетках
	worldWidth    int
	worldHeight   int
}

// NewViewport создаёт новый вьюпорт.
func NewViewport(worldWidth, worldHeight, width, height int) *Viewport {
	return &Viewport{
		X:           0,
		Y:           0,
		Width:       width,
		Height:      height,
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
	}
}

// CenterOn центрирует вьюпорт на заданных координатах.
func (v *Viewport) CenterOn(x, y int) {
	v.X = x - v.Width/2
	v.Y = y - v.Height/2
	v.Clamp()
}

// Move сдвигает вьюпорт на указанные дельты.
func (v *Viewport) Move(dx, dy int) {
	v.X += dx
	v.Y += dy
	v.Clamp()
}

// Clamp ограничивает вьюпорт в пределах мира.
func (v *Viewport) Clamp() {
	if v.X < 0 {
		v.X = 0
	}
	if v.Y < 0 {
		v.Y = 0
	}
	if v.X > v.worldWidth-v.Width {
		v.X = v.worldWidth - v.Width
	}
	if v.Y > v.worldHeight-v.Height {
		v.Y = v.worldHeight - v.Height
	}
}

// WorldToViewport преобразует мировые координаты в экранные.
// Возвращает (screenX, screenY, ok), где ok = true, если точка внутри вьюпорта.
func (v *Viewport) WorldToViewport(worldX, worldY int) (int, int, bool) {
	sx := worldX - v.X
	sy := worldY - v.Y
	if sx < 0 || sx >= v.Width || sy < 0 || sy >= v.Height {
		return 0, 0, false
	}
	return sx, sy, true
}

// ViewportToWorld преобразует экранные координаты в мировые.
func (v *Viewport) ViewportToWorld(screenX, screenY int) (int, int) {
	return v.X + screenX, v.Y + screenY
}

// SetSize изменяет размер вьюпорта.
func (v *Viewport) SetSize(width, height int) {
	v.Width = width
	v.Height = height
	v.Clamp()
}

// GetBounds возвращает границы вьюпорта в мировых координатах.
func (v *Viewport) GetBounds() (minX, minY, maxX, maxY int) {
	return v.X, v.Y, v.X + v.Width - 1, v.Y + v.Height - 1
}

// IsVisible проверяет, видна ли клетка.
func (v *Viewport) IsVisible(x, y int) bool {
	_, _, ok := v.WorldToViewport(x, y)
	return ok
}
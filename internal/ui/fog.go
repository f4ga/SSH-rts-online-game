package ui

import (
	// no internal import needed
)

// FogMap представляет карту тумана войны для одного игрока.
type FogMap struct {
	width, height int
	visible       [][]bool // true если клетка видима
	explored      [][]bool // true если клетка была исследована
}

// NewFogMap создаёт новую карту тумана заданного размера.
func NewFogMap(width, height int) *FogMap {
	visible := make([][]bool, height)
	explored := make([][]bool, height)
	for i := range visible {
		visible[i] = make([]bool, width)
		explored[i] = make([]bool, width)
	}
	return &FogMap{
		width:    width,
		height:   height,
		visible:  visible,
		explored: explored,
	}
}

// UpdateVisibility обновляет видимые клетки на основе источников зрения.
// sources — список мировых координат (x, y) с радиусом зрения.
func (f *FogMap) UpdateVisibility(sources []VisionSource) {
	// Сначала сбрасываем видимость.
	for y := 0; y < f.height; y++ {
		for x := 0; x < f.width; x++ {
			f.visible[y][x] = false
		}
	}

	// Для каждого источника отмечаем клетки в радиусе.
	for _, src := range sources {
		minX := src.X - src.Radius
		if minX < 0 {
			minX = 0
		}
		maxX := src.X + src.Radius
		if maxX >= f.width {
			maxX = f.width - 1
		}
		minY := src.Y - src.Radius
		if minY < 0 {
			minY = 0
		}
		maxY := src.Y + src.Radius
		if maxY >= f.height {
			maxY = f.height - 1
		}

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				// Простой квадратный радиус (можно заменить на круглый).
				if abs(x-src.X)+abs(y-src.Y) <= src.Radius {
					f.visible[y][x] = true
					f.explored[y][x] = true
				}
			}
		}
	}
}

// IsVisible возвращает true, если клетка видима.
func (f *FogMap) IsVisible(x, y int) bool {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return false
	}
	return f.visible[y][x]
}

// IsExplored возвращает true, если клетка была исследована.
func (f *FogMap) IsExplored(x, y int) bool {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return false
	}
	return f.explored[y][x]
}

// GetVisibilitySymbol возвращает символ для отображения клетки с учётом тумана.
// Если клетка видима, возвращает переданный символ.
// Если исследована, но не видима, возвращает затемнённый символ (например, '.').
// Если не исследована, возвращает '?'.
func (f *FogMap) GetVisibilitySymbol(x, y int, normalSymbol rune) rune {
	if f.IsVisible(x, y) {
		return normalSymbol
	}
	if f.IsExplored(x, y) {
		return '.' // затемнённый символ
	}
	return '?'
}

// VisionSource описывает источник зрения.
type VisionSource struct {
	X, Y  int
	Radius int
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
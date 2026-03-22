package ui

import (
	"strings"

	"ssh-arena-app/internal"
	"ssh-arena-app/internal/world"
)

// Minimap представляет миникарту мира.
type Minimap struct {
	width      int   // ширина в символах
	height     int   // высота в символах
	scale      int   // масштаб (сколько тайлов на один символ)
	playerX    int
	playerY    int
	world      internal.World
}

// NewMinimap создаёт новую миникарту.
func NewMinimap(width, height, scale int, world internal.World) *Minimap {
	return &Minimap{
		width:   width,
		height:  height,
		scale:   scale,
		world:   world,
	}
}

// UpdatePosition обновляет позицию игрока на миникарте.
func (m *Minimap) UpdatePosition(x, y int) {
	m.playerX = x
	m.playerY = y
}

// Render возвращает строковое представление миникарты.
func (m *Minimap) Render() string {
	if m.world == nil {
		return strings.Repeat("?", m.width) + "\n"
	}

	// Определяем границы мира, которые нужно отобразить.
	// Центрируем на игроке.
	startX := m.playerX - (m.width*m.scale)/2
	startY := m.playerY - (m.height*m.scale)/2

	var out strings.Builder
	for sy := 0; sy < m.height; sy++ {
		for sx := 0; sx < m.width; sx++ {
			// Координаты в мире для этого блока масштаба.
			wx := startX + sx*m.scale
			wy := startY + sy*m.scale
			symbol := m.getSymbol(wx, wy)
			out.WriteRune(symbol)
		}
		out.WriteString("\n")
	}
	return out.String()
}

// getSymbol возвращает символ для тайла в данных мировых координатах.
func (m *Minimap) getSymbol(x, y int) rune {
	// Если координаты совпадают с позицией игрока.
	if x == m.playerX && y == m.playerY {
		return '☺'
	}

	tile, err := m.world.GetTile(x, y)
	if err != nil || tile == nil {
		return ' '
	}

	// Упрощённое отображение типов тайлов.
	switch tile.Type {
	case world.TileTypeWater:
		return '~'
	case world.TileTypeForest:
		return '♣'
	case world.TileTypeMountain:
		return '⛰'
	case world.TileTypeGrass:
		return '·'
	case world.TileTypeGround:
		return '.'
	case world.TileTypeStone:
		return '#'
	default:
		return '?'
	}
}

// CompactRender возвращает компактную миникарту (одна строка) для статусной панели.
func (m *Minimap) CompactRender() string {
	if m.world == nil {
		return "[no map]"
	}
	// Очень упрощённая версия: просто показываем направление к центру.
	return "[Minimap]"
}
package ui

import (
	"strings"
	"time"
	"ssh-arena-app/internal"
	"ssh-arena-app/internal/world"
)

// ANSIRenderer реализует internal.Renderer с использованием ANSI-кодов.
type ANSIRenderer struct {
	viewport         *Viewport
	statusRenderer   *StatusRenderer
	animationManager *AnimationManager
	theme            *Theme
}

// Theme определяет цветовую схему и символы.
type Theme struct {
	TileSymbols   map[world.TileType]rune
	TileColors    map[world.TileType]Color
	BuildingColor Color
	UnitColor     Color
	CitizenColor  Color
	PlayerColor   Color
}

// DefaultTheme возвращает тему по умолчанию.
func DefaultTheme() *Theme {
	return &Theme{
		TileSymbols: map[world.TileType]rune{
			world.TileTypeGrass:   '·',
			world.TileTypeForest:  '♣',
			world.TileTypeMountain: '⛰',
			world.TileTypeWater:   '~',
			world.TileTypeGround:  '.',
			world.TileTypeStone:   '#',
		},
		TileColors: map[world.TileType]Color{
			world.TileTypeGrass:   ColorGreen,
			world.TileTypeForest:  ColorGreen,
			world.TileTypeMountain: ColorWhite,
			world.TileTypeWater:   ColorBlue,
			world.TileTypeGround:  ColorYellow,
			world.TileTypeStone:   ColorBlack,
		},
		BuildingColor: ColorYellow,
		UnitColor:     ColorRed,
		CitizenColor:  ColorWhite,
		PlayerColor:   ColorCyan,
	}
}

// NewANSIRenderer создаёт новый рендерер.
func NewANSIRenderer(viewport *Viewport, statusWidth int) *ANSIRenderer {
	return &ANSIRenderer{
		viewport:         viewport,
		statusRenderer:   NewStatusRenderer(statusWidth),
		animationManager: NewAnimationManager(),
		theme:            DefaultTheme(),
	}
}

// Render генерирует кадр для игрока.
func (r *ANSIRenderer) Render(playerID string, viewport internal.Viewport) ([]byte, error) {
	// Обновляем анимации.
	r.animationManager.Update(time.Now())

	// Создаём буфер для вывода.
	var out strings.Builder

	// Очистка экрана и перемещение курсора в начало.
	out.WriteString(ClearScreen())

	// Рендерим статусную панель (заглушка).
	// TODO: получить реальные данные игрока.
	status := r.statusRenderer.Render(&internal.Player{Name: "Player"}, nil, 0, "", nil)
	out.WriteString(status)
	out.WriteString("\n")

	// Рендерим мир.
	for sy := 0; sy < r.viewport.Height; sy++ {
		for sx := 0; sx < r.viewport.Width; sx++ {
			wx, wy := r.viewport.ViewportToWorld(sx, sy)
			// TODO: получить тайл из мира.
			tileType := world.TileTypeGrass // заглушка
			symbol := r.theme.TileSymbols[tileType]
			color := r.theme.TileColors[tileType]

			// Проверяем анимации в этой клетке.
			for _, anim := range r.animationManager.GetActive() {
				if anim.X == wx && anim.Y == wy {
					symbol = anim.Symbol
					color = anim.Color
					break
				}
			}

			// TODO: добавить отрисовку зданий, юнитов, жителей.

			// Устанавливаем цвет и выводим символ.
			out.WriteString(SetColor(color, 0))
			out.WriteRune(symbol)
		}
		out.WriteString(ResetColor())
		out.WriteString("\n")
	}

	// Подсказки.
	out.WriteString("\nCommands: /move, /build, /attack, /help\n")

	return []byte(out.String()), nil
}

// SetTheme изменяет тему рендерера.
func (r *ANSIRenderer) SetTheme(theme *Theme) error {
	r.theme = theme
	return nil
}

// AddAnimation добавляет анимацию для отображения.
func (r *ANSIRenderer) AddAnimation(anim *Animation) {
	r.animationManager.Add(anim)
}
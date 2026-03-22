package ui

import (
	"log"
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
	colorScheme      *ColorScheme
	symbolMapping    *SymbolMapping
	world            internal.World
	buildingManager  internal.BuildingManager
	citizenManager   internal.CitizenManager
	combatManager    internal.CombatManager
	economyManager   internal.EconomyManager
	researchManager  internal.ResearchManager
	minimap          *Minimap
	fogMap           *FogMap
	fogEnabled       bool
	useTrueColor     bool
	frame            *Frame
	frameEnabled     bool
}

// NewANSIRenderer создаёт новый рендерер.
func NewANSIRenderer(viewport *Viewport, statusWidth int) *ANSIRenderer {
	return &ANSIRenderer{
		viewport:         viewport,
		statusRenderer:   NewStatusRenderer(statusWidth),
		animationManager: NewAnimationManager(),
		colorScheme:      DefaultColorScheme(),
		symbolMapping:    DefaultSymbolMapping(),
		minimap:          nil, // будет создан после установки world
		fogMap:           nil,
		fogEnabled:       false,
		useTrueColor:     true,
		frame:            nil,
		frameEnabled:     false,
	}
}

// SetManagers устанавливает менеджеры для доступа к игровым данным.
func (r *ANSIRenderer) SetManagers(
	world internal.World,
	building internal.BuildingManager,
	citizen internal.CitizenManager,
	combat internal.CombatManager,
	economy internal.EconomyManager,
	research internal.ResearchManager,
) {
	r.world = world
	r.buildingManager = building
	r.citizenManager = citizen
	r.combatManager = combat
	r.economyManager = economy
	r.researchManager = research

	// Создаём миникарту, если мир доступен.
	if world != nil {
		// Размер миникарты: 20x10 символов, масштаб 8 тайлов на символ.
		r.minimap = NewMinimap(20, 10, 8, world)
	}
}

// SetUseTrueColor включает или выключает true color.
func (r *ANSIRenderer) SetUseTrueColor(use bool) {
	r.useTrueColor = use
}

// EnableFog включает туман войны и создаёт карту тумана.
func (r *ANSIRenderer) EnableFog(width, height int) {
	r.fogEnabled = true
	r.fogMap = NewFogMap(width, height)
}

// DisableFog выключает туман войны.
func (r *ANSIRenderer) DisableFog() {
	r.fogEnabled = false
	r.fogMap = nil
}

// UpdateFog обновляет видимость на основе источников зрения.
func (r *ANSIRenderer) UpdateFog(sources []VisionSource) {
	if r.fogEnabled && r.fogMap != nil {
		r.fogMap.UpdateVisibility(sources)
	}
}

// EnableFrame включает отображение рамки вокруг карты.
func (r *ANSIRenderer) EnableFrame(title string) {
	r.frameEnabled = true
	r.frame = NewFrame(r.viewport.Width, r.viewport.Height)
	r.frame.SetTitle(title)
	// Используем яркие цвета для лучшей видимости: граница белая, заголовок жёлтый, фон тёмный.
	r.frame.SetColors(ColorCitizen, ColorHighlight, ColorDark)
}

// DisableFrame выключает рамку.
func (r *ANSIRenderer) DisableFrame() {
	r.frameEnabled = false
	r.frame = nil
}

// Render генерирует кадр для игрока.
func (r *ANSIRenderer) Render(playerID string, viewport internal.Viewport) ([]byte, error) {
	// Обновляем анимации.
	r.animationManager.Update(time.Now())

	// Создаём буфер для вывода.
	var out strings.Builder

	// Очистка экрана и перемещение курсора в начало.
	out.WriteString(ClearScreen())

	// Рендерим статусную панель с реальными данными.
	player, resources, citizens, research, notifications := r.getPlayerData(playerID)
	status := r.statusRenderer.Render(player, resources, citizens, research, notifications)
	out.WriteString(status)
	out.WriteString("\n")

	// Обновляем позицию игрока на миникарте.
	if r.minimap != nil && player != nil {
		r.minimap.UpdatePosition(player.Location.X, player.Location.Y)
	}

	// Определяем границы вьюпорта.
	// minX, minY, maxX, maxY := r.viewport.GetBounds() // не используется

	// Рендерим карту, возможно с рамкой.
	if r.frameEnabled && r.frame != nil {
		log.Printf("[DEBUG] Frame enabled, width=%d height=%d", r.frame.Width, r.frame.Height)
		// Используем frame.Render с contentFunc, которая рендерит клетку.
		frameLines := r.frame.Render(func(x, y int) (rune, ColorRGB, ColorRGB) {
			wx, wy := r.viewport.ViewportToWorld(x, y)
			symbol, fg, bg := r.renderCellComponents(wx, wy, playerID)
			return symbol, fg, bg
		})
		log.Printf("[DEBUG] Frame lines count: %d", len(frameLines))
		for i, line := range frameLines {
			if i == 0 {
				log.Printf("[DEBUG] First line of frame: %q", line)
			}
			out.WriteString(line)
			out.WriteString("\n")
		}
	} else {
		log.Printf("[DEBUG] Frame disabled or nil")
		// Рендерим мир слой за слоем без рамки.
		for sy := 0; sy < r.viewport.Height; sy++ {
			for sx := 0; sx < r.viewport.Width; sx++ {
				wx, wy := r.viewport.ViewportToWorld(sx, sy)
				cell := r.renderCell(wx, wy, playerID)
				out.WriteString(cell)
			}
			out.WriteString(ResetColor())
			out.WriteString("\n")
		}
	}

	// Рендерим миникарту в правом верхнем углу (поверх карты).
	if r.minimap != nil {
		minimapStr := r.minimap.Render()
		// Позиционируем миникарту: например, справа от статусной панели.
		// Для простоты выведем после основного поля.
		out.WriteString("\n--- Minimap ---\n")
		out.WriteString(minimapStr)
	}

	// Подсказки.
	out.WriteString("\nCommands: WASD to move, /help for commands\n")

	return []byte(out.String()), nil
}

// renderCellComponents возвращает символ и цвет для клетки (без ANSI кодов).
func (r *ANSIRenderer) renderCellComponents(x, y int, playerID string) (symbol rune, fg ColorRGB, bg ColorRGB) {
	// Получаем тайл.
	var tile *world.Tile
	if r.world != nil {
		tile, _ = r.world.GetTile(x, y)
	}
	if tile == nil {
		tile = &world.Tile{Type: world.TileTypeGround}
	}

	// Определяем символ и цвет по умолчанию (тайл).
	symbol = r.symbolMapping.GetTileSymbol(tile.Type)
	fg = r.colorScheme.TileColors[TileType(tile.Type)] // преобразование типа
	bg = ColorDark // фон по умолчанию

	// Проверяем анимации в этой клетке.
	for _, anim := range r.animationManager.GetActive() {
		if anim.X == x && anim.Y == y {
			symbol = anim.Symbol
			fg = anim.Color
			break
		}
	}

	// Отрисовываем сущности в порядке слоёв (снизу вверх):
	// 1. Ресурсы на земле
	if tile.Resource != nil {
		symbol = r.symbolMapping.GetResourceSymbol(tile.Resource.Type)
		fg = r.colorScheme.ResourceColor
	}

	// 2. Здания
	if r.buildingManager != nil {
		buildings, _ := r.buildingManager.GetBuildings(playerID) // упрощённо, нужно все здания в этой клетке
		for _, b := range buildings {
			if b.Position.X == x && b.Position.Y == y {
				symbol = r.symbolMapping.BuildingSymbol
				fg = r.colorScheme.BuildingColor
				break
			}
		}
	}

	// 3. Жители
	if r.citizenManager != nil {
		citizens, _ := r.citizenManager.GetCitizens(playerID)
		for _, c := range citizens {
			if c.Position.X == x && c.Position.Y == y {
				symbol = r.symbolMapping.CitizenSymbol
				fg = r.colorScheme.CitizenColor
				break
			}
		}
	}

	// 4. Юниты
	if r.combatManager != nil {
		units, _ := r.combatManager.GetUnits(playerID)
		for _, u := range units {
			if u.Position.X == x && u.Position.Y == y {
				symbol = r.symbolMapping.GetUnitSymbol(u.Type)
				if u.PlayerID == playerID {
					fg = r.colorScheme.PlayerColor
				} else {
					fg = r.colorScheme.EnemyColor
				}
				break
			}
		}
	}

	// 5. Игрок (маркер)
	player := r.getPlayer(playerID)
	if player != nil && player.Location.X == x && player.Location.Y == y {
		symbol = r.symbolMapping.PlayerSymbol
		fg = r.colorScheme.PlayerColor
	}

	// Применяем туман войны, если включён.
	if r.fogEnabled && r.fogMap != nil {
		symbol = r.fogMap.GetVisibilitySymbol(x, y, symbol)
		// Для невидимых клеток можно изменить цвет на серый.
		if !r.fogMap.IsVisible(x, y) {
			fg = ColorGray
		}
	}

	return symbol, fg, bg
}

// renderCell возвращает отформатированную строку для одной клетки.
func (r *ANSIRenderer) renderCell(x, y int, playerID string) string {
	symbol, fg, bg := r.renderCellComponents(x, y, playerID)
	// Формируем ANSI-последовательность цвета.
	var colorSeq string
	if r.useTrueColor {
		colorSeq = SetTrueColorForeground(fg[0], fg[1], fg[2]) + SetTrueColorBackground(bg[0], bg[1], bg[2])
	} else {
		// Fallback к 8-цветной палитре (упрощённо)
		colorSeq = SetColor(ColorGreen, 0) // заглушка
	}
	return colorSeq + string(symbol)
}

// getPlayerData собирает данные игрока для статусной панели.
func (r *ANSIRenderer) getPlayerData(playerID string) (
	player *internal.Player,
	resources map[string]int,
	citizens int,
	research string,
	notifications []string,
) {
	// Заглушки, нужно реализовать через менеджеры.
	player = r.getPlayer(playerID)
	if player == nil {
		player = &internal.Player{Name: "Player"}
	}
	if r.economyManager != nil {
		resources, _ = r.economyManager.GetBalance(playerID)
	}
	if r.citizenManager != nil {
		citizensList, _ := r.citizenManager.GetCitizens(playerID)
		citizens = len(citizensList)
	}
	if r.researchManager != nil {
		techs, _ := r.researchManager.GetDiscovered(playerID)
		if len(techs) > 0 {
			research = techs[0].Name
		}
	}
	notifications = []string{} // TODO: получить из EventBus
	return
}

// getPlayer возвращает объект игрока (заглушка).
func (r *ANSIRenderer) getPlayer(playerID string) *internal.Player {
	// В реальности нужно получить из GameEngine.
	// Пока возвращаем заглушку с координатами (0,0).
	return &internal.Player{
		ID:   playerID,
		Name: "Player",
		Location: struct{ X, Y int }{
			X: 0,
			Y: 0,
		},
	}
}

// SetTheme изменяет тему рендерера (совместимость с интерфейсом).
func (r *ANSIRenderer) SetTheme(theme *internal.Theme) error {
	// Конвертируем internal.Theme в наши структуры (упрощённо)
	return nil
}

// AddAnimation добавляет анимацию для отображения.
func (r *ANSIRenderer) AddAnimation(anim *Animation) {
	r.animationManager.Add(anim)
}
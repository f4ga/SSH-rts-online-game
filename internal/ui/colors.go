package ui

// ColorRGB представляет цвет в формате RGB (0-255).
type ColorRGB [3]int

// Предопределённые цвета в true color (RGB).
var (
	ColorWater      = ColorRGB{64, 164, 223}   // светло-голубой
	ColorGrass      = ColorRGB{34, 139, 34}    // лесной зелёный
	ColorForest     = ColorRGB{0, 100, 0}      // тёмно-зелёный
	ColorSand       = ColorRGB{238, 203, 173}  // песочный
	ColorMountain   = ColorRGB{139, 137, 137}  // серый
	ColorStone      = ColorRGB{105, 105, 105}  // тёмно-серый
	ColorGround     = ColorRGB{139, 69, 19}    // коричневый
	ColorBuilding   = ColorRGB{255, 215, 0}    // золотой
	ColorCitizen    = ColorRGB{255, 255, 255}  // белый
	ColorUnit       = ColorRGB{255, 0, 0}      // красный
	ColorPlayer     = ColorRGB{0, 255, 0}      // зелёный
	ColorEnemy      = ColorRGB{255, 0, 0}      // красный
	ColorNeutral    = ColorRGB{200, 200, 200}  // серый
	ColorResource   = ColorRGB{255, 165, 0}    // оранжевый
	ColorHighlight  = ColorRGB{255, 255, 0}    // жёлтый
	ColorDark       = ColorRGB{30, 30, 30}     // почти чёрный
	ColorGray       = ColorRGB{128, 128, 128}  // серый
)

// NightTimeDim уменьшает яркость цвета для ночного времени.
func (c ColorRGB) NightTimeDim(factor float64) ColorRGB {
	if factor <= 0 || factor > 1 {
		factor = 0.5
	}
	return ColorRGB{
		int(float64(c[0]) * factor),
		int(float64(c[1]) * factor),
		int(float64(c[2]) * factor),
	}
}

// ToANSI возвращает ANSI escape последовательность для установки цвета текста (true color).
func (c ColorRGB) ToANSI() string {
	return SetTrueColorForeground(c[0], c[1], c[2])
}

// ToANSIBackground возвращает ANSI escape последовательность для установки цвета фона (true color).
func (c ColorRGB) ToANSIBackground() string {
	return SetTrueColorBackground(c[0], c[1], c[2])
}

// ColorScheme определяет цветовую схему для рендеринга.
type ColorScheme struct {
	TileColors      map[TileType]ColorRGB
	BuildingColor   ColorRGB
	CitizenColor    ColorRGB
	UnitColor       ColorRGB
	PlayerColor     ColorRGB
	EnemyColor      ColorRGB
	ResourceColor   ColorRGB
	BackgroundColor ColorRGB
}

// DefaultColorScheme возвращает цветовую схему по умолчанию.
func DefaultColorScheme() *ColorScheme {
	return &ColorScheme{
		TileColors: map[TileType]ColorRGB{
			TileTypeGrass:   ColorGrass,
			TileTypeForest:  ColorForest,
			TileTypeMountain: ColorMountain,
			TileTypeWater:   ColorWater,
			TileTypeGround:  ColorGround,
			TileTypeStone:   ColorStone,
			TileTypeSand:    ColorSand,
		},
		BuildingColor:   ColorBuilding,
		CitizenColor:    ColorCitizen,
		UnitColor:       ColorUnit,
		PlayerColor:     ColorPlayer,
		EnemyColor:      ColorEnemy,
		ResourceColor:   ColorResource,
		BackgroundColor: ColorDark,
	}
}

// TileType - тип тайла (дублируется из world.TileType для удобства).
type TileType int

const (
	TileTypeGrass TileType = iota
	TileTypeForest
	TileTypeMountain
	TileTypeWater
	TileTypeGround
	TileTypeStone
	TileTypeSand
)
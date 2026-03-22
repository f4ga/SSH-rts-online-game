package ui

import (
	"ssh-arena-app/internal/world"
)

// SymbolMapping определяет символы для отображения сущностей.
type SymbolMapping struct {
	TileSymbols      map[world.TileType]rune
	BuildingSymbol   rune
	CitizenSymbol    rune
	UnitSymbols      map[string]rune // тип юнита -> символ
	ResourceSymbols  map[string]rune // тип ресурса -> символ
	PlayerSymbol     rune
	EnemySymbol      rune
	DefaultSymbol    rune
}

// DefaultSymbolMapping возвращает маппинг символов по умолчанию (Unicode).
func DefaultSymbolMapping() *SymbolMapping {
	return &SymbolMapping{
		TileSymbols: map[world.TileType]rune{
			world.TileTypeGrass:   '░', // светлый оттенок
			world.TileTypeForest:  '♣', // клевер
			world.TileTypeMountain: '▲', // треугольник
			world.TileTypeWater:   '≈', // двойная тильда
			world.TileTypeGround:  '.', // точка
			world.TileTypeStone:   '■', // чёрный квадрат
			world.TileTypeSand:    '▒', // средний оттенок
		},
		BuildingSymbol:  '🏠', // дом
		CitizenSymbol:   '👤', // силуэт человека
		UnitSymbols: map[string]rune{
			"swordsman": '⚔',
			"archer":    '🏹',
			"catapult":  '☄',
			"soldier":   '♞',
			"scout":     '⚐',
		},
		ResourceSymbols: map[string]rune{
			"wood":  '🪵', // полено
			"stone": '🪨', // камень
			"iron":  '⚙',
			"food":  '🍎',
			"gold":  '💰',
		},
		PlayerSymbol:  '@',
		EnemySymbol:   '☠',
		DefaultSymbol: '?',
	}
}

// ASCIIFallbackSymbolMapping возвращает маппинг символов с использованием только ASCII.
func ASCIIFallbackSymbolMapping() *SymbolMapping {
	return &SymbolMapping{
		TileSymbols: map[world.TileType]rune{
			world.TileTypeGrass:   '.',
			world.TileTypeForest:  'T',
			world.TileTypeMountain: '^',
			world.TileTypeWater:   '~',
			world.TileTypeGround:  ',',
			world.TileTypeStone:   '#',
			world.TileTypeSand:    ':',
		},
		BuildingSymbol:  'B',
		CitizenSymbol:   'C',
		UnitSymbols: map[string]rune{
			"swordsman": 'S',
			"archer":    'A',
			"catapult":  'C',
			"soldier":   'U',
			"scout":     's',
		},
		ResourceSymbols: map[string]rune{
			"wood":  'W',
			"stone": 'S',
			"iron":  'I',
			"food":  'F',
			"gold":  'G',
		},
		PlayerSymbol:  '@',
		EnemySymbol:   'E',
		DefaultSymbol: '?',
	}
}

// GetTileSymbol возвращает символ для типа тайла.
func (sm *SymbolMapping) GetTileSymbol(tileType world.TileType) rune {
	if sym, ok := sm.TileSymbols[tileType]; ok {
		return sym
	}
	return sm.DefaultSymbol
}

// GetUnitSymbol возвращает символ для типа юнита.
func (sm *SymbolMapping) GetUnitSymbol(unitType string) rune {
	if sym, ok := sm.UnitSymbols[unitType]; ok {
		return sym
	}
	return sm.DefaultSymbol
}

// GetResourceSymbol возвращает символ для типа ресурса.
func (sm *SymbolMapping) GetResourceSymbol(resourceType string) rune {
	if sym, ok := sm.ResourceSymbols[resourceType]; ok {
		return sym
	}
	return sm.DefaultSymbol
}
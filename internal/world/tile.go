// Package world implements the game world: tiles, chunks, and terrain.
package world

// TileType represents the type of a tile (terrain).
type TileType int

const (
	TileTypeGround TileType = iota // земля
	TileTypeForest
	TileTypeMountain
	TileTypeWater
	TileTypeGrass
	TileTypeStone
	TileTypeSand
)

// Biome represents the ecological region of a tile.
type Biome int

const (
	BiomePlains Biome = iota
	BiomeForest
	BiomeDesert
	BiomeMountains
)

// Tile represents a single cell in the world grid.
type Tile struct {
	X, Y  int      // world coordinates
	Type  TileType
	Biome Biome
	Height int
	// Additional fields for game logic (optional)
	Humidity   float64
	Resource   *Resource
	BuildingID string // empty if no building
	OwnerID    string // player ID who owns this tile (if any)
}

// Resource describes a harvestable resource on a tile.
type Resource struct {
	Type   string
	Amount int
	Max    int
}

// IsPassable returns true if units can move through this tile.
func (t *Tile) IsPassable() bool {
	switch t.Type {
	case TileTypeWater, TileTypeMountain:
		return false
	default:
		return true
	}
}

// IsBuildable returns true if a building can be placed here.
func (t *Tile) IsBuildable() bool {
	return t.IsPassable() && t.BuildingID == ""
}
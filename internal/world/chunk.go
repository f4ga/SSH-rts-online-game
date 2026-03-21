package world

import (
	"sync"
)

// ChunkSize defines the dimensions of a chunk (square).
const ChunkSize = 16

// Chunk represents a fixed‑size region of tiles for efficient loading and saving.
type Chunk struct {
	X, Y  int // chunk coordinates, not world coordinates
	Tiles [ChunkSize][ChunkSize]Tile
	mu    sync.RWMutex
	dirty bool // whether the chunk has unsaved changes
}

// NewChunk creates an empty chunk at the given chunk coordinates.
func NewChunk(chunkX, chunkY int) *Chunk {
	c := &Chunk{
		X: chunkX,
		Y: chunkY,
	}
	// Initialize tiles with default ground/plains
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			c.Tiles[x][y] = Tile{
				X:      chunkX*ChunkSize + x,
				Y:      chunkY*ChunkSize + y,
				Type:   TileTypeGround,
				Biome:  BiomePlains,
				Height: 0,
			}
		}
	}
	return c
}

// GetTile returns the tile at local coordinates (x,y) within the chunk.
// Coordinates are relative to the chunk origin (0‑ChunkSize‑1).
func (c *Chunk) GetTile(x, y int) *Tile {
	if x < 0 || x >= ChunkSize || y < 0 || y >= ChunkSize {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Return a pointer to the tile inside the array.
	return &c.Tiles[x][y]
}

// SetTile updates a tile within the chunk.
func (c *Chunk) SetTile(x, y int, tile Tile) {
	if x < 0 || x >= ChunkSize || y < 0 || y >= ChunkSize {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Tiles[x][y] = tile
	c.dirty = true
}

// IsDirty returns true if the chunk has unsaved modifications.
func (c *Chunk) IsDirty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dirty
}

// MarkClean marks the chunk as clean (saved).
func (c *Chunk) MarkClean() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirty = false
}
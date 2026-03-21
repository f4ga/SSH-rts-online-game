package world

import (
	"math"
	"sync"
	"time"

	"ssh-arena-app/pkg/errors"
)

// World implements the internal.World interface.
type World struct {
	chunks map[[2]int]*Chunk
	seed   int64
	mu     sync.RWMutex
	time   *WorldTime
}

// WorldTime represents the in‑game calendar and clock.
type WorldTime struct {
	Day   int     // days since epoch
	Hour  int     // hour of the day (0‑23)
	Speed float64 // game seconds per real second
	mu    sync.RWMutex
}

// NewWorld creates an empty world with the given seed.
func NewWorld(seed int64) *World {
	return &World{
		chunks: make(map[[2]int]*Chunk),
		seed:   seed,
		time: &WorldTime{
			Day:   1,
			Hour:  8,
			Speed: 60.0, // one game minute per real second
		},
	}
}

// chunkKey returns the map key for chunk coordinates.
func chunkKey(cx, cy int) [2]int {
	return [2]int{cx, cy}
}

// getOrCreateChunk loads a chunk from memory or creates a new one.
func (w *World) getOrCreateChunk(cx, cy int) *Chunk {
	key := chunkKey(cx, cy)
	w.mu.RLock()
	chunk, exists := w.chunks[key]
	w.mu.RUnlock()
	if exists {
		return chunk
	}
	// Create and generate the chunk
	w.mu.Lock()
	defer w.mu.Unlock()
	// Double‑check after acquiring write lock
	if chunk, exists = w.chunks[key]; exists {
		return chunk
	}
	chunk = NewChunk(cx, cy)
	w.chunks[key] = chunk
	// Generate terrain for this chunk
	w.generateChunk(chunk, cx, cy)
	return chunk
}

// generateChunk fills a chunk with terrain based on world seed and coordinates.
func (w *World) generateChunk(chunk *Chunk, cx, cy int) {
	// Simple pseudo‑random heightmap using sine/cosine.
	// In a real implementation you would use a proper noise library.
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			wx := float64(cx*ChunkSize + x)
			wy := float64(cy*ChunkSize + y)
			// Simple height function
			h := int(10*math.Sin(wx/20)*math.Cos(wy/20) + 10)
			// Determine tile type based on height
			var tileType TileType
			var biome Biome
			if h < 5 {
				tileType = TileTypeWater
				biome = BiomePlains
			} else if h < 10 {
				tileType = TileTypeGround
				biome = BiomePlains
			} else if h < 15 {
				tileType = TileTypeGrass
				biome = BiomeForest
			} else {
				tileType = TileTypeMountain
				biome = BiomeMountains
			}
			tile := Tile{
				X:      cx*ChunkSize + x,
				Y:      cy*ChunkSize + y,
				Type:   tileType,
				Biome:  biome,
				Height: h,
			}
			chunk.SetTile(x, y, tile)
		}
	}
}

// GetTile returns the tile at the given world coordinates.
func (w *World) GetTile(x, y int) (*Tile, error) {
	cx := x / ChunkSize
	cy := y / ChunkSize
	lx := x % ChunkSize
	ly := y % ChunkSize
	if lx < 0 {
		lx += ChunkSize
		cx--
	}
	if ly < 0 {
		ly += ChunkSize
		cy--
	}
	chunk := w.getOrCreateChunk(cx, cy)
	tile := chunk.GetTile(lx, ly)
	if tile == nil {
		// Should not happen if coordinates are valid
		return nil, errors.NotFound("tile")
	}
	return tile, nil
}

// GetChunk returns the chunk containing the given world coordinates.
func (w *World) GetChunk(x, y int) (*Chunk, error) {
	cx := x / ChunkSize
	cy := y / ChunkSize
	chunk := w.getOrCreateChunk(cx, cy)
	return chunk, nil
}

// SetTile modifies a tile at the given world coordinates.
func (w *World) SetTile(x, y int, tile *Tile) error {
	cx := x / ChunkSize
	cy := y / ChunkSize
	lx := x % ChunkSize
	ly := y % ChunkSize
	if lx < 0 {
		lx += ChunkSize
		cx--
	}
	if ly < 0 {
		ly += ChunkSize
		cy--
	}
	chunk := w.getOrCreateChunk(cx, cy)
	chunk.SetTile(lx, ly, *tile)
	return nil
}

// Generate (re)generates the entire world using the stored seed.
func (w *World) Generate(seed int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seed = seed
	// Clear existing chunks
	w.chunks = make(map[[2]int]*Chunk)
	// For now we generate chunks lazily; could pre‑generate a region here.
	return nil
}

// Save serializes the world to persistent storage (stub).
func (w *World) Save() error {
	// TODO: implement
	return nil
}

// Load restores the world from storage (stub).
func (w *World) Load() error {
	// TODO: implement
	return nil
}

// Update advances the world simulation by delta time.
func (w *World) Update(delta time.Duration) {
	w.time.Update(delta)
	// Other world‑wide updates could go here (weather, resource regeneration, etc.)
}

// Update advances the game time according to delta and speed.
func (wt *WorldTime) Update(delta time.Duration) {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	// Convert delta to game seconds
	gameSeconds := wt.Speed * delta.Seconds()
	// Convert to hours and days
	secondsPerHour := 3600.0
	hours := int(gameSeconds / secondsPerHour)
	wt.Hour += hours
	if wt.Hour >= 24 {
		days := wt.Hour / 24
		wt.Day += days
		wt.Hour %= 24
	}
}

// CurrentTime returns the current day and hour.
func (wt *WorldTime) CurrentTime() (day, hour int) {
	wt.mu.RLock()
	defer wt.mu.RUnlock()
	return wt.Day, wt.Hour
}

package building

import (
	"fmt"
	"sync"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/logger"
)

// manager implements internal.BuildingManager.
type manager struct {
	mu        sync.RWMutex
	buildings map[string]*internal.Building // key: building ID
	log       logger.Logger
}

// NewManager creates a new building manager.
func NewManager() internal.BuildingManager {
	return &manager{
		buildings: make(map[string]*internal.Building),
		log:       logger.Get(),
	}
}

// Construct places a new building at the specified location.
func (m *manager) Construct(playerID string, blueprintID string, x, y int) (*internal.Building, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a unique ID (in production use UUID)
	id := fmt.Sprintf("bld_%d", time.Now().UnixNano())
	building := &internal.Building{
		ID:          id,
		PlayerID:    playerID,
		BlueprintID: blueprintID,
		Level:       1,
		Health:      100,
		Position:    struct{ X, Y int }{X: x, Y: y},
	}
	m.buildings[id] = building
	m.log.Info("building constructed", "id", id, "player", playerID, "blueprint", blueprintID)
	return building, nil
}

// Upgrade increases the level of an existing building.
func (m *manager) Upgrade(buildingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	building, exists := m.buildings[buildingID]
	if !exists {
		return fmt.Errorf("building not found: %s", buildingID)
	}
	building.Level++
	m.log.Debug("building upgraded", "id", buildingID, "new_level", building.Level)
	return nil
}

// Demolish removes a building, possibly refunding resources.
func (m *manager) Demolish(buildingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.buildings[buildingID]; !exists {
		return fmt.Errorf("building not found: %s", buildingID)
	}
	delete(m.buildings, buildingID)
	m.log.Info("building demolished", "id", buildingID)
	return nil
}

// GetBuildings returns all buildings owned by a player.
func (m *manager) GetBuildings(playerID string) ([]*internal.Building, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*internal.Building
	for _, b := range m.buildings {
		if b.PlayerID == playerID {
			result = append(result, b)
		}
	}
	return result, nil
}
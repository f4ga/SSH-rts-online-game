package citizen

import (
	"fmt"
	"sync"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/logger"
)

// manager implements internal.CitizenManager.
type manager struct {
	mu       sync.RWMutex
	citizens map[string]*internal.Citizen // key: citizen ID
	log      logger.Logger
}

// NewManager creates a new citizen manager.
func NewManager() internal.CitizenManager {
	return &manager{
		citizens: make(map[string]*internal.Citizen),
		log:      logger.Get(),
	}
}

// Spawn creates a new citizen for a player.
func (m *manager) Spawn(playerID string, locationX, locationY int) (*internal.Citizen, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("cit_%d", time.Now().UnixNano())
	citizen := &internal.Citizen{
		ID:       id,
		PlayerID: playerID,
		Name:     fmt.Sprintf("Citizen %s", id),
		Age:      0,
		Health:   100,
		Hunger:   0,
		JobID:    "",
		Position: struct{ X, Y int }{X: locationX, Y: locationY},
	}
	m.citizens[id] = citizen
	m.log.Info("citizen spawned", "id", id, "player", playerID)
	return citizen, nil
}

// AssignJob assigns a citizen to a job (building, resource, etc.).
func (m *manager) AssignJob(citizenID, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	citizen, exists := m.citizens[citizenID]
	if !exists {
		return fmt.Errorf("citizen not found: %s", citizenID)
	}
	citizen.JobID = jobID
	m.log.Debug("citizen job assigned", "citizen", citizenID, "job", jobID)
	return nil
}

// UpdateNeeds updates the needs of all citizens (called each tick).
func (m *manager) UpdateNeeds(delta time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simplified: increase hunger over time
	for _, c := range m.citizens {
		c.Hunger += int(delta.Seconds()) * 1 // arbitrary
		if c.Hunger > 100 {
			c.Hunger = 100
			c.Health -= 1 // health deteriorates if starving
			if c.Health < 0 {
				c.Health = 0
			}
		}
	}
}

// GetCitizens returns all citizens belonging to a player.
func (m *manager) GetCitizens(playerID string) ([]*internal.Citizen, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*internal.Citizen
	for _, c := range m.citizens {
		if c.PlayerID == playerID {
			result = append(result, c)
		}
	}
	return result, nil
}
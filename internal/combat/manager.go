package combat

import (
	"sync"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/errors"
	"ssh-arena-app/pkg/logger"
)

// combatManager implements internal.CombatManager.
type combatManager struct {
	mu    sync.RWMutex
	units map[string]*Unit // unitID -> Unit
	world internal.World
	log   logger.Logger
}

// NewCombatManager returns a new CombatManager instance.
func NewCombatManager(w internal.World) internal.CombatManager {
	return &combatManager{
		units: make(map[string]*Unit),
		world: w,
		log:   logger.Get(),
	}
}

// CreateUnit creates a military unit for a player.
func (cm *combatManager) CreateUnit(playerID, unitType string, x, y int) (*internal.Unit, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate unit type
	ut := UnitType(unitType)
	if _, ok := unitStatsMap[ut]; !ok {
		return nil, errors.InvalidInput("unit type")
	}

	// Generate a unique ID (in production use UUID)
	id := generateUnitID(playerID, unitType)
	unit, err := NewUnit(id, playerID, ut, x, y)
	if err != nil {
		return nil, err
	}

	cm.units[unit.ID] = unit
	cm.log.Info("unit created", "unit", unit.ID, "player", playerID, "type", unitType, "position", []int{x, y})

	// Convert to internal.Unit
	return &internal.Unit{
		ID:       unit.ID,
		PlayerID: unit.PlayerID,
		Type:     unitType,
		Health:   unit.Health,
		Attack:   unit.Attack,
		Defense:  unit.Defense,
		Position: unit.Position,
	}, nil
}

// MoveUnit orders a unit to move to a target location.
func (cm *combatManager) MoveUnit(unitID string, targetX, targetY int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	unit, ok := cm.units[unitID]
	if !ok {
		return errors.NotFound("unit")
	}

	// Check if target tile is passable
	tile, err := cm.world.GetTile(targetX, targetY)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInternal, "failed to get tile", err)
	}
	if !tile.IsPassable() {
		return errors.New(errors.ErrCodeConflict, "tile not passable")
	}

	// Move unit
	if err := unit.MoveTo(targetX, targetY, cm.world); err != nil {
		return err
	}

	cm.log.Debug("unit move ordered", "unit", unitID, "target", []int{targetX, targetY})
	return nil
}

// Attack orders a unit to attack another unit or building.
func (cm *combatManager) Attack(attackerID, targetID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	attacker, ok := cm.units[attackerID]
	if !ok {
		return errors.NotFound("attacker unit")
	}

	// Determine if target is a unit or building (simplified: assume unit for now)
	target, ok := cm.units[targetID]
	if !ok {
		// Could be a building; we would need to fetch from building manager.
		// For now, return error.
		return errors.NotFound("target unit")
	}

	// Check range
	if !attacker.CanAttack(target.Position.X, target.Position.Y) {
		return errors.New(errors.ErrCodeConflict, "target out of range")
	}

	// Perform attack
	_, killed := attacker.AttackUnit(target)
	if killed {
		delete(cm.units, target.ID)
		cm.log.Info("unit destroyed", "unit", target.ID, "by", attacker.ID)
	}

	cm.log.Debug("attack executed", "attacker", attackerID, "target", targetID)
	return nil
}

// UpdateCombat processes combat for all units (called each tick).
func (cm *combatManager) UpdateCombat(delta time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, unit := range cm.units {
		unit.Update(delta, cm.world)
		// Additional combat logic (e.g., auto‑attack if target in range) could go here.
	}
}

// GetUnits returns all units belonging to a player.
func (cm *combatManager) GetUnits(playerID string) ([]*internal.Unit, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*internal.Unit
	for _, unit := range cm.units {
		if unit.PlayerID == playerID {
			result = append(result, &internal.Unit{
				ID:       unit.ID,
				PlayerID: unit.PlayerID,
				Type:     string(unit.Type),
				Health:   unit.Health,
				Attack:   unit.Attack,
				Defense:  unit.Defense,
				Position: unit.Position,
			})
		}
	}
	return result, nil
}

// generateUnitID creates a simple unique ID for a unit (for demonstration).
func generateUnitID(playerID, unitType string) string {
	return playerID + "_" + unitType + "_" + time.Now().Format("150405")
}
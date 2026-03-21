package combat

import (
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/internal/world"
	"ssh-arena-app/pkg/errors"
	"ssh-arena-app/pkg/logger"
)

// UnitType represents the type of a military unit.
type UnitType string

const (
	UnitTypeSwordsman UnitType = "swordsman"
	UnitTypeArcher    UnitType = "archer"
	UnitTypeCatapult  UnitType = "catapult"
)

// UnitStats defines the base attributes of a unit type.
type UnitStats struct {
	Health   int
	Attack   int
	Defense  int
	Range    int // attack range in tiles
	Speed    float64 // tiles per second
	Siege    int // bonus damage against buildings
}

// unitStatsMap holds the stats for each unit type.
var unitStatsMap = map[UnitType]UnitStats{
	UnitTypeSwordsman: {Health: 100, Attack: 15, Defense: 10, Range: 1, Speed: 2.0, Siege: 5},
	UnitTypeArcher:    {Health: 70, Attack: 12, Defense: 5, Range: 4, Speed: 1.5, Siege: 2},
	UnitTypeCatapult:  {Health: 150, Attack: 25, Defense: 20, Range: 6, Speed: 0.8, Siege: 30},
}

// Unit represents a military unit on the battlefield.
type Unit struct {
	ID        string
	PlayerID  string
	Type      UnitType
	Health    int
	MaxHealth int
	Attack    int
	Defense   int
	Range     int
	Speed     float64
	Siege     int

	Position struct{ X, Y int }
	TargetID string // ID of unit or building being attacked, empty if none
	Path     []world.Tile // path for movement
	LastMove time.Time

	log logger.Logger
}

// NewUnit creates a new unit with the given type and position.
func NewUnit(id, playerID string, unitType UnitType, x, y int) (*Unit, error) {
	stats, ok := unitStatsMap[unitType]
	if !ok {
		return nil, errors.InvalidInput("unit type")
	}
	return &Unit{
		ID:        id,
		PlayerID:  playerID,
		Type:      unitType,
		Health:    stats.Health,
		MaxHealth: stats.Health,
		Attack:    stats.Attack,
		Defense:   stats.Defense,
		Range:     stats.Range,
		Speed:     stats.Speed,
		Siege:     stats.Siege,
		Position:  struct{ X, Y int }{x, y},
		LastMove:  time.Now(),
		log:       logger.Get(),
	}, nil
}

// MoveTo sets a path to the target coordinates (simplified).
func (u *Unit) MoveTo(targetX, targetY int, w internal.World) error {
	// For simplicity, we just set the target position directly.
	// In a real implementation, you would compute a path using A*.
	u.Position.X = targetX
	u.Position.Y = targetY
	u.LastMove = time.Now()
	u.log.Debug("unit moved", "unit", u.ID, "x", targetX, "y", targetY)
	return nil
}

// CanAttack returns true if the target is within attack range.
func (u *Unit) CanAttack(targetX, targetY int) bool {
	dx := u.Position.X - targetX
	if dx < 0 {
		dx = -dx
	}
	dy := u.Position.Y - targetY
	if dy < 0 {
		dy = -dy
	}
	distance := dx + dy // Manhattan distance (simplified)
	return distance <= u.Range
}

// AttackUnit calculates damage against another unit using combat logic.
func (u *Unit) AttackUnit(target *Unit) (damage int, killed bool) {
	damage, killed = PerformUnitAttack(u, target)
	u.log.Debug("unit attacked", "attacker", u.ID, "target", target.ID, "damage", damage, "target_health", target.Health)
	return damage, killed
}

// AttackBuilding calculates damage against a building using combat logic.
func (u *Unit) AttackBuilding(building *internal.Building) (damage int, destroyed bool) {
	damage, destroyed = PerformBuildingAttack(u, building)
	u.log.Debug("unit attacked building", "unit", u.ID, "building", building.ID, "damage", damage, "building_health", building.Health)
	return damage, destroyed
}

// Update performs periodic updates (e.g., movement along path).
func (u *Unit) Update(delta time.Duration, w internal.World) {
	// If unit has a path, move along it.
	// This is a placeholder; actual implementation would interpolate.
}
package research

import (
	"ssh-arena-app/internal/economy"
	"time"
)

// TechEffectType represents the type of effect a technology has.
type TechEffectType string

const (
	EffectProductionBoost TechEffectType = "production_boost"
	EffectUnlockUnit      TechEffectType = "unlock_unit"
	EffectUnlockBuilding  TechEffectType = "unlock_building"
	EffectStatBonus       TechEffectType = "stat_bonus"
	EffectDiscount        TechEffectType = "discount"
)

// TechEffect describes a concrete effect of researching a technology.
type TechEffect struct {
	Type   TechEffectType
	Target string            // e.g., unit type, building type, resource type
	Value  float64           // multiplier or absolute bonus
	Data   map[string]interface{} // additional parameters
}

// TechNode represents a single technology in the tree.
type TechNode struct {
	ID           string
	Name         string
	Description  string
	Requirements []string // IDs of prerequisite technologies
	Cost         economy.ResourceSet
	Duration     time.Duration // base research time
	Effects      []TechEffect
	Icon         string // UI icon
}

// Technology represents a player's progress on a specific tech.
type Technology struct {
	Node      *TechNode
	PlayerID  string
	Progress  time.Duration // time already invested
	Completed bool
	StartedAt time.Time
}

// IsResearched returns true if the technology has been completed.
func (t *Technology) IsResearched() bool {
	return t.Completed
}

// RemainingTime returns the remaining research duration.
func (t *Technology) RemainingTime() time.Duration {
	if t.Completed {
		return 0
	}
	remaining := t.Node.Duration - t.Progress
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Advance adds research time to the technology.
func (t *Technology) Advance(delta time.Duration) bool {
	if t.Completed {
		return false
	}
	t.Progress += delta
	if t.Progress >= t.Node.Duration {
		t.Completed = true
		return true
	}
	return false
}

// DefaultTechTree returns a sample technology tree.
func DefaultTechTree() []*TechNode {
	return []*TechNode{
		{
			ID:           "agriculture",
			Name:         "Agriculture",
			Description:  "Increases food production by 20%",
			Requirements: []string{},
			Cost:         economy.ResourceSet{economy.ResourceFood: 50},
			Duration:     30 * time.Second,
			Effects: []TechEffect{
				{Type: EffectProductionBoost, Target: "food", Value: 1.2},
			},
		},
		{
			ID:           "metalworking",
			Name:         "Metalworking",
			Description:  "Unlocks swordsman unit and improves metal tools",
			Requirements: []string{"agriculture"},
			Cost:         economy.ResourceSet{economy.ResourceIron: 100, economy.ResourceWood: 50},
			Duration:     60 * time.Second,
			Effects: []TechEffect{
				{Type: EffectUnlockUnit, Target: "swordsman", Value: 1},
				{Type: EffectProductionBoost, Target: "iron", Value: 1.1},
			},
		},
		{
			ID:           "archery",
			Name:         "Archery",
			Description:  "Unlocks archer unit",
			Requirements: []string{"agriculture"},
			Cost:         economy.ResourceSet{economy.ResourceWood: 80, economy.ResourceStone: 30},
			Duration:     45 * time.Second,
			Effects: []TechEffect{
				{Type: EffectUnlockUnit, Target: "archer", Value: 1},
			},
		},
		{
			ID:           "engineering",
			Name:         "Engineering",
			Description:  "Unlocks catapult and improves building construction speed",
			Requirements: []string{"metalworking", "archery"},
			Cost:         economy.ResourceSet{economy.ResourceStone: 150, economy.ResourceWood: 100, economy.ResourceIron: 80},
			Duration:     120 * time.Second,
			Effects: []TechEffect{
				{Type: EffectUnlockUnit, Target: "catapult", Value: 1},
				{Type: EffectDiscount, Target: "building_cost", Value: 0.9},
			},
		},
	}
}
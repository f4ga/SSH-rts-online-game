package events

import (
	"time"

	"ssh-arena-app/internal"
)

// Константы типов событий.
const (
	EventTypeBuildingBuilt     = "building_built"
	EventTypeBuildingDestroyed = "building_destroyed"
	EventTypeUnitDied          = "unit_died"
	EventTypeCitizenCreated    = "citizen_created"
	EventTypeTechResearched    = "tech_researched"
	EventTypeResourceChanged   = "resource_changed"
	EventTypeCombatStarted     = "combat_started"
	EventTypeQuestCompleted    = "quest_completed"
)

// BuildingBuiltEvent данные для события постройки здания.
type BuildingBuiltEvent struct {
	PlayerID    string
	BuildingID  string
	BuildingType string
	X, Y        int
}

// BuildingDestroyedEvent данные для события разрушения здания.
type BuildingDestroyedEvent struct {
	BuildingID string
}

// UnitDiedEvent данные для события смерти юнита.
type UnitDiedEvent struct {
	UnitID   string
	UnitType string
	KillerID string
}

// CitizenCreatedEvent данные для события создания жителя.
type CitizenCreatedEvent struct {
	CitizenID  string
	PlayerID   string
	BuildingID string
}

// TechResearchedEvent данные для события исследования технологии.
type TechResearchedEvent struct {
	PlayerID string
	TechID   string
}

// ResourceChangedEvent данные для события изменения ресурсов.
type ResourceChangedEvent struct {
	PlayerID     string
	ResourceType string
	Amount       int
	Delta        int
}

// CombatStartedEvent данные для события начала боя.
type CombatStartedEvent struct {
	AttackerID string
	DefenderID string
	LocationX  int
	LocationY  int
}

// QuestCompletedEvent данные для события завершения квеста.
type QuestCompletedEvent struct {
	PlayerID string
	QuestID  string
}

// NewBuildingBuilt создаёт событие постройки здания.
func NewBuildingBuilt(playerID, buildingID, buildingType string, x, y int) internal.Event {
	return internal.Event{
		Type:      EventTypeBuildingBuilt,
		Timestamp: time.Now(),
		Payload: BuildingBuiltEvent{
			PlayerID:    playerID,
			BuildingID:  buildingID,
			BuildingType: buildingType,
			X:           x,
			Y:           y,
		},
	}
}

// NewBuildingDestroyed создаёт событие разрушения здания.
func NewBuildingDestroyed(buildingID string) internal.Event {
	return internal.Event{
		Type:      EventTypeBuildingDestroyed,
		Timestamp: time.Now(),
		Payload: BuildingDestroyedEvent{
			BuildingID: buildingID,
		},
	}
}

// NewUnitDied создаёт событие смерти юнита.
func NewUnitDied(unitID, unitType, killerID string) internal.Event {
	return internal.Event{
		Type:      EventTypeUnitDied,
		Timestamp: time.Now(),
		Payload: UnitDiedEvent{
			UnitID:   unitID,
			UnitType: unitType,
			KillerID: killerID,
		},
	}
}

// NewCitizenCreated создаёт событие создания жителя.
func NewCitizenCreated(citizenID, playerID, buildingID string) internal.Event {
	return internal.Event{
		Type:      EventTypeCitizenCreated,
		Timestamp: time.Now(),
		Payload: CitizenCreatedEvent{
			CitizenID:  citizenID,
			PlayerID:   playerID,
			BuildingID: buildingID,
		},
	}
}

// NewTechResearched создаёт событие исследования технологии.
func NewTechResearched(playerID, techID string) internal.Event {
	return internal.Event{
		Type:      EventTypeTechResearched,
		Timestamp: time.Now(),
		Payload: TechResearchedEvent{
			PlayerID: playerID,
			TechID:   techID,
		},
	}
}

// NewResourceChanged создаёт событие изменения ресурсов.
func NewResourceChanged(playerID, resourceType string, amount, delta int) internal.Event {
	return internal.Event{
		Type:      EventTypeResourceChanged,
		Timestamp: time.Now(),
		Payload: ResourceChangedEvent{
			PlayerID:     playerID,
			ResourceType: resourceType,
			Amount:       amount,
			Delta:        delta,
		},
	}
}

// NewCombatStarted создаёт событие начала боя.
func NewCombatStarted(attackerID, defenderID string, x, y int) internal.Event {
	return internal.Event{
		Type:      EventTypeCombatStarted,
		Timestamp: time.Now(),
		Payload: CombatStartedEvent{
			AttackerID: attackerID,
			DefenderID: defenderID,
			LocationX:  x,
			LocationY:  y,
		},
	}
}

// NewQuestCompleted создаёт событие завершения квеста.
func NewQuestCompleted(playerID, questID string) internal.Event {
	return internal.Event{
		Type:      EventTypeQuestCompleted,
		Timestamp: time.Now(),
		Payload: QuestCompletedEvent{
			PlayerID: playerID,
			QuestID:  questID,
		},
	}
}
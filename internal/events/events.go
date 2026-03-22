package events

import (
	"time"

	"ssh-arena-app/internal"
)

// Константы типов событий.
const (
	EventTypeBuildingBuilt      = "building_built"
	EventTypeBuildingDestroyed  = "building_destroyed"
	EventTypeBuildingUpgraded   = "building_upgraded"
	EventTypeUnitSpawned        = "unit_spawned"
	EventTypeUnitDied           = "unit_died"
	EventTypeUnitMoved          = "unit_moved"
	EventTypeCitizenCreated     = "citizen_created"
	EventTypeCitizenDied        = "citizen_died"
	EventTypeCitizenJobChanged  = "citizen_job_changed"
	EventTypeResourceChanged    = "resource_changed"
	EventTypeTechResearched     = "tech_researched"
	EventTypeCombatStarted      = "combat_started"
	EventTypeCombatEnded        = "combat_ended"
	EventTypeTaxRateChanged     = "tax_rate_changed"
	EventTypeTradeCompleted     = "trade_completed"
	EventTypeQuestStarted       = "quest_started"
	EventTypeQuestCompleted     = "quest_completed"
	EventTypeNotification       = "notification"
)

// BuildingBuiltEvent данные для события постройки здания.
type BuildingBuiltEvent struct {
	PlayerID     string
	BuildingID   string
	BuildingType string
	X, Y         int
}

// BuildingDestroyedEvent данные для события разрушения здания.
type BuildingDestroyedEvent struct {
	BuildingID string
}

// BuildingUpgradedEvent данные для события улучшения здания.
type BuildingUpgradedEvent struct {
	BuildingID string
	NewLevel   int
	PlayerID   string
}

// UnitSpawnedEvent данные для события создания юнита.
type UnitSpawnedEvent struct {
	UnitID   string
	PlayerID string
	UnitType string
	X, Y     int
}

// UnitDiedEvent данные для события смерти юнита.
type UnitDiedEvent struct {
	UnitID   string
	UnitType string
	KillerID string
}

// UnitMovedEvent данные для события перемещения юнита.
type UnitMovedEvent struct {
	UnitID     string
	FromX, FromY int
	ToX, ToY   int
}

// CitizenCreatedEvent данные для события создания жителя.
type CitizenCreatedEvent struct {
	CitizenID  string
	PlayerID   string
	BuildingID string
}

// CitizenDiedEvent данные для события смерти жителя.
type CitizenDiedEvent struct {
	CitizenID string
	PlayerID  string
	Cause     string
}

// CitizenJobChangedEvent данные для события смены работы жителя.
type CitizenJobChangedEvent struct {
	CitizenID string
	OldJobID  string
	NewJobID  string
}

// ResourceChangedEvent данные для события изменения ресурсов.
type ResourceChangedEvent struct {
	PlayerID     string
	ResourceType string
	Amount       int
	Delta        int
}

// TechResearchedEvent данные для события исследования технологии.
type TechResearchedEvent struct {
	PlayerID string
	TechID   string
}

// CombatStartedEvent данные для события начала боя.
type CombatStartedEvent struct {
	AttackerID string
	DefenderID string
	LocationX  int
	LocationY  int
}

// CombatEndedEvent данные для события окончания боя.
type CombatEndedEvent struct {
	AttackerID string
	DefenderID string
	WinnerID   string
	LocationX  int
	LocationY  int
}

// TaxRateChangedEvent данные для события изменения налоговой ставки.
type TaxRateChangedEvent struct {
	PlayerID  string
	OldRate   float64
	NewRate   float64
}

// TradeCompletedEvent данные для события завершения торговой сделки.
type TradeCompletedEvent struct {
	FromPlayerID string
	ToPlayerID   string
	Resources    map[string]int
	GoldAmount   int
}

// QuestStartedEvent данные для события начала квеста.
type QuestStartedEvent struct {
	PlayerID string
	QuestID  string
}

// QuestCompletedEvent данные для события завершения квеста.
type QuestCompletedEvent struct {
	PlayerID string
	QuestID  string
}

// NotificationEvent данные для уведомления игрока.
type NotificationEvent struct {
	PlayerID string
	Message  string
	Severity string // "info", "warning", "error"
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

// NewBuildingUpgraded создаёт событие улучшения здания.
func NewBuildingUpgraded(buildingID string, newLevel int, playerID string) internal.Event {
	return internal.Event{
		Type:      EventTypeBuildingUpgraded,
		Timestamp: time.Now(),
		Payload: BuildingUpgradedEvent{
			BuildingID: buildingID,
			NewLevel:   newLevel,
			PlayerID:   playerID,
		},
	}
}

// NewUnitSpawned создаёт событие создания юнита.
func NewUnitSpawned(unitID, playerID, unitType string, x, y int) internal.Event {
	return internal.Event{
		Type:      EventTypeUnitSpawned,
		Timestamp: time.Now(),
		Payload: UnitSpawnedEvent{
			UnitID:   unitID,
			PlayerID: playerID,
			UnitType: unitType,
			X:        x,
			Y:        y,
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

// NewUnitMoved создаёт событие перемещения юнита.
func NewUnitMoved(unitID string, fromX, fromY, toX, toY int) internal.Event {
	return internal.Event{
		Type:      EventTypeUnitMoved,
		Timestamp: time.Now(),
		Payload: UnitMovedEvent{
			UnitID: unitID,
			FromX:  fromX,
			FromY:  fromY,
			ToX:    toX,
			ToY:    toY,
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

// NewCitizenDied создаёт событие смерти жителя.
func NewCitizenDied(citizenID, playerID, cause string) internal.Event {
	return internal.Event{
		Type:      EventTypeCitizenDied,
		Timestamp: time.Now(),
		Payload: CitizenDiedEvent{
			CitizenID: citizenID,
			PlayerID:  playerID,
			Cause:     cause,
		},
	}
}

// NewCitizenJobChanged создаёт событие смены работы жителя.
func NewCitizenJobChanged(citizenID, oldJobID, newJobID string) internal.Event {
	return internal.Event{
		Type:      EventTypeCitizenJobChanged,
		Timestamp: time.Now(),
		Payload: CitizenJobChangedEvent{
			CitizenID: citizenID,
			OldJobID:  oldJobID,
			NewJobID:  newJobID,
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

// NewCombatEnded создаёт событие окончания боя.
func NewCombatEnded(attackerID, defenderID, winnerID string, x, y int) internal.Event {
	return internal.Event{
		Type:      EventTypeCombatEnded,
		Timestamp: time.Now(),
		Payload: CombatEndedEvent{
			AttackerID: attackerID,
			DefenderID: defenderID,
			WinnerID:   winnerID,
			LocationX:  x,
			LocationY:  y,
		},
	}
}

// NewTaxRateChanged создаёт событие изменения налоговой ставки.
func NewTaxRateChanged(playerID string, oldRate, newRate float64) internal.Event {
	return internal.Event{
		Type:      EventTypeTaxRateChanged,
		Timestamp: time.Now(),
		Payload: TaxRateChangedEvent{
			PlayerID: playerID,
			OldRate:  oldRate,
			NewRate:  newRate,
		},
	}
}

// NewTradeCompleted создаёт событие завершения торговой сделки.
func NewTradeCompleted(fromPlayerID, toPlayerID string, resources map[string]int, goldAmount int) internal.Event {
	return internal.Event{
		Type:      EventTypeTradeCompleted,
		Timestamp: time.Now(),
		Payload: TradeCompletedEvent{
			FromPlayerID: fromPlayerID,
			ToPlayerID:   toPlayerID,
			Resources:    resources,
			GoldAmount:   goldAmount,
		},
	}
}

// NewQuestStarted создаёт событие начала квеста.
func NewQuestStarted(playerID, questID string) internal.Event {
	return internal.Event{
		Type:      EventTypeQuestStarted,
		Timestamp: time.Now(),
		Payload: QuestStartedEvent{
			PlayerID: playerID,
			QuestID:  questID,
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

// NewNotification создаёт событие уведомления.
func NewNotification(playerID, message, severity string) internal.Event {
	return internal.Event{
		Type:      EventTypeNotification,
		Timestamp: time.Now(),
		Payload: NotificationEvent{
			PlayerID: playerID,
			Message:  message,
			Severity: severity,
		},
	}
}
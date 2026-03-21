package events

import (
	"sync"

	"ssh-arena-app/internal"
)

// Quest представляет квест в игре.
type Quest interface {
	ID() string
	Title() string
	Description() string
	IsCompleted(playerID string) bool
	GetReward() map[string]int // resource type -> amount
}

// QuestManager управляет квестами игроков.
type QuestManager struct {
	mu      sync.RWMutex
	quests  map[string]Quest                    // все квесты по ID
	active  map[string]map[string]bool          // playerID -> set of quest IDs
	bus     internal.EventBus
}

// NewQuestManager создаёт новый менеджер квестов.
func NewQuestManager(bus internal.EventBus) *QuestManager {
	return &QuestManager{
		quests: make(map[string]Quest),
		active: make(map[string]map[string]bool),
		bus:    bus,
	}
}

// RegisterQuest регистрирует квест в менеджере.
func (qm *QuestManager) RegisterQuest(q Quest) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.quests[q.ID()] = q
}

// ActivateQuest активирует квест для игрока.
func (qm *QuestManager) ActivateQuest(playerID, questID string) bool {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	if _, ok := qm.quests[questID]; !ok {
		return false
	}
	if qm.active[playerID] == nil {
		qm.active[playerID] = make(map[string]bool)
	}
	qm.active[playerID][questID] = true
	return true
}

// CheckProgress проверяет прогресс всех активных квестов игрока и завершает выполненные.
func (qm *QuestManager) CheckProgress(playerID string) {
	qm.mu.RLock()
	activeQuests := make([]string, 0, len(qm.active[playerID]))
	for qid := range qm.active[playerID] {
		activeQuests = append(activeQuests, qid)
	}
	qm.mu.RUnlock()

	for _, qid := range activeQuests {
		qm.mu.RLock()
		quest, ok := qm.quests[qid]
		qm.mu.RUnlock()
		if !ok {
			continue
		}
		if quest.IsCompleted(playerID) {
			qm.completeQuest(playerID, quest)
		}
	}
}

// completeQuest завершает квест, выдаёт награду и публикует событие.
func (qm *QuestManager) completeQuest(playerID string, quest Quest) {
	qm.mu.Lock()
	delete(qm.active[playerID], quest.ID())
	qm.mu.Unlock()

	// Выдать награду (здесь должна быть интеграция с EconomyManager)
	// Пока просто логируем.
	_ = quest.GetReward() // награда пока не используется
	// TODO: добавить ресурсы игроку через EconomyManager.

	// Опубликовать событие завершения квеста.
	qm.bus.Publish(NewQuestCompleted(playerID, quest.ID()))
}

// GetActiveQuests возвращает список активных квестов игрока.
func (qm *QuestManager) GetActiveQuests(playerID string) []Quest {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	var result []Quest
	for qid := range qm.active[playerID] {
		if quest, ok := qm.quests[qid]; ok {
			result = append(result, quest)
		}
	}
	return result
}

// Пример реализации простого квеста.
type simpleQuest struct {
	id          string
	title       string
	description string
	condition   func(playerID string) bool
	reward      map[string]int
}

func (q *simpleQuest) ID() string                { return q.id }
func (q *simpleQuest) Title() string             { return q.title }
func (q *simpleQuest) Description() string       { return q.description }
func (q *simpleQuest) IsCompleted(playerID string) bool {
	if q.condition == nil {
		return false
	}
	return q.condition(playerID)
}
func (q *simpleQuest) GetReward() map[string]int { return q.reward }

// NewSimpleQuest создаёт простой квест.
func NewSimpleQuest(id, title, desc string, condition func(playerID string) bool, reward map[string]int) Quest {
	return &simpleQuest{
		id:          id,
		title:       title,
		description: desc,
		condition:   condition,
		reward:      reward,
	}
}
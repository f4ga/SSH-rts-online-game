package economy

import (
	"sync"

	"ssh-arena-app/pkg/errors"
	"ssh-arena-app/pkg/logger"
)

// Treasury manages a player's gold and resources.
type Treasury struct {
	playerID string
	gold     int
	resources ResourceSet
	mu       sync.RWMutex
	log      logger.Logger
}

// NewTreasury creates a new treasury for a player.
func NewTreasury(playerID string, initialGold int, initialResources ResourceSet) *Treasury {
	if initialResources == nil {
		initialResources = make(ResourceSet)
	}
	return &Treasury{
		playerID:  playerID,
		gold:      initialGold,
		resources: initialResources.Clone(),
		log:       logger.Get(),
	}
}

// AddGold adds gold to the treasury.
func (t *Treasury) AddGold(amount int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.gold += amount
	t.log.Debug("gold added", "player", t.playerID, "amount", amount, "total", t.gold)
}

// SpendGold attempts to spend gold, returns error if insufficient.
func (t *Treasury) SpendGold(amount int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.gold < amount {
		return errors.New(errors.ErrCodeConflict, "insufficient gold")
	}
	t.gold -= amount
	t.log.Debug("gold spent", "player", t.playerID, "amount", amount, "remaining", t.gold)
	return nil
}

// GetGold returns the current gold balance.
func (t *Treasury) GetGold() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.gold
}

// AddResource adds a resource amount.
func (t *Treasury) AddResource(resource ResourceType, amount int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resources[resource] += amount
	t.log.Debug("resource added", "player", t.playerID, "resource", resource, "amount", amount, "total", t.resources[resource])
}

// SpendResource attempts to spend a resource, returns error if insufficient.
func (t *Treasury) SpendResource(resource ResourceType, amount int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.resources[resource] < amount {
		return errors.New(errors.ErrCodeConflict, "insufficient resource")
	}
	t.resources[resource] -= amount
	t.log.Debug("resource spent", "player", t.playerID, "resource", resource, "amount", amount, "remaining", t.resources[resource])
	return nil
}

// GetResources returns a copy of the resource set.
func (t *Treasury) GetResources() ResourceSet {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.resources.Clone()
}

// TransferTo transfers resources and/or gold to another treasury.
func (t *Treasury) TransferTo(target *Treasury, gold int, resources ResourceSet) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	target.mu.Lock()
	defer target.mu.Unlock()

	// Check gold
	if t.gold < gold {
		return errors.New(errors.ErrCodeConflict, "insufficient gold for transfer")
	}
	// Check resources
	for typ, amount := range resources {
		if t.resources[typ] < amount {
			return errors.New(errors.ErrCodeConflict, "insufficient resource for transfer")
		}
	}

	// Perform transfer
	t.gold -= gold
	target.gold += gold
	for typ, amount := range resources {
		t.resources[typ] -= amount
		target.resources[typ] += amount
	}

	t.log.Debug("transfer completed", "from", t.playerID, "to", target.playerID, "gold", gold, "resources", resources)
	return nil
}
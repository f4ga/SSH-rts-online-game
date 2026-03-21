package economy

import (
	"sync"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/errors"
	"ssh-arena-app/pkg/logger"
)

// economyManager implements internal.EconomyManager.
type economyManager struct {
	mu          sync.RWMutex
	treasuries  map[string]*Treasury // playerID -> Treasury
	market      *Market
	tax         *TaxManager
	trade       *TradeManager
	log         logger.Logger
}

// NewEconomyManager creates a new economy manager.
func NewEconomyManager() internal.EconomyManager {
	return &economyManager{
		treasuries: make(map[string]*Treasury),
		market:     NewMarket(),
		tax:        NewTaxManager(),
		trade:      NewTradeManager(),
		log:        logger.Get(),
	}
}

// Produce generates resources based on buildings and citizens (simplified).
func (em *economyManager) Produce(delta time.Duration) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// In a real implementation, you would iterate over players' production buildings
	// and add resources accordingly. For now, we just add a small amount of gold.
	for playerID, treasury := range em.treasuries {
		// Example: each player gets 10 gold per second
		goldPerSecond := 10.0
		goldEarned := int(goldPerSecond * delta.Seconds())
		if goldEarned > 0 {
			treasury.AddGold(goldEarned)
			em.log.Debug("production", "player", playerID, "gold", goldEarned)
		}
	}
}

// Transfer transfers resources from one player to another (trade).
func (em *economyManager) Transfer(fromPlayerID, toPlayerID string, resources map[string]int) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	fromTreasury, ok := em.treasuries[fromPlayerID]
	if !ok {
		return errors.NotFound("player treasury")
	}
	toTreasury, ok := em.treasuries[toPlayerID]
	if !ok {
		return errors.NotFound("player treasury")
	}

	// Convert map[string]int to ResourceSet
	rs := make(ResourceSet)
	for typStr, amount := range resources {
		rs[ResourceType(typStr)] = amount
	}

	// Use treasury's TransferTo method
	if err := fromTreasury.TransferTo(toTreasury, 0, rs); err != nil {
		return err
	}

	em.log.Info("resources transferred", "from", fromPlayerID, "to", toPlayerID, "resources", resources)
	return nil
}

// CollectTaxes collects taxes from citizens and buildings.
func (em *economyManager) CollectTaxes(playerID string) (int, error) {
	em.mu.RLock()
	treasury, ok := em.treasuries[playerID]
	em.mu.RUnlock()
	if !ok {
		return 0, errors.NotFound("player treasury")
	}

	// Calculate income (simplified: use gold production)
	income := treasury.GetGold() / 10 // 10% of current gold as income
	tax, err := em.tax.CollectTaxes(playerID, income)
	if err != nil {
		return 0, err
	}

	// Deduct tax from treasury
	if err := treasury.SpendGold(tax); err != nil {
		return 0, err
	}

	em.log.Debug("taxes collected", "player", playerID, "amount", tax)
	return tax, nil
}

// GetBalance returns the current resource balances for a player.
func (em *economyManager) GetBalance(playerID string) (map[string]int, error) {
	em.mu.RLock()
	treasury, ok := em.treasuries[playerID]
	em.mu.RUnlock()
	if !ok {
		return nil, errors.NotFound("player treasury")
	}

	resources := treasury.GetResources()
	// Convert ResourceSet to map[string]int
	result := make(map[string]int)
	for typ, amount := range resources {
		result[string(typ)] = amount
	}
	result["gold"] = treasury.GetGold()
	return result, nil
}

// EnsureTreasury creates a treasury for a player if it doesn't exist.
func (em *economyManager) EnsureTreasury(playerID string) *Treasury {
	em.mu.Lock()
	defer em.mu.Unlock()
	if treasury, ok := em.treasuries[playerID]; ok {
		return treasury
	}
	treasury := NewTreasury(playerID, 100, ResourceSet{
		ResourceWood:  50,
		ResourceStone: 30,
		ResourceFood:  100,
	})
	em.treasuries[playerID] = treasury
	em.log.Info("treasury created", "player", playerID)
	return treasury
}

// GetMarket returns the market instance.
func (em *economyManager) GetMarket() *Market {
	return em.market
}

// GetTaxManager returns the tax manager instance.
func (em *economyManager) GetTaxManager() *TaxManager {
	return em.tax
}

// GetTradeManager returns the trade manager instance.
func (em *economyManager) GetTradeManager() *TradeManager {
	return em.trade
}
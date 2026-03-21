package economy

import (
	"sync"
	"time"

	"ssh-arena-app/pkg/logger"
)

// TaxRate represents a tax rate (0‑1) for a player.
type TaxRate float64

// TaxManager handles tax collection from citizens and buildings.
type TaxManager struct {
	rates      map[string]TaxRate // playerID -> tax rate
	collected  map[string]int     // playerID -> total collected this period
	mu         sync.RWMutex
	log        logger.Logger
	lastCollect time.Time
}

// NewTaxManager creates a new tax manager.
func NewTaxManager() *TaxManager {
	return &TaxManager{
		rates:     make(map[string]TaxRate),
		collected: make(map[string]int),
		log:       logger.Get(),
		lastCollect: time.Now(),
	}
}

// SetTaxRate sets the tax rate for a player.
func (tm *TaxManager) SetTaxRate(playerID string, rate TaxRate) error {
	if rate < 0 || rate > 1 {
		return ErrInvalidTaxRate
	}
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.rates[playerID] = rate
	tm.log.Info("tax rate set", "player", playerID, "rate", rate)
	return nil
}

// GetTaxRate returns the tax rate for a player.
func (tm *TaxManager) GetTaxRate(playerID string) TaxRate {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if rate, ok := tm.rates[playerID]; ok {
		return rate
	}
	return 0.1 // default 10%
}

// CollectTaxes calculates and collects taxes based on player's income.
// income is the total gold income from production, trade, etc.
func (tm *TaxManager) CollectTaxes(playerID string, income int) (int, error) {
	if income < 0 {
		return 0, ErrInvalidIncome
	}
	rate := tm.GetTaxRate(playerID)
	tax := int(float64(income) * float64(rate))
	tm.mu.Lock()
	tm.collected[playerID] += tax
	tm.mu.Unlock()
	tm.log.Debug("taxes collected", "player", playerID, "income", income, "rate", rate, "tax", tax)
	return tax, nil
}

// GetCollected returns the total taxes collected for a player in the current period.
func (tm *TaxManager) GetCollected(playerID string) int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.collected[playerID]
}

// ResetPeriod resets collected taxes for all players (e.g., at the end of a month).
func (tm *TaxManager) ResetPeriod() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.collected = make(map[string]int)
	tm.lastCollect = time.Now()
	tm.log.Info("tax period reset")
}

// CalculateHappinessImpact returns a happiness modifier based on tax rate.
// High tax rates decrease happiness.
func (tm *TaxManager) CalculateHappinessImpact(playerID string) float64 {
	rate := tm.GetTaxRate(playerID)
	// Linear impact: happiness = 1 - rate*0.5 (max 50% reduction at 100% tax)
	impact := 1.0 - float64(rate)*0.5
	if impact < 0.3 {
		impact = 0.3
	}
	return impact
}

// Errors
var (
	ErrInvalidTaxRate = newTaxError("invalid tax rate (must be between 0 and 1)")
	ErrInvalidIncome  = newTaxError("income cannot be negative")
)

type taxError string

func newTaxError(msg string) error {
	return taxError(msg)
}

func (e taxError) Error() string {
	return string(e)
}
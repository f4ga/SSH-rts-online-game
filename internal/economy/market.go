package economy

import (
	"sync"
	"time"

	"ssh-arena-app/pkg/logger"
)

// MarketPrice represents the current price of a resource.
type MarketPrice struct {
	Resource   ResourceType
	BasePrice  float64 // nominal price
	Current    float64 // current price after fluctuations
	Demand     float64 // current demand (0‑1 normalized)
	Supply     float64 // current supply (0‑1 normalized)
	LastUpdate time.Time
}

// Market implements a dynamic price market based on supply and demand.
type Market struct {
	prices   map[ResourceType]*MarketPrice
	mu       sync.RWMutex
	log      logger.Logger
}

// NewMarket creates a new market with default prices.
func NewMarket() *Market {
	m := &Market{
		prices: make(map[ResourceType]*MarketPrice),
		log:    logger.Get(),
	}
	// Initialize prices for all resource types
	defaults := map[ResourceType]float64{
		ResourceGold:    1.0,
		ResourceWood:    0.5,
		ResourceStone:   0.7,
		ResourceFood:    0.3,
		ResourceIron:    1.2,
		ResourceCoal:    0.9,
		ResourceCrystal: 5.0,
	}
	for typ, base := range defaults {
		m.prices[typ] = &MarketPrice{
			Resource:   typ,
			BasePrice:  base,
			Current:    base,
			Demand:     0.5,
			Supply:     0.5,
			LastUpdate: time.Now(),
		}
	}
	return m
}

// GetPrice returns the current price for a resource.
func (m *Market) GetPrice(resource ResourceType) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.prices[resource]; ok {
		return p.Current
	}
	return 1.0 // fallback
}

// UpdateSupplyDemand adjusts supply and demand values.
// supplyChange and demandChange are in range [-1, 1] indicating relative change.
func (m *Market) UpdateSupplyDemand(resource ResourceType, supplyChange, demandChange float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.prices[resource]
	if !ok {
		return
	}

	// Clamp changes
	supplyChange = clamp(supplyChange, -1, 1)
	demandChange = clamp(demandChange, -1, 1)

	// Update supply/demand (simple moving average)
	p.Supply = clamp(p.Supply+supplyChange*0.1, 0.01, 0.99)
	p.Demand = clamp(p.Demand+demandChange*0.1, 0.01, 0.99)

	// Recalculate price: price = base * (demand / supply)
	ratio := p.Demand / p.Supply
	// Limit ratio to avoid extreme prices
	ratio = clamp(ratio, 0.2, 5.0)
	p.Current = p.BasePrice * ratio

	p.LastUpdate = time.Now()
	m.log.Debug("market updated",
		"resource", resource,
		"supply", p.Supply,
		"demand", p.Demand,
		"price", p.Current,
	)
}

// Buy calculates cost to buy a quantity of a resource.
func (m *Market) Buy(resource ResourceType, quantity int) float64 {
	price := m.GetPrice(resource)
	total := price * float64(quantity)
	// Increase demand slightly
	m.UpdateSupplyDemand(resource, 0.0, 0.05)
	return total
}

// Sell calculates revenue from selling a quantity of a resource.
func (m *Market) Sell(resource ResourceType, quantity int) float64 {
	price := m.GetPrice(resource)
	total := price * float64(quantity)
	// Increase supply slightly
	m.UpdateSupplyDemand(resource, 0.05, 0.0)
	return total
}

// GetPriceHistory returns a snapshot of all current prices.
func (m *Market) GetPriceHistory() map[ResourceType]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[ResourceType]float64)
	for typ, p := range m.prices {
		result[typ] = p.Current
	}
	return result
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// CalculateTax calculates tax based on income and tax rate.
func CalculateTax(income float64, taxRate float64) float64 {
	if taxRate < 0 {
		taxRate = 0
	}
	if taxRate > 1 {
		taxRate = 1
	}
	return income * taxRate
}

// ApplyInflation adjusts prices over time (called periodically).
func (m *Market) ApplyInflation(inflationRate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.prices {
		p.BasePrice *= (1 + inflationRate)
		p.Current = p.BasePrice * (p.Demand / p.Supply)
	}
}
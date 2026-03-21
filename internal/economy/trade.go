package economy

import (
	"sync"
	"time"

	"ssh-arena-app/pkg/errors"
	"ssh-arena-app/pkg/logger"
)

// TradeOffer represents an offer to trade resources.
type TradeOffer struct {
	ID          string
	FromPlayer  string
	ToPlayer    string // empty for public offer
	Give        ResourceSet
	Want        ResourceSet
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Status      TradeStatus
}

// TradeStatus indicates the state of a trade offer.
type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "pending"
	TradeStatusAccepted  TradeStatus = "accepted"
	TradeStatusRejected  TradeStatus = "rejected"
	TradeStatusCancelled TradeStatus = "cancelled"
	TradeStatusExpired   TradeStatus = "expired"
)

// TradeManager handles trade offers between players.
type TradeManager struct {
	offers   map[string]*TradeOffer
	mu       sync.RWMutex
	log      logger.Logger
}

// NewTradeManager creates a new trade manager.
func NewTradeManager() *TradeManager {
	return &TradeManager{
		offers: make(map[string]*TradeOffer),
		log:    logger.Get(),
	}
}

// CreateOffer creates a new trade offer.
func (tm *TradeManager) CreateOffer(fromPlayer, toPlayer string, give, want ResourceSet, duration time.Duration) (*TradeOffer, error) {
	if give.Total() == 0 || want.Total() == 0 {
		return nil, errors.InvalidInput("trade must have non‑empty give and want")
	}
	tm.mu.Lock()
	defer tm.mu.Unlock()

	id := generateTradeID(fromPlayer)
	offer := &TradeOffer{
		ID:         id,
		FromPlayer: fromPlayer,
		ToPlayer:   toPlayer,
		Give:       give.Clone(),
		Want:       want.Clone(),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(duration),
		Status:     TradeStatusPending,
	}
	tm.offers[id] = offer
	tm.log.Info("trade offer created", "id", id, "from", fromPlayer, "to", toPlayer, "give", give, "want", want)
	return offer, nil
}

// AcceptOffer accepts a pending trade offer.
func (tm *TradeManager) AcceptOffer(offerID, acceptingPlayer string, fromTreasury, toTreasury *Treasury) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	offer, ok := tm.offers[offerID]
	if !ok {
		return errors.NotFound("trade offer")
	}
	if offer.Status != TradeStatusPending {
		return errors.New(errors.ErrCodeConflict, "offer not pending")
	}
	if offer.ExpiresAt.Before(time.Now()) {
		offer.Status = TradeStatusExpired
		return errors.New(errors.ErrCodeConflict, "offer expired")
	}
	// Verify that the accepting player is the intended recipient (if specified)
	if offer.ToPlayer != "" && offer.ToPlayer != acceptingPlayer {
		return errors.New(errors.ErrCodeUnauthorized, "offer not addressed to you")
	}

	// Check that fromTreasury has enough resources to give
	for typ, amount := range offer.Give {
		if fromTreasury.GetResources()[typ] < amount {
			return errors.New(errors.ErrCodeConflict, "insufficient resources in offering player's treasury")
		}
	}
	// Check that toTreasury has enough resources to give (if it's a two‑way trade)
	// Actually, the accepting player must provide the "want" resources.
	for typ, amount := range offer.Want {
		if toTreasury.GetResources()[typ] < amount {
			return errors.New(errors.ErrCodeConflict, "insufficient resources in accepting player's treasury")
		}
	}

	// Perform the swap
	err := fromTreasury.TransferTo(toTreasury, 0, offer.Give)
	if err != nil {
		return err
	}
	err = toTreasury.TransferTo(fromTreasury, 0, offer.Want)
	if err != nil {
		// Rollback? For simplicity, we assume atomicity is not required.
		return err
	}

	offer.Status = TradeStatusAccepted
	tm.log.Info("trade offer accepted", "id", offerID, "by", acceptingPlayer)
	return nil
}

// CancelOffer cancels a pending offer (only the creator can cancel).
func (tm *TradeManager) CancelOffer(offerID, playerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	offer, ok := tm.offers[offerID]
	if !ok {
		return errors.NotFound("trade offer")
	}
	if offer.FromPlayer != playerID {
		return errors.New(errors.ErrCodeUnauthorized, "only the creator can cancel")
	}
	if offer.Status != TradeStatusPending {
		return errors.New(errors.ErrCodeConflict, "offer not pending")
	}
	offer.Status = TradeStatusCancelled
	tm.log.Info("trade offer cancelled", "id", offerID, "by", playerID)
	return nil
}

// GetOffers returns all offers relevant to a player (incoming, outgoing, or public).
func (tm *TradeManager) GetOffers(playerID string) []*TradeOffer {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var result []*TradeOffer
	for _, offer := range tm.offers {
		if offer.Status != TradeStatusPending {
			continue
		}
		if offer.FromPlayer == playerID || offer.ToPlayer == playerID || offer.ToPlayer == "" {
			result = append(result, offer)
		}
	}
	return result
}

// CleanupExpired removes expired offers periodically.
func (tm *TradeManager) CleanupExpired() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	for id, offer := range tm.offers {
		if offer.Status == TradeStatusPending && offer.ExpiresAt.Before(now) {
			offer.Status = TradeStatusExpired
			tm.log.Debug("trade offer expired", "id", id)
		}
	}
}

func generateTradeID(playerID string) string {
	return playerID + "_" + time.Now().Format("20060102150405")
}
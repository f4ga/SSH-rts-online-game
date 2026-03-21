package research

import (
	"sync"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/errors"
	"ssh-arena-app/pkg/logger"
)

// researchManager implements internal.ResearchManager.
type researchManager struct {
	mu          sync.RWMutex
	techTree    []*TechNode
	researching map[string]*Technology // playerID -> currently researching tech
	completed   map[string][]*Technology // playerID -> slice of completed techs
	eventBus    internal.EventBus
	log         logger.Logger
}

// NewResearchManager creates a new research manager with the default tech tree.
func NewResearchManager(eventBus internal.EventBus) internal.ResearchManager {
	return &researchManager{
		techTree:    DefaultTechTree(),
		researching: make(map[string]*Technology),
		completed:   make(map[string][]*Technology),
		eventBus:    eventBus,
		log:         logger.Get(),
	}
}

// StartResearch begins researching a technology for a player.
func (rm *researchManager) StartResearch(playerID, techID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check if already researching something
	if current, ok := rm.researching[playerID]; ok && !current.Completed {
		return errors.New(errors.ErrCodeConflict, "already researching another technology")
	}

	// Find tech node
	var node *TechNode
	for _, n := range rm.techTree {
		if n.ID == techID {
			node = n
			break
		}
	}
	if node == nil {
		return errors.NotFound("technology")
	}

	// Check prerequisites
	if !rm.hasPrerequisites(playerID, node) {
		return errors.New(errors.ErrCodeConflict, "prerequisites not met")
	}

	// Check cost (simplified: assume player has resources via economy manager)
	// In a real implementation, you would deduct resources.
	// For now, we just log.

	// Create technology progress
	tech := &Technology{
		Node:      node,
		PlayerID:  playerID,
		Progress:  0,
		Completed: false,
		StartedAt: time.Now(),
	}
	rm.researching[playerID] = tech

	rm.log.Info("research started", "player", playerID, "tech", techID, "duration", node.Duration)
	return nil
}

// CancelResearch cancels ongoing research.
func (rm *researchManager) CancelResearch(playerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	tech, ok := rm.researching[playerID]
	if !ok || tech.Completed {
		return errors.New(errors.ErrCodeConflict, "no active research")
	}

	delete(rm.researching, playerID)
	rm.log.Info("research cancelled", "player", playerID, "tech", tech.Node.ID)
	return nil
}

// GetDiscovered returns all technologies discovered by a player.
func (rm *researchManager) GetDiscovered(playerID string) ([]*internal.Technology, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var result []*internal.Technology
	// Add completed techs
	for _, tech := range rm.completed[playerID] {
		result = append(result, &internal.Technology{
			ID:          tech.Node.ID,
			Name:        tech.Node.Name,
			Description: tech.Node.Description,
			Cost:        int(tech.Node.Cost.Total()),
			Researched:  true,
		})
	}
	// Add currently researching tech
	if tech, ok := rm.researching[playerID]; ok && !tech.Completed {
		result = append(result, &internal.Technology{
			ID:          tech.Node.ID,
			Name:        tech.Node.Name,
			Description: tech.Node.Description,
			Cost:        int(tech.Node.Cost.Total()),
			Researched:  false,
		})
	}
	return result, nil
}

// UpdateResearch advances research progress (called each tick).
func (rm *researchManager) UpdateResearch(delta time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for playerID, tech := range rm.researching {
		if tech.Completed {
			continue
		}
		completed := tech.Advance(delta)
		if completed {
			rm.finishResearch(playerID, tech)
		}
	}
}

// finishResearch marks a technology as completed and applies its effects.
func (rm *researchManager) finishResearch(playerID string, tech *Technology) {
	tech.Completed = true
	// Move from researching to completed
	rm.completed[playerID] = append(rm.completed[playerID], tech)
	delete(rm.researching, playerID)

	// Apply effects
	rm.applyEffects(playerID, tech.Node.Effects)

	// Publish event
	if rm.eventBus != nil {
		rm.eventBus.Publish(internal.Event{
			Type:      "TechResearchedEvent",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"playerID": playerID,
				"techID":   tech.Node.ID,
				"effects":  tech.Node.Effects,
			},
		})
	}

	rm.log.Info("research completed", "player", playerID, "tech", tech.Node.ID)
}

// applyEffects applies the effects of a technology to the game state.
func (rm *researchManager) applyEffects(playerID string, effects []TechEffect) {
	// In a full implementation, you would modify game state (e.g., unlock units, boost production).
	// For now, we just log.
	for _, effect := range effects {
		rm.log.Debug("applying tech effect", "player", playerID, "effect", effect.Type, "target", effect.Target, "value", effect.Value)
	}
}

// hasPrerequisites checks if a player has researched all required technologies.
func (rm *researchManager) hasPrerequisites(playerID string, node *TechNode) bool {
	completedSet := make(map[string]bool)
	for _, tech := range rm.completed[playerID] {
		completedSet[tech.Node.ID] = true
	}
	for _, req := range node.Requirements {
		if !completedSet[req] {
			return false
		}
	}
	return true
}

// IsResearched returns true if the player has researched the given technology.
func (rm *researchManager) IsResearched(playerID, techID string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Check completed
	for _, tech := range rm.completed[playerID] {
		if tech.Node.ID == techID {
			return true
		}
	}
	// Check researching (but not completed)
	if tech, ok := rm.researching[playerID]; ok && tech.Node.ID == techID && tech.Completed {
		return true
	}
	return false
}

// GetTechNode returns the TechNode for a given ID.
func (rm *researchManager) GetTechNode(techID string) (*TechNode, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	for _, node := range rm.techTree {
		if node.ID == techID {
			return node, nil
		}
	}
	return nil, errors.NotFound("technology")
}
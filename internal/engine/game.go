// Package engine implements the core game loop and state management.
package engine

import (
	"context"
	"sync"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/errors"
	"ssh-arena-app/pkg/logger"
)

// DefaultTickInterval is the duration between game ticks if not configured.
const DefaultTickInterval = 50 * time.Millisecond // 20 ticks per second

// commandEnvelope wraps a command with its player ID for asynchronous processing.
type commandEnvelope struct {
	playerID string
	cmd      internal.Command
}

// Game implements the GameEngine interface.
type Game struct {
	world   internal.World
	players map[string]*internal.Player
	mu      sync.RWMutex

	building internal.BuildingManager
	citizen  internal.CitizenManager
	combat   internal.CombatManager
	economy  internal.EconomyManager
	research internal.ResearchManager
	events   internal.EventBus
	storage  internal.Storage

	commandsChan chan commandEnvelope
	ctx          context.Context
	cancel       context.CancelFunc

	tickInterval time.Duration
	ticker       *time.Ticker
	stopChan     chan struct{}
	running      bool

	log logger.Logger
}

// NewGame creates a new Game instance with the required dependencies.
func NewGame(
	world internal.World,
	building internal.BuildingManager,
	citizen internal.CitizenManager,
	combat internal.CombatManager,
	economy internal.EconomyManager,
	research internal.ResearchManager,
	events internal.EventBus,
	storage internal.Storage,
	tickInterval time.Duration,
) *Game {
	if tickInterval <= 0 {
		tickInterval = DefaultTickInterval
	}
	return &Game{
		world:        world,
		players:      make(map[string]*internal.Player),
		building:     building,
		citizen:      citizen,
		combat:       combat,
		economy:      economy,
		research:     research,
		events:       events,
		storage:      storage,
		commandsChan: make(chan commandEnvelope, 100),
		ctx:          nil,
		cancel:       nil,
		tickInterval: tickInterval,
		stopChan:     make(chan struct{}),
		log:          logger.Get(),
	}
}

// Start begins the game loop in a separate goroutine.
func (g *Game) Start(ctx context.Context) error {
	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return errors.New(errors.ErrCodeConflict, "game already running")
	}
	g.running = true
	g.ticker = time.NewTicker(g.tickInterval)
	g.mu.Unlock()

	g.log.Info("game engine started", "tick_interval", g.tickInterval)

	go g.loop(ctx)
	return nil
}

// Stop gracefully stops the game loop and saves state.
func (g *Game) Stop(ctx context.Context) error {
	g.mu.Lock()
	if !g.running {
		g.mu.Unlock()
		return nil
	}
	g.running = false
	if g.ticker != nil {
		g.ticker.Stop()
	}
	close(g.stopChan)
	g.mu.Unlock()

	// Wait for loop to finish
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		g.log.Warn("game stop timed out")
	}

	// Save game state
	if err := g.saveState(); err != nil {
		g.log.Error("failed to save game state on stop", "error", err)
	}

	g.log.Info("game engine stopped")
	return nil
}

// ProcessCommand handles a player command.
func (g *Game) ProcessCommand(playerID string, cmd internal.Command) (internal.Response, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.running {
		return internal.Response{Success: false, Message: "game not running"}, errors.New(errors.ErrCodeUnavailable, "game not running")
	}

	player, exists := g.players[playerID]
	if !exists {
		return internal.Response{Success: false, Message: "player not found"}, errors.NotFound("player")
	}

	g.log.Debug("processing command", "player", playerID, "command", cmd.Type)

	switch cmd.Type {
	case "build":
		return g.handleBuildCommand(player, cmd)
	case "move":
		return g.handleMoveCommand(player, cmd)
	case "attack":
		return g.handleAttackCommand(player, cmd)
	case "research":
		return g.handleResearchCommand(player, cmd)
	default:
		return internal.Response{Success: false, Message: "unknown command"}, errors.InvalidInput("command type")
	}
}

// SubmitCommand sends a command to the game engine for asynchronous processing.
func (g *Game) SubmitCommand(playerID string, cmd internal.Command) error {
	if !g.running {
		return errors.New(errors.ErrCodeUnavailable, "game not running")
	}
	env := commandEnvelope{playerID: playerID, cmd: cmd}
	select {
	case g.commandsChan <- env:
		g.log.Debug("command submitted", "player", playerID, "command", cmd.Type)
		return nil
	default:
		return errors.New(errors.ErrCodeUnavailable, "command queue full")
	}
}

// Tick advances the game simulation by one tick.
func (g *Game) Tick() {
	g.mu.Lock()
	defer g.mu.Unlock()

	delta := g.tickInterval

	// Update world
	if g.world != nil {
		g.world.Update(delta)
	}

	// Update subsystems (if they exist)
	if g.citizen != nil {
		g.citizen.UpdateNeeds(delta)
	}
	if g.combat != nil {
		g.combat.UpdateCombat(delta)
	}
	if g.economy != nil {
		g.economy.Produce(delta)
	}
	if g.research != nil {
		g.research.UpdateResearch(delta)
	}

	// Emit tick event (if event bus exists)
	if g.events != nil {
		g.events.Publish(internal.Event{
			Type:      "tick",
			Timestamp: time.Now(),
			Payload:   delta,
		})
	}
}

// GetState returns a snapshot of the game state for a given player.
func (g *Game) GetState(playerID string) (*internal.GameState, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.players[playerID]; !exists {
		return nil, errors.NotFound("player")
	}

	return &internal.GameState{
		World:     g.world,
		Players:   g.players,
		Timestamp: time.Now(),
	}, nil
}

// loop runs the main game loop.
func (g *Game) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			g.log.Info("game loop cancelled via context")
			return
		case <-g.stopChan:
			g.log.Info("game loop stopped via stop channel")
			return
		case env := <-g.commandsChan:
			g.processCommandAsync(env.playerID, env.cmd)
		case <-g.ticker.C:
			g.Tick()
		}
	}
}

// processCommandAsync processes a command received from the command channel.
func (g *Game) processCommandAsync(playerID string, cmd internal.Command) {
	g.log.Debug("processing async command", "player", playerID, "command", cmd.Type)
	// Call the synchronous command processor.
	// We ignore the response for now, but could log errors.
	_, err := g.ProcessCommand(playerID, cmd)
	if err != nil {
		g.log.Error("failed to process command", "player", playerID, "error", err)
	}
}

func (g *Game) handleBuildCommand(player *internal.Player, cmd internal.Command) (internal.Response, error) {
	// Simplified implementation
	g.log.Info("build command", "player", player.ID, "payload", cmd.Payload)
	return internal.Response{Success: true, Message: "building queued"}, nil
}

func (g *Game) handleMoveCommand(player *internal.Player, cmd internal.Command) (internal.Response, error) {
	g.log.Info("move command", "player", player.ID, "payload", cmd.Payload)
	return internal.Response{Success: true, Message: "unit moving"}, nil
}

func (g *Game) handleAttackCommand(player *internal.Player, cmd internal.Command) (internal.Response, error) {
	g.log.Info("attack command", "player", player.ID, "payload", cmd.Payload)
	return internal.Response{Success: true, Message: "attack ordered"}, nil
}

func (g *Game) handleResearchCommand(player *internal.Player, cmd internal.Command) (internal.Response, error) {
	g.log.Info("research command", "player", player.ID, "payload", cmd.Payload)
	return internal.Response{Success: true, Message: "research started"}, nil
}

func (g *Game) saveState() error {
	// TODO: implement proper serialization
	g.log.Debug("saving game state")
	return nil
}
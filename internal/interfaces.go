// Package internal defines core interfaces for the SSH Arena game engine.
// These interfaces follow the dependency inversion principle, allowing
// different implementations to be swapped (e.g., for testing).
package internal

import (
	"context"
	"time"

	"ssh-arena-app/internal/world"
)

// GameEngine is the central orchestrator of the game loop.
// It manages the game state, processes commands, and drives updates.
type GameEngine interface {
	// Start begins the game loop and any background goroutines.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the engine, saving state if needed.
	Stop(ctx context.Context) error
	// ProcessCommand handles a player command, returning a response or error.
	ProcessCommand(playerID string, cmd Command) (Response, error)
	// Tick advances the game simulation by one tick.
	Tick()
	// GetState returns a snapshot of the game state for a given player.
	GetState(playerID string) (*GameState, error)
}

// World represents the game world: tiles, chunks, and terrain.
type World interface {
	// GetTile returns the tile at the given coordinates.
	GetTile(x, y int) (*world.Tile, error)
	// GetChunk returns the chunk containing the given coordinates.
	GetChunk(x, y int) (*world.Chunk, error)
	// SetTile modifies a tile (e.g., building placement).
	SetTile(x, y int, tile *world.Tile) error
	// Generate generates the world using the provided seed.
	Generate(seed int64) error
	// Save serializes the world to persistent storage.
	Save() error
	// Load restores the world from storage.
	Load() error
	// Update advances the world simulation by delta time.
	Update(delta time.Duration)
}

// BuildingManager handles construction, upgrading, and demolition of buildings.
type BuildingManager interface {
	// Construct places a new building at the specified location.
	Construct(playerID string, blueprintID string, x, y int) (*Building, error)
	// Upgrade increases the level of an existing building.
	Upgrade(buildingID string) error
	// Demolish removes a building, possibly refunding resources.
	Demolish(buildingID string) error
	// GetBuildings returns all buildings owned by a player.
	GetBuildings(playerID string) ([]*Building, error)
}

// CitizenManager manages population, needs, jobs, and AI.
type CitizenManager interface {
	// Spawn creates a new citizen for a player.
	Spawn(playerID string, locationX, locationY int) (*Citizen, error)
	// AssignJob assigns a citizen to a job (building, resource, etc.).
	AssignJob(citizenID, jobID string) error
	// UpdateNeeds updates the needs of all citizens (called each tick).
	UpdateNeeds(delta time.Duration)
	// GetCitizens returns all citizens belonging to a player.
	GetCitizens(playerID string) ([]*Citizen, error)
}

// CombatManager handles unit movement, attacks, and battles.
type CombatManager interface {
	// CreateUnit creates a military unit for a player.
	CreateUnit(playerID, unitType string, x, y int) (*Unit, error)
	// MoveUnit orders a unit to move to a target location.
	MoveUnit(unitID string, targetX, targetY int) error
	// Attack orders a unit to attack another unit or building.
	Attack(attackerID, targetID string) error
	// UpdateCombat processes combat for all units (called each tick).
	UpdateCombat(delta time.Duration)
	// GetUnits returns all units belonging to a player.
	GetUnits(playerID string) ([]*Unit, error)
}

// EconomyManager manages resources, trade, taxes, and player treasury.
type EconomyManager interface {
	// Produce generates resources based on buildings and citizens.
	Produce(delta time.Duration)
	// Transfer transfers resources from one player to another (trade).
	Transfer(fromPlayerID, toPlayerID string, resources map[string]int) error
	// CollectTaxes collects taxes from citizens and buildings.
	CollectTaxes(playerID string) (int, error)
	// GetBalance returns the current resource balances for a player.
	GetBalance(playerID string) (map[string]int, error)
}

// ResearchManager handles technology tree and research progress.
type ResearchManager interface {
	// StartResearch begins researching a technology for a player.
	StartResearch(playerID, techID string) error
	// CancelResearch cancels ongoing research.
	CancelResearch(playerID string) error
	// GetDiscovered returns all technologies discovered by a player.
	GetDiscovered(playerID string) ([]*Technology, error)
	// UpdateResearch advances research progress (called each tick).
	UpdateResearch(delta time.Duration)
}

// EventBus is a publish‑subscribe system for game events.
type EventBus interface {
	// Publish sends an event to all subscribed listeners.
	Publish(event Event)
	// Subscribe adds a listener for events of the given type.
	Subscribe(eventType string, listener EventListener)
	// Unsubscribe removes a listener.
	Unsubscribe(eventType string, listener EventListener)
}

// NetworkTransport defines how the server communicates with clients.
type NetworkTransport interface {
	// Start begins listening for client connections.
	Start() error
	// Stop closes all connections and stops listening.
	Stop() error
	// Send sends a message to a specific client.
	Send(clientID string, msg []byte) error
	// Broadcast sends a message to all connected clients.
	Broadcast(msg []byte)
	// Receive returns a channel of incoming client messages.
	Receive() <-chan ClientMessage
}

// Renderer produces a visual representation of the game state for a client.
type Renderer interface {
	// Render generates a frame (ASCII/ANSI) for the given player viewport.
	Render(playerID string, viewport Viewport) ([]byte, error)
	// SetTheme changes the rendering theme (colors, symbols).
	SetTheme(theme Theme) error
}

// Storage provides persistence for game data (save/load).
type Storage interface {
	// SaveGame stores the entire game state.
	SaveGame(state *GameState) error
	// LoadGame restores a game state by ID.
	LoadGame(gameID string) (*GameState, error)
	// SavePlayer stores player‑specific data.
	SavePlayer(player *Player) error
	// LoadPlayer retrieves a player by ID.
	LoadPlayer(playerID string) (*Player, error)
	// Close releases any resources (e.g., database connections).
	Close() error
}

// Supporting types (simplified for brevity)

type Command struct {
	Type    string
	Payload map[string]interface{}
}

type Response struct {
	Success bool
	Message string
	Data    interface{}
}

type GameState struct {
	World      World
	Players    map[string]*Player
	Timestamp  time.Time
}

type Player struct {
	ID       string
	Name     string
	Credits  int
	Location struct{ X, Y int }
}

type Building struct {
	ID         string
	PlayerID   string
	BlueprintID string
	Level      int
	Health     int
	Position   struct{ X, Y int }
}

type Citizen struct {
	ID       string
	PlayerID string
	Name     string
	Age      int
	Health   int
	Hunger   int
	JobID    string
	Position struct{ X, Y int }
}

type Unit struct {
	ID       string
	PlayerID string
	Type     string
	Health   int
	Attack   int
	Defense  int
	Position struct{ X, Y int }
	Target   *Unit
}

type Technology struct {
	ID          string
	Name        string
	Description string
	Cost        int
	Researched  bool
}

type Event struct {
	Type      string
	Timestamp time.Time
	Payload   interface{}
}

type EventListener func(Event)

type ClientMessage struct {
	ClientID string
	Data     []byte
}

type Viewport struct {
	X, Y, Width, Height int
}

type Theme struct {
	Colors map[string]string
	Symbols map[string]rune
}
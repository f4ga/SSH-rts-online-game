т// Package storage provides persistent storage for game state.
package storage

import (
	"ssh-arena-app/internal"
	"time"
)

// GameState represents the complete state of the game world.
type GameState struct {
	// World is the game world (chunks, tiles).
	World internal.World `json:"-"` // not serialized directly; we store its raw data
	// Players list.
	Players map[string]*internal.Player `json:"players"`
	// Buildings list.
	Buildings []*internal.Building `json:"buildings"`
	// Citizens list.
	Citizens []*internal.Citizen `json:"citizens"`
	// Units list.
	Units []*internal.Unit `json:"units"`
	// ResearchProgress maps player ID to technology IDs.
	ResearchProgress map[string][]string `json:"research_progress"`
	// EconomyState includes resource balances, market, treasury.
	EconomyState map[string]interface{} `json:"economy_state"`
	// EventLog is a circular buffer of recent events.
	EventLog []internal.Event `json:"event_log"`
	// Timestamp when the state was saved.
	Timestamp time.Time `json:"timestamp"`
}

// Storage is the interface for saving and loading game state.
type Storage interface {
	// Save persists the game state.
	Save(gameState *GameState) error
	// Load retrieves the most recent game state.
	Load() (*GameState, error)
}
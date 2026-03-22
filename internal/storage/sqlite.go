package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/logger"
)

// SQLiteStorage implements Storage using SQLite3.
type SQLiteStorage struct {
	db     *sql.DB
	mu     sync.RWMutex
	log    logger.Logger
	dbPath string
}

// NewSQLiteStorage creates a new SQLiteStorage instance.
// It opens the database and ensures the schema exists.
func NewSQLiteStorage(dsn string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &SQLiteStorage{
		db:     db,
		log:    logger.Get(),
		dbPath: dsn,
	}

	if err := s.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return s, nil
}

// init creates the necessary tables if they don't exist.
func (s *SQLiteStorage) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS game_state (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER NOT NULL DEFAULT 1,
			timestamp DATETIME NOT NULL,
			data BLOB NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS players (
			player_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			credits INTEGER NOT NULL DEFAULT 0,
			location_x INTEGER NOT NULL DEFAULT 0,
			location_y INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS buildings (
			building_id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			blueprint_id TEXT NOT NULL,
			level INTEGER NOT NULL DEFAULT 1,
			health INTEGER NOT NULL DEFAULT 100,
			position_x INTEGER NOT NULL,
			position_y INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (player_id) REFERENCES players(player_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS citizens (
			citizen_id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			name TEXT NOT NULL,
			age INTEGER NOT NULL DEFAULT 0,
			health INTEGER NOT NULL DEFAULT 100,
			hunger INTEGER NOT NULL DEFAULT 0,
			job_id TEXT,
			position_x INTEGER NOT NULL,
			position_y INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (player_id) REFERENCES players(player_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS units (
			unit_id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			type TEXT NOT NULL,
			health INTEGER NOT NULL DEFAULT 100,
			attack INTEGER NOT NULL DEFAULT 10,
			defense INTEGER NOT NULL DEFAULT 5,
			position_x INTEGER NOT NULL,
			position_y INTEGER NOT NULL,
			target_unit_id TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (player_id) REFERENCES players(player_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS research_progress (
			player_id TEXT NOT NULL,
			tech_id TEXT NOT NULL,
			progress REAL NOT NULL DEFAULT 0.0,
			started_at DATETIME NOT NULL,
			PRIMARY KEY (player_id, tech_id),
			FOREIGN KEY (player_id) REFERENCES players(player_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS economy_state (
			player_id TEXT PRIMARY KEY,
			resources_json TEXT NOT NULL,
			market_json TEXT NOT NULL,
			tax_rate REAL NOT NULL DEFAULT 0.05,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (player_id) REFERENCES players(player_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS event_log (
			event_id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_type TEXT NOT NULL,
			player_id TEXT,
			payload_json TEXT NOT NULL,
			timestamp DATETIME NOT NULL
		)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute query %q: %w", q, err)
		}
	}

	s.log.Info("database initialized", "path", s.dbPath)
	return nil
}

// Save persists the game state.
func (s *SQLiteStorage) Save(gameState *GameState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Serialize the entire game state as JSON (for simplicity).
	data, err := json.Marshal(gameState)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	// Insert into game_state table.
	_, err = tx.Exec(
		`INSERT INTO game_state (version, timestamp, data) VALUES (?, ?, ?)`,
		1,
		time.Now().UTC(),
		data,
	)
	if err != nil {
		return fmt.Errorf("failed to insert game state: %w", err)
	}

	// Also update individual tables for faster queries (optional).
	// For now we just keep the snapshot.

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.log.Debug("game state saved", "size", len(data))
	return nil
}

// Load retrieves the most recent game state.
func (s *SQLiteStorage) Load() (*GameState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var data []byte
	var timestamp time.Time
	err := s.db.QueryRow(
		`SELECT data, timestamp FROM game_state ORDER BY id DESC LIMIT 1`,
	).Scan(&data, &timestamp)
	if err == sql.ErrNoRows {
		return nil, nil // No saved state
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query game state: %w", err)
	}

	var state GameState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game state: %w", err)
	}

	state.Timestamp = timestamp
	s.log.Info("game state loaded", "timestamp", timestamp)
	return &state, nil
}

// Close releases the database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// SaveGame stores the entire game state (implements internal.Storage).
func (s *SQLiteStorage) SaveGame(state *internal.GameState) error {
	// Convert internal.GameState to storage.GameState.
	// For now, we only save players and world reference; other fields are empty.
	// This is a temporary implementation; should be expanded later.
	gs := &GameState{
		Players:          state.Players,
		Timestamp:        state.Timestamp,
		Buildings:        nil,
		Citizens:         nil,
		Units:            nil,
		ResearchProgress: make(map[string][]string),
		EconomyState:     make(map[string]interface{}),
		EventLog:         nil,
	}
	// Note: World is not serialized; we would need to serialize chunks.
	// For simplicity, we ignore World for now.
	return s.Save(gs)
}

// LoadGame restores a game state by ID (implements internal.Storage).
func (s *SQLiteStorage) LoadGame(gameID string) (*internal.GameState, error) {
	// We ignore gameID and load the most recent snapshot.
	gs, err := s.Load()
	if err != nil {
		return nil, err
	}
	if gs == nil {
		return nil, nil
	}
	// Convert storage.GameState to internal.GameState.
	// World is nil; we need to restore it from elsewhere.
	state := &internal.GameState{
		World:     nil, // TODO: restore world from separate storage
		Players:   gs.Players,
		Timestamp: gs.Timestamp,
	}
	return state, nil
}

// SavePlayer stores player‑specific data (implements internal.Storage).
func (s *SQLiteStorage) SavePlayer(player *internal.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO players (player_id, name, credits, location_x, location_y, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		player.ID,
		player.Name,
		player.Credits,
		player.Location.X,
		player.Location.Y,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to save player: %w", err)
	}
	return nil
}

// LoadPlayer retrieves a player by ID (implements internal.Storage).
func (s *SQLiteStorage) LoadPlayer(playerID string) (*internal.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var player internal.Player
	var locX, locY int
	err := s.db.QueryRow(
		`SELECT player_id, name, credits, location_x, location_y FROM players WHERE player_id = ?`,
		playerID,
	).Scan(&player.ID, &player.Name, &player.Credits, &locX, &locY)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load player: %w", err)
	}
	player.Location.X = locX
	player.Location.Y = locY
	return &player, nil
}
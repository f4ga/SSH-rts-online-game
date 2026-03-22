package storage

import (
	"fmt"

	"ssh-arena-app/internal"
	"ssh-arena-app/pkg/config"
)

// NewStorage creates a concrete Storage implementation based on configuration.
func NewStorage(cfg *config.DatabaseConfig) (internal.Storage, error) {
	switch cfg.Driver {
	case "sqlite3":
		return NewSQLiteStorage(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// Storage returns a stub implementation of the Storage interface.
// Deprecated: Use NewStorage instead.
// func Storage() internal.Storage {
// 	// For backward compatibility, try to create a default SQLite storage.
// 	// This may fail if config not loaded; returns nil.
// 	cfg := config.Get()
// 	if cfg == nil {
// 		return nil
// 	}
// 	s, err := NewStorage(&cfg.Database)
// 	if err != nil {
// 		return nil
// 	}
// 	return s
// }
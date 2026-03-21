// Package config provides configuration loading and management for the SSH Arena game server.
// It uses viper to read from config files, environment variables, and command-line flags.
package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// Config holds all configuration parameters for the game server.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Game     GameConfig     `mapstructure:"game"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	SSH      SSHConfig      `mapstructure:"ssh"`
}

// ServerConfig defines network and runtime server settings.
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	MaxPlayers   int    `mapstructure:"max_players"`
	TickRate     int    `mapstructure:"tick_rate"`
	SaveInterval int    `mapstructure:"save_interval"`
}

// GameConfig defines game‑specific parameters.
type GameConfig struct {
	WorldWidth      int     `mapstructure:"world_width"`
	WorldHeight     int     `mapstructure:"world_height"`
	ChunkSize       int     `mapstructure:"chunk_size"`
	StartingCredits int     `mapstructure:"starting_credits"`
	TaxRate         float64 `mapstructure:"tax_rate"`
	ResearchSpeed   float64 `mapstructure:"research_speed"`
}

// DatabaseConfig defines storage settings.
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	DSN      string `mapstructure:"dsn"`
	Migrate  bool   `mapstructure:"migrate"`
	PoolSize int    `mapstructure:"pool_size"`
}

// LoggingConfig defines logging behavior.
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	WithCaller bool   `mapstructure:"with_caller"`
}

// SSHConfig defines SSH server parameters.
type SSHConfig struct {
	Port           int    `mapstructure:"port"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
	AuthorizedKeys string `mapstructure:"authorized_keys"`
	Banner         string `mapstructure:"banner"`
	IdleTimeout    int    `mapstructure:"idle_timeout"`
}

var (
	globalConfig *Config
	configOnce   sync.Once
	configMu     sync.RWMutex
)

// Load reads configuration from the given path and environment.
// It returns a populated Config or an error.
func Load(configPath string) (*Config, error) {
	var cfg Config

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setDefaults(&cfg)
	globalConfig = &cfg
	return &cfg, nil
}

// Get returns the global configuration instance.
// It panics if Load has not been called first.
func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	if globalConfig == nil {
		panic("config not loaded; call Load first")
	}
	return globalConfig
}

// Reload reloads the configuration from disk and updates the global instance.
func Reload(configPath string) error {
	configMu.Lock()
	defer configMu.Unlock()

	newCfg, err := Load(configPath)
	if err != nil {
		return err
	}
	globalConfig = newCfg
	return nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.MaxPlayers == 0 {
		cfg.Server.MaxPlayers = 100
	}
	if cfg.Server.TickRate == 0 {
		cfg.Server.TickRate = 20
	}
	if cfg.Server.SaveInterval == 0 {
		cfg.Server.SaveInterval = 300
	}
	if cfg.Game.WorldWidth == 0 {
		cfg.Game.WorldWidth = 1000
	}
	if cfg.Game.WorldHeight == 0 {
		cfg.Game.WorldHeight = 1000
	}
	if cfg.Game.ChunkSize == 0 {
		cfg.Game.ChunkSize = 32
	}
	if cfg.Game.StartingCredits == 0 {
		cfg.Game.StartingCredits = 1000
	}
	if cfg.Game.TaxRate == 0 {
		cfg.Game.TaxRate = 0.05
	}
	if cfg.Game.ResearchSpeed == 0 {
		cfg.Game.ResearchSpeed = 1.0
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "sqlite3"
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "./game.db"
	}
	if cfg.Database.PoolSize == 0 {
		cfg.Database.PoolSize = 10
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}
	if cfg.SSH.Port == 0 {
		cfg.SSH.Port = 2222
	}
	if cfg.SSH.PrivateKeyPath == "" {
		cfg.SSH.PrivateKeyPath = "./ssh_host_key"
	}
	if cfg.SSH.AuthorizedKeys == "" {
		cfg.SSH.AuthorizedKeys = "./authorized_keys"
	}
	if cfg.SSH.IdleTimeout == 0 {
		cfg.SSH.IdleTimeout = 300
	}
}
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/internal/engine"
	"ssh-arena-app/internal/events"
	"ssh-arena-app/internal/world"
	"ssh-arena-app/pkg/config"
	"ssh-arena-app/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	err = logger.Init(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.Output,
		cfg.Logging.WithCaller,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log := logger.Get()

	log.Info("SSH Arena server starting",
		"version", "0.1.0",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	// Create concrete or stub components
	seed := int64(cfg.Game.WorldWidth + cfg.Game.WorldHeight) // simple deterministic seed
	world := world.NewWorld(seed)
	events := events.EventBus() // returns nil stub for now
	// Other managers are still stubs (nil)
	var (
		building  internal.BuildingManager = nil
		citizen   internal.CitizenManager  = nil
		combat    internal.CombatManager   = nil
		economy   internal.EconomyManager  = nil
		research  internal.ResearchManager = nil
		storage   internal.Storage         = nil
		network   internal.NetworkTransport = nil
	)

	// Create game engine
	game := engine.NewGame(
		world,
		building,
		citizen,
		combat,
		economy,
		research,
		events,
		storage,
		time.Duration(cfg.Server.TickRate)*time.Millisecond,
	)

	// Start game engine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := game.Start(ctx); err != nil {
		log.Error("Failed to start game engine", "error", err)
		os.Exit(1)
	}
	log.Info("Game engine started")

	// Start network transport (SSH server)
	if network != nil {
		if err := network.Start(); err != nil {
			log.Error("Failed to start network transport", "error", err)
			// Continue without network? For now, just log.
		}
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Server is running. Press Ctrl+C to stop.")
	<-sigChan

	log.Info("Shutdown signal received, stopping...")

	// Stop game engine
	if err := game.Stop(ctx); err != nil {
		log.Error("Error stopping game engine", "error", err)
	}

	// Stop network
	if network != nil {
		// network.Stop() // TODO: implement Stop method
	}

	log.Info("Server stopped gracefully")
}
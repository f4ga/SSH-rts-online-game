package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/internal/combat"
	"ssh-arena-app/internal/economy"
	"ssh-arena-app/internal/engine"
	"ssh-arena-app/internal/events"
	"ssh-arena-app/internal/network"
	"ssh-arena-app/internal/research"
	"ssh-arena-app/internal/ui"
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
	eventBus := events.EventBus() // теперь реальная реализация

	// Create managers
	var (
		building internal.BuildingManager = nil // TODO: implement
		citizen  internal.CitizenManager  = nil // TODO: implement
		combat   internal.CombatManager   = combat.NewCombatManager(world)
		economy  internal.EconomyManager  = economy.NewEconomyManager()
		research internal.ResearchManager = research.NewResearchManager(eventBus)
		storage  internal.Storage         = nil
	)

	// Create network transport (SSH server)
	networkTransport, err := network.NetworkTransport()
	if err != nil {
		log.Error("Failed to create network transport", "error", err)
		os.Exit(1)
	}

	// Create UI renderer and viewport
	viewport := ui.NewViewport(cfg.Game.WorldWidth, cfg.Game.WorldHeight, 80, 24)
	renderer := ui.NewANSIRenderer(viewport, 80)

	// Create game engine
	game := engine.NewGame(
		world,
		building,
		citizen,
		combat,
		economy,
		research,
		eventBus,
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
	if err := networkTransport.Start(); err != nil {
		log.Error("Failed to start network transport", "error", err)
		// Continue without network? For now, just log.
	} else {
		log.Info("SSH server started", "port", 2222)
	}

	// Start a goroutine for processing network commands
	go func() {
		for msg := range networkTransport.Receive() {
			// Парсим команду
			cmd, err := network.ParseCommand(msg.Data)
			if err != nil {
				log.Warn("Failed to parse command", "error", err)
				continue
			}
			// Отправляем в движок
			if err := game.SubmitCommand(msg.ClientID, cmd); err != nil {
				log.Warn("Failed to submit command", "error", err)
			}
		}
	}()

	// Start a goroutine for rendering and sending frames
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // 10 FPS
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Для каждого игрока рендерим кадр и отправляем
				// TODO: получить список игроков из game
				// Временно просто отправляем тестовый кадр
				frame, err := renderer.Render("test", internal.Viewport{})
				if err != nil {
					log.Warn("Failed to render frame", "error", err)
					continue
				}
				// Отправляем всем клиентам
				networkTransport.Broadcast(frame)
			}
		}
	}()

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
	if networkTransport != nil {
		if err := networkTransport.Stop(); err != nil {
			log.Error("Error stopping network transport", "error", err)
		}
	}

	log.Info("Server stopped gracefully")
}
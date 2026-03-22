package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ssh-arena-app/internal"
	"ssh-arena-app/internal/building"
	"ssh-arena-app/internal/citizen"
	"ssh-arena-app/internal/combat"
	"ssh-arena-app/internal/economy"
	"ssh-arena-app/internal/engine"
	"ssh-arena-app/internal/events"
	"ssh-arena-app/internal/network"
	"ssh-arena-app/internal/research"
	"ssh-arena-app/internal/storage"
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

	// Create storage
	storage, err := storage.NewStorage(&cfg.Database)
	if err != nil {
		log.Error("Failed to create storage", "error", err)
		os.Exit(1)
	}

	// Load saved game state if any
	var loadedState *internal.GameState
	if storage != nil {
		loadedState, err = storage.LoadGame("default")
		if err != nil {
			log.Warn("Failed to load game state", "error", err)
		} else if loadedState != nil {
			log.Info("Game state loaded from storage")
			// Restore world and players (simplified)
			// In a full implementation, we would replace world and players.
			// For now, we just log.
		}
	}

	// Create managers
	building := building.NewManager()
	citizen := citizen.NewManager()
	combat := combat.NewCombatManager(world)
	economy := economy.NewEconomyManager()
	research := research.NewResearchManager(eventBus)

	// Create QuestManager with a reward handler that uses EconomyManager
	rewardHandler := func(playerID string, resources map[string]int) error {
		// TODO: integrate with EconomyManager to add resources
		log.Info("Quest reward", "player", playerID, "resources", resources)
		return nil
	}
	questManager := events.NewQuestManager(eventBus, rewardHandler)
	// Register some example quests (optional)
	// questManager.RegisterQuest(...)
	_ = questManager // suppress unused variable warning

	// Create UI renderer and viewport with larger size for 1600x900 terminal
	mapWidth := 150
	mapHeight := 35
	viewport := ui.NewViewport(cfg.Game.WorldWidth, cfg.Game.WorldHeight, mapWidth, mapHeight)
	renderer := ui.NewANSIRenderer(viewport, mapWidth)
	// Enable frame with title
	renderer.EnableFrame("SSH Arena")
	// Pass managers to renderer so it can fetch live data
	renderer.SetManagers(world, building, citizen, combat, economy, research)

	// Create game engine
	saveInterval := time.Duration(cfg.Server.SaveInterval) * time.Second
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
		saveInterval,
	)

	// Create network transport (SSH server) with callback to add players
	networkTransport, err := network.NewSSHTransport(func(playerID string) {
		if err := game.AddPlayer(playerID); err != nil {
			log.Warn("Failed to add player", "player", playerID, "error", err)
		} else {
			log.Info("Player added via SSH connection", "player", playerID)
		}
	})
	if err != nil {
		log.Error("Failed to create network transport", "error", err)
		os.Exit(1)
	}

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
				// Получаем состояние игры, чтобы получить список игроков
				state, err := game.GetState("") // GetState требует playerID, но мы можем использовать пустой или первого игрока
				if err != nil {
					// Если нет игроков, отправляем тестовый кадр
					frame, err := renderer.Render("test", internal.Viewport{})
					if err != nil {
						log.Warn("Failed to render frame", "error", err)
						continue
					}
					networkTransport.Broadcast(frame)
					continue
				}
				// Для каждого игрока рендерим кадр
				for playerID, player := range state.Players {
					// Центрируем вьюпорт на игроке
					viewport.CenterOn(player.Location.X, player.Location.Y)
					// Рендерим
					frame, err := renderer.Render(playerID, internal.Viewport{
						X:      viewport.X,
						Y:      viewport.Y,
						Width:  viewport.Width,
						Height: viewport.Height,
					})
					if err != nil {
						log.Warn("Failed to render frame for player", "player", playerID, "error", err)
						continue
					}
					// Отправляем только этому игроку
					if err := networkTransport.Send(playerID, frame); err != nil {
						log.Warn("Failed to send frame to player", "player", playerID, "error", err)
					}
				}
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

	// Close storage
	if storage != nil {
		if err := storage.Close(); err != nil {
			log.Error("Error closing storage", "error", err)
		}
	}

	log.Info("Server stopped gracefully")
}
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// AutoObserver is the main application controller
type AutoObserver struct {
	gsiServer  *GSIServer
	analyzer   *PlayerAnalyzer
	controller *SpectatorController
	running    bool
}

// NewAutoObserver creates a new AutoObserver instance
func NewAutoObserver() *AutoObserver {
	return &AutoObserver{
		gsiServer:  NewGSIServer("8000"),
		analyzer:   NewPlayerAnalyzer(),
		controller: NewSpectatorController(),
		running:    false,
	}
}

// Start starts the Auto-Observer system
func (ao *AutoObserver) Start() {
	LogInfo("Starting Better Auto Observer...")

	// Start GSI Server in goroutine
	go func() {
		if err := ao.gsiServer.Start(); err != nil {
			log.Fatalf("Failed to start GSI server: %v", err)
		}
	}()

	LogInfo("GSI Server started on port 8000")
	time.Sleep(1 * time.Second)

	// Start main loop
	ao.running = true
	ao.mainLoop()
}

// Stop stops the Auto-Observer system
func (ao *AutoObserver) Stop() {
	LogInfo("Stopping Auto Observer...")
	ao.running = false
}

// mainLoop is the main program loop
func (ao *AutoObserver) mainLoop() {
	LogInfo("Auto Observer running. Press Ctrl+C to stop.")
	LogInfo("Make sure you are in spectator mode in CS:GO/CS2!")

	updateInterval := 500 * time.Millisecond
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	lastRoundPhase := ""

	// Signal handler for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for ao.running {
		select {
		case <-sigChan:
			LogInfo("Received interrupt signal, shutting down...")
			ao.Stop()
			return

		case <-ticker.C:
			gameState := ao.gsiServer.GetCurrentGameState()

			// Wait for data
			if len(gameState) == 0 {
				LogDebug("No game state data received yet. Make sure CS is running and GSI config is installed!")
				time.Sleep(5 * time.Second)
				continue
			}

			// Check round phase
			var roundPhase string
			if round, ok := gameState["round"].(map[string]interface{}); ok {
				if phase, ok := round["phase"].(string); ok {
					roundPhase = phase
				}
			}

			// On phase change: update player slots
			if roundPhase != lastRoundPhase && roundPhase != "" {
				LogVerbose("[MAIN] Round phase changed to: %s", roundPhase)
				ao.controller.UpdatePlayerSlots(gameState)
				lastRoundPhase = roundPhase
			}

			// Only switch automatically during "live" or "freezetime"
			if roundPhase == "live" || roundPhase == "freezetime" {
				LogVerbose("[MAIN] Analyzing game state (phase: %s)...", roundPhase)
				bestSteamID := ao.analyzer.GetBestPlayerToSpectate(gameState)

				if bestSteamID != "" {
					ao.controller.SwitchToPlayer(bestSteamID)
				} else {
					LogVerbose("[MAIN] No player switch needed (rate limiting or no encounters)")
				}
			} else {
				LogVerbose("[MAIN] Waiting for live/freezetime phase (current: %s)", roundPhase)
			}
		}
	}
}

func main() {
	// Parse command line flags
	verbose := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	// Initialize logging
	if err := InitLogging(*verbose); err != nil {
		fmt.Printf("Warning: Could not initialize logging: %v\n", err)
	}
	defer CloseLogging()

	// Banner
	fmt.Println(`
╔══════════════════════════════════════════════════════════════╗
║        Better Auto Observer for Counter-Strike              ║
║                                                              ║
║  Intelligent automatic spectating                           ║
║  Automatically switches to exciting player encounters       ║
╚══════════════════════════════════════════════════════════════╝

SETUP:
1. Copy 'gamestate_integration_autoobserver.cfg' to:
   CS:GO/CS2: Steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/

2. Start CS:GO/CS2

3. Enter spectator mode (watch a game)

4. This program will automatically switch between players

IMPORTANT:
- You must be in spectator mode!
- The game must be in the foreground for keyboard inputs to work
- Press Ctrl+C to exit

`)

	observer := NewAutoObserver()
	observer.Start()
}

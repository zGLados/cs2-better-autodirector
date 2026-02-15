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

// AutoDirector is the main application controller
type AutoDirector struct {
	gsiServer  *GSIServer
	analyzer   *PlayerAnalyzer
	controller *SpectatorController
	running    bool
}

// NewAutoDirector creates a new AutoDirector instance
func NewAutoDirector() *AutoDirector {
	return &AutoDirector{
		gsiServer:  NewGSIServer("8000"),
		analyzer:   NewPlayerAnalyzer(),
		controller: NewSpectatorController(),
		running:    false,
	}
}

// Start starts the Auto-Director system
func (ao *AutoDirector) Start() {
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

// Stop stops the Auto-Director system
func (ao *AutoDirector) Stop() {
	LogInfo("Stopping Auto Observer...")
	ao.running = false
}

// mainLoop is the main program loop
func (ao *AutoDirector) mainLoop() {
	LogInfo("Auto Observer running. Press Ctrl+C to stop.")
	LogInfo("Make sure you are in spectator mode in CS2!")

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

			// Only switch automatically during gameplay phases (not gameover, paused, etc.)
			// Allow: live, freezetime, warmup, timeout
			if roundPhase == "live" || roundPhase == "freezetime" || roundPhase == "warmup" || roundPhase == "timeout" {
				LogVerbose("[MAIN] Analyzing game state (phase: %s)...", roundPhase)

				// IMPORTANT: Sync our internal state with actual spectated player from GSI
				// This prevents us from thinking we're on one player when CS actually switched to another
				actualSpectatedID := ao.gsiServer.GetCurrentlySpectatedPlayer()
				if actualSpectatedID != "" && actualSpectatedID != ao.analyzer.currentSpectatedID {
					LogInfo("🔄 Synced spectated player: CS is on different player than expected")
					LogVerbose("[MAIN] Expected: %s, Actual: %s", ao.analyzer.currentSpectatedID, actualSpectatedID)
					ao.analyzer.SyncCurrentPlayer(actualSpectatedID)
				}

				bestSteamID := ao.analyzer.GetBestPlayerToSpectate(gameState, roundPhase)

				if bestSteamID != "" {
					ao.controller.SwitchToPlayer(bestSteamID)
				} else {
					LogVerbose("[MAIN] No player switch needed (rate limiting or no encounters)")
				}
			} else {
				LogVerbose("[MAIN] Waiting for gameplay phase (current: %s)", roundPhase)
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
║        CS2 Better Auto Director                              ║
║                                                              ║
║  Intelligent automatic spectating                           ║
║  Automatically switches to exciting player encounters       ║
╚══════════════════════════════════════════════════════════════╝

SETUP:
1. Copy config files to CS2 folder:
   - gamestate_integration_autodirector.cfg
   - spectator_bindings.cfg
   Location: Steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/cfg/

2. In CS2 console, run once per session:
   exec spectator_bindings
   bind F9 spec_mode_toggle

3. Join spectator mode (GOTV/Demo), press F9 to enable spectator bindings

4. This program will automatically switch between players

IMPORTANT:
- Press F9 to toggle: Spectator mode (1-0 = players) ↔ Normal mode (1-0 = weapons)
- Test manually: Press F9, then keys 1-9 - camera should switch between players
- CS2 window must be in foreground for keyboard inputs to work
- Run as Administrator if switches don't work
- Press Ctrl+C to exit

TIP: Use -v flag for verbose logging (cs2-better-autodirector.exe -v)

`)

	director := NewAutoDirector()
	director.Start()
}

package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Parse command line flags
	nogui := flag.Bool("nogui", false, "Run in CLI mode without GUI")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	if *nogui {
		// Run in CLI mode (old behavior)
		runCLI(*verbose)
	} else {
		// Run in GUI mode (new Wails UI)
		runGUI()
	}
}

// runGUI starts the application with Wails GUI
func runGUI() {
	// Initialize logging for GUI mode (non-verbose by default)
	if err := InitLogging(false); err != nil {
		fmt.Printf("Warning: Could not initialize logging: %v\n", err)
	}
	defer CloseLogging()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "CS2 Better Auto Director",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fmt.Printf("\n=== ERROR ===\n")
		fmt.Printf("Failed to start GUI: %v\n", err)
		fmt.Printf("=============\n\n")
		fmt.Printf("Possible solutions:\n")
		fmt.Printf("1. Make sure WebView2 Runtime is installed\n")
		fmt.Printf("   Download from: https://go.microsoft.com/fwlink/p/?LinkId=2124703\n")
		fmt.Printf("2. Try running as Administrator\n")
		fmt.Printf("3. Use CLI mode instead: cs2-better-autodirector.exe -nogui\n\n")
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}

// runCLI starts the application in CLI mode
func runCLI(verbose bool) {
	// Initialize logging
	if err := InitLogging(verbose); err != nil {
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

MODE: CLI (No GUI)
TIP: Run without -nogui flag to use graphical interface

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

TIP: Use -v flag for verbose logging

`)

	director := NewCLIAutoDirector()
	director.Start()
}

// CLIAutoDirector is the CLI version of the auto director
type CLIAutoDirector struct {
	gsiServer  *GSIServer
	analyzer   *PlayerAnalyzer
	controller *SpectatorController
	running    bool
}

// NewCLIAutoDirector creates a new CLI AutoDirector instance
func NewCLIAutoDirector() *CLIAutoDirector {
	return &CLIAutoDirector{
		gsiServer:  NewGSIServer("3000"), // Port 3000 (same as GUI)
		analyzer:   NewPlayerAnalyzer(),
		controller: NewSpectatorController(),
		running:    false,
	}
}

// Start starts the CLI Auto-Director system
func (ao *CLIAutoDirector) Start() {
	LogInfo("Starting CS2 Auto Director (CLI Mode)...")

	// Start GSI Server in goroutine
	go func() {
		if err := ao.gsiServer.Start(); err != nil {
			log.Fatalf("Failed to start GSI server: %v", err)
		}
	}()

	LogInfo("GSI Server started on port 3000")
	time.Sleep(1 * time.Second)

	// Start main loop
	ao.running = true
	ao.mainLoop()
}

// Stop stops the CLI Auto-Director system
func (ao *CLIAutoDirector) Stop() {
	LogInfo("Stopping Auto Director...")
	ao.running = false
}

// mainLoop is the main program loop for CLI mode
func (ao *CLIAutoDirector) mainLoop() {
	LogInfo("Auto Director running. Press Ctrl+C to stop.")
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
				LogVerbose("[CLI] Round phase changed to: %s", roundPhase)
				ao.controller.UpdatePlayerSlots(gameState)
				lastRoundPhase = roundPhase
			}

			// Only switch automatically during gameplay phases
			if roundPhase == "live" || roundPhase == "freezetime" || roundPhase == "warmup" || roundPhase == "timeout" {
				LogVerbose("[CLI] Analyzing game state (phase: %s)...", roundPhase)

				// Sync with actual spectated player from GSI
				actualSpectatedID := ao.gsiServer.GetCurrentlySpectatedPlayer()
				if actualSpectatedID != "" && actualSpectatedID != ao.analyzer.currentSpectatedID {
					LogInfo("🔄 Synced spectated player: CS is on different player than expected")
					ao.analyzer.SyncCurrentPlayer(actualSpectatedID)
				}

				bestSteamID := ao.analyzer.GetBestPlayerToSpectate(gameState, roundPhase)

				if bestSteamID != "" {
					ao.controller.SwitchToPlayer(bestSteamID)
				} else {
					LogVerbose("[CLI] No player switch needed (rate limiting or no encounters)")
				}
			} else {
				LogVerbose("[CLI] Waiting for gameplay phase (current: %s)", roundPhase)
			}
		}
	}

	LogInfo("Auto Director stopped")
}

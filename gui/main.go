package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Global panic recovery to show errors even if the GUI doesn't load
	defer func() {
		if r := recover(); r != nil {
			errTitle := "Critical Error - CS2 Better Auto Director"
			errMessage := fmt.Sprintf("The application crashed:\n\n%v\n\nStack trace:\n%s", r, debug.Stack())

			// Print to console/file as well
			fmt.Fprintf(os.Stderr, "%s: %s\n", errTitle, errMessage)

			// On Windows, show a native message box so the user sees the error
			if os.Getenv("OS") == "Windows_NT" {
				showWindowsMessageBox(errTitle, errMessage)
			}
		}
	}()

	// Parse command line flags
	nogui := flag.Bool("nogui", false, "Run in CLI mode without GUI")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	veryVerbose := flag.Bool("vv", false, "Enable very verbose (trace) logging")
	flag.Parse()

	logLevel := 0
	if *veryVerbose {
		logLevel = 2
	} else if *verbose {
		logLevel = 1
	}

	// If verbose mode is enabled on Windows, we need to attach or create a console
	if logLevel > 0 && os.Getenv("OS") == "Windows_NT" {
		setupWindowsConsole()
	}

	if *nogui {
		// Run in CLI mode (old behavior)
		runCLI(logLevel)
	} else {
		// Run in GUI mode (new Wails UI)
		runGUI(logLevel)
	}
}

// runGUI starts the application with Wails GUI
func runGUI(level int) {
	// Initialize logging with current verbose flag
	if err := InitLogging(level); err != nil {
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
		errMsg := fmt.Sprintf("Failed to start GUI: %v\n\nPossible solutions:\n1. Install WebView2 Runtime\n2. Run as Administrator\n3. Check logs in the app folder or AppData", err)
		if os.Getenv("OS") == "Windows_NT" {
			showWindowsMessageBox("Startup Error", errMsg)
		} else {
			fmt.Println(errMsg)
		}
	}
}

// setupWindowsConsole attaches the application to the parent console or creates a new one
func setupWindowsConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	attachConsole := kernel32.NewProc("AttachConsole")

	// Try to attach to the console of the parent process (e.g., CMD or PowerShell)
	// 0xFFFFFFFF is ATTACH_PARENT_PROCESS
	r, _, _ := attachConsole.Call(uintptr(0xFFFFFFFF))
	if r == 0 {
		// If attaching fails, create a new console window
		allocConsole := kernel32.NewProc("AllocConsole")
		allocConsole.Call()
	}

	// Redirect standard handles to the console
	hout, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	os.Stdout = os.NewFile(uintptr(hout), "/dev/stdout")
	os.Stderr = os.NewFile(uintptr(os.Stderr.Fd()), "/dev/stderr")
	log.SetOutput(os.Stdout)
}

// showWindowsMessageBox shows a native Windows error dialog without requiring any GUI toolkit
func showWindowsMessageBox(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	lpText, _ := syscall.UTF16PtrFromString(message)
	lpCaption, _ := syscall.UTF16PtrFromString(title)

	// MB_OK | MB_ICONERROR = 0x00000000 | 0x00000010
	messageBox.Call(0, uintptr(unsafe.Pointer(lpText)), uintptr(unsafe.Pointer(lpCaption)), 0x00000010)
}

// runCLI starts the application in CLI mode
func runCLI(level int) {
	// Initialize logging
	if err := InitLogging(level); err != nil {
		fmt.Printf("Warning: Could not initialize logging: %v\n", err)
	}
	defer CloseLogging()

	// Banner
	fmt.Println(`
╔══════════════════════════════════════════════════════════════╗
║        CS2 Better Auto Director                              ║
║                                                              ║
║  Intelligent automatic spectating                            ║
║  Automatically switches to exciting player encounters        ║
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

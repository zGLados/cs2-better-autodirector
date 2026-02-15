package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	autoDirector   *AutoDirector
	mu             sync.RWMutex
	isRunning      bool
	stats          *Statistics
	lastEncounters []EncounterInfo
	lastPlayers    []PlayerInfo
	lastGameState  map[string]interface{}
}

// Statistics holds runtime statistics
type Statistics struct {
	TotalSwitches    int
	SwitchesPerMin   float64
	Uptime           time.Duration
	StartTime        time.Time
	LastSwitchTime   time.Time
	SniperKills      int
	UpsetVictories   int
	DamageDetections int
}

// StatusInfo represents current system status
type StatusInfo struct {
	IsRunning         bool
	CurrentPlayer     string
	CurrentPlayerName string
	CurrentEncounter  string
	LastSwitch        string
	Uptime            string
	TotalSwitches     int
	RoundPhase        string
	AlivePlayers      int
	RoundTime         string
	RoundNumber       int
	ScoreCT           int
	ScoreT            int
}

// EncounterInfo represents an encounter for the frontend
type EncounterInfo struct {
	Player1         string
	Player2         string
	Player1Team     string
	Player2Team     string
	Distance        float64
	Priority        float64
	IsCurrentPlayer bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		stats: &Statistics{
			StartTime: time.Now(),
		},
		lastEncounters: make([]EncounterInfo, 0),
		lastPlayers:    make([]PlayerInfo, 0),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	LogInfo("GUI Started")
}

// StartAutoDirector starts the auto director system
func (a *App) StartAutoDirector() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return fmt.Errorf("AutoDirector is already running")
	}

	// Initialize AutoDirector
	a.autoDirector = &AutoDirector{
		gsiServer:           NewGSIServer("3000"),
		playerAnalyzer:      NewPlayerAnalyzer(),
		spectatorController: NewSpectatorController(),
		switchCheckInterval: 500 * time.Millisecond,
		app:                 a, // Set reference to App for callbacks
	}

	// Start GSI server in background
	go func() {
		if err := a.autoDirector.gsiServer.Start(); err != nil {
			LogInfo("Failed to start GSI server: %v", err)
			runtime.EventsEmit(a.ctx, "error", fmt.Sprintf("GSI Server Error: %v", err))
		}
	}()

	// Start main loop in background
	go a.autoDirector.Run()

	a.isRunning = true
	a.stats.StartTime = time.Now()

	LogInfo("AutoDirector started")
	runtime.EventsEmit(a.ctx, "status_changed", "running")

	return nil
}

// StopAutoDirector stops the auto director system
func (a *App) StopAutoDirector() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return fmt.Errorf("AutoDirector is not running")
	}

	// Stop the AutoDirector loop
	if a.autoDirector != nil && a.autoDirector.stopChan != nil {
		close(a.autoDirector.stopChan)
	}

	// Stop the GSI Server
	if a.autoDirector != nil && a.autoDirector.gsiServer != nil {
		if err := a.autoDirector.gsiServer.Stop(); err != nil {
			LogInfo("Error stopping GSI server: %v", err)
		}
	}

	// Wait a bit for goroutines to finish
	time.Sleep(500 * time.Millisecond)

	a.isRunning = false
	a.autoDirector = nil

	LogInfo("AutoDirector stopped")
	runtime.EventsEmit(a.ctx, "status_changed", "stopped")

	return nil
}

// GetStatus returns current system status
func (a *App) GetStatus() StatusInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := StatusInfo{
		IsRunning:     a.isRunning,
		TotalSwitches: a.stats.TotalSwitches,
		Uptime:        formatDuration(time.Since(a.stats.StartTime)),
	}

	if a.autoDirector != nil && a.autoDirector.playerAnalyzer != nil {
		status.CurrentPlayer = a.autoDirector.playerAnalyzer.currentSpectatedID
		// Get current player name from last players data
		for _, p := range a.lastPlayers {
			if p.SteamID == status.CurrentPlayer {
				status.CurrentPlayerName = p.Name
				break
			}
		}

		if !a.stats.LastSwitchTime.IsZero() {
			status.LastSwitch = time.Since(a.stats.LastSwitchTime).Round(time.Second).String() + " ago"
		}
	}

	status.AlivePlayers = len(a.lastPlayers)

	// Extract round information from last game state
	if a.lastGameState != nil {
		if round, ok := a.lastGameState["round"].(map[string]interface{}); ok {
			if phase, ok := round["phase"].(string); ok {
				status.RoundPhase = phase
			}
		}

		if mapData, ok := a.lastGameState["map"].(map[string]interface{}); ok {
			if roundNum, ok := mapData["round"].(float64); ok {
				status.RoundNumber = int(roundNum)
			}
			if teamCT, ok := mapData["team_ct"].(map[string]interface{}); ok {
				if score, ok := teamCT["score"].(float64); ok {
					status.ScoreCT = int(score)
				}
			}
			if teamT, ok := mapData["team_t"].(map[string]interface{}); ok {
				if score, ok := teamT["score"].(float64); ok {
					status.ScoreT = int(score)
				}
			}
		}

		// Get round time from phase_ends_in
		if round, ok := a.lastGameState["round"].(map[string]interface{}); ok {
			if phaseEndsIn, ok := round["phase_ends_in"].(string); ok {
				status.RoundTime = phaseEndsIn
			}
		}
	}

	return status
}

// GetPlayers returns current player list
func (a *App) GetPlayers() []PlayerInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastPlayers
}

// GetEncounters returns top encounters
func (a *App) GetEncounters() []EncounterInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastEncounters
}

// GetStatistics returns runtime statistics
func (a *App) GetStatistics() Statistics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := *a.stats
	stats.Uptime = time.Since(a.stats.StartTime)

	// Calculate switches per minute
	if stats.Uptime.Minutes() > 0 {
		stats.SwitchesPerMin = float64(stats.TotalSwitches) / stats.Uptime.Minutes()
	}

	return stats
}

// UpdatePlayers is called internally to update player data and emit events
func (a *App) UpdatePlayers(players []PlayerInfo) {
	a.mu.Lock()
	a.lastPlayers = players
	a.mu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "players_updated", players)
	}
}

// UpdateEncounters is called internally to update encounter data and emit events
func (a *App) UpdateEncounters(encounters []EncounterInfo) {
	a.mu.Lock()
	a.lastEncounters = encounters
	a.mu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "encounters_updated", encounters)
	}
}

// RecordSwitch is called internally when a switch occurs
func (a *App) RecordSwitch() {
	a.mu.Lock()
	a.stats.TotalSwitches++
	a.stats.LastSwitchTime = time.Now()
	a.mu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "switch_occurred", a.GetStatus())
	}
}

// RecordSniperKill increments the sniper kill counter
func (a *App) RecordSniperKill() {
	a.mu.Lock()
	a.stats.SniperKills++
	a.mu.Unlock()
}

// RecordUpsetVictory increments the upset victory counter
func (a *App) RecordUpsetVictory() {
	a.mu.Lock()
	a.stats.UpsetVictories++
	a.mu.Unlock()
}

// RecordDamageDetection increments the damage detection counter
func (a *App) RecordDamageDetection() {
	a.mu.Lock()
	a.stats.DamageDetections++
	a.mu.Unlock()
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

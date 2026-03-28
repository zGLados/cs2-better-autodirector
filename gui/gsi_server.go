package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// GSIServer receives Game State Integration data from CS
type GSIServer struct {
	currentGameState    map[string]interface{}
	mu                  sync.RWMutex
	port                string
	debugDumped         bool
	lastDataReceivedLog time.Time // Track when we last logged "Data received"
	server              *http.Server
}

// NewGSIServer creates a new GSI Server
func NewGSIServer(port string) *GSIServer {
	return &GSIServer{
		currentGameState: make(map[string]interface{}),
		port:             port,
	}
}

// HandleGameState processes incoming GSI requests
func (s *GSIServer) HandleGameState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var gameState map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&gameState); err != nil {
		LogDebug("Error decoding gamestate: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save the current game state
	s.mu.Lock()
	s.currentGameState = gameState
	// Debug: dump raw GSI data (only first time to avoid spam)
	if LogLevel >= 1 && !s.debugDumped && len(gameState) > 0 {
		DumpGameState(gameState)
		s.debugDumped = true
	}
	s.mu.Unlock()

	// Debug logging
	if round, ok := gameState["round"].(map[string]interface{}); ok {
		if phase, ok := round["phase"].(string); ok {
			LogVerbose("[GSI] Round Phase: %s", phase)
		}
	}

	// Count players - but only log every 10 seconds to reduce spam
	if allPlayers, ok := gameState["allplayers"].(map[string]interface{}); ok {
		now := time.Now()
		if now.Sub(s.lastDataReceivedLog) >= 10*time.Second {
			LogInfo("Data received: %d players", len(allPlayers))
			s.lastDataReceivedLog = now
		}
	}

	w.WriteHeader(http.StatusOK)
}

// GetCurrentGameState returns the current game state
func (s *GSIServer) GetCurrentGameState() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Copy the map to avoid race conditions
	result := make(map[string]interface{})
	for k, v := range s.currentGameState {
		result[k] = v
	}
	return result
}

// GetCurrentlySpectatedPlayer tries to determine which player is currently being spectated
// Returns the SteamID of the spectated player, or empty string if unknown
func (s *GSIServer) GetCurrentlySpectatedPlayer() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// In spectator mode, GSI has a "player" field that represents the currently spectated player
	// This is the same field that would be the local player in normal gameplay
	if player, ok := s.currentGameState["player"].(map[string]interface{}); ok {
		// Try to get steamid from player data
		if steamidRaw, ok := player["steamid"]; ok {
			if steamid, ok := steamidRaw.(string); ok {
				LogVerbose("[GSI] Currently spectating player with SteamID: %s", steamid)
				return steamid
			}
		}
	}

	// Alternative: Check if there's observer data
	if observer, ok := s.currentGameState["observer"].(map[string]interface{}); ok {
		if steamidRaw, ok := observer["target"]; ok {
			if steamid, ok := steamidRaw.(string); ok {
				LogVerbose("[GSI] Observer target SteamID: %s", steamid)
				return steamid
			}
		}
	}

	return ""
}

// Start starts the GSI Server
func (s *GSIServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HandleGameState)

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	LogInfo("GSI Server starting on port %s", s.port)
	return s.server.ListenAndServe()
}

// Stop stops the GSI Server
func (s *GSIServer) Stop() error {
	if s.server == nil {
		return nil
	}

	LogInfo("Stopping GSI Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

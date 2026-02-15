package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

// GSIServer receives Game State Integration data from CS
type GSIServer struct {
	currentGameState map[string]interface{}
	mu               sync.RWMutex
	port             string
	debugDumped      bool
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
	if VerboseMode && !s.debugDumped && len(gameState) > 0 {
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

	// Count players
	if allPlayers, ok := gameState["allplayers"].(map[string]interface{}); ok {
		LogInfo("Data received: %d players", len(allPlayers))
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

// Start starts the GSI Server
func (s *GSIServer) Start() error {
	http.HandleFunc("/", s.HandleGameState)
	LogVerbose("GSI Server starting on port %s", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

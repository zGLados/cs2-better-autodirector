package main

import (
	"encoding/json"
)

// DumpGameState prints the raw GSI data for debugging
func DumpGameState(gameState map[string]interface{}) {
	if !VerboseMode {
		return
	}

	// Check if we have allplayers data
	allPlayers, ok := gameState["allplayers"].(map[string]interface{})
	if !ok {
		LogVerbose("[DEBUG] No 'allplayers' data in game state")
		return
	}

	LogVerbose("[DEBUG] ========== RAW GSI DATA SAMPLE ==========")

	// Show first player's full data structure
	count := 0
	for steamID, playerData := range allPlayers {
		if count > 0 {
			break // Only show first player
		}

		LogVerbose("[DEBUG] Sample player SteamID: %s", steamID)

		playerMap, ok := playerData.(map[string]interface{})
		if !ok {
			LogVerbose("[DEBUG] Could not parse player data")
			continue
		}

		// Show available keys
		LogVerbose("[DEBUG] Available keys in player data:")
		for key := range playerMap {
			LogVerbose("[DEBUG]   - %s", key)
		}

		// Check for observer_slot
		if observerSlot, ok := playerMap["observer_slot"]; ok {
			LogVerbose("[DEBUG] ✓ observer_slot found: %v", observerSlot)
		} else {
			LogVerbose("[DEBUG] ⚠️  NO 'observer_slot' field found")
		}

		// Check for position data
		if position, ok := playerMap["position"].(map[string]interface{}); ok {
			LogVerbose("[DEBUG] Position data found:")
			for key, value := range position {
				LogVerbose("[DEBUG]   position.%s = %v", key, value)
			}
		} else {
			LogVerbose("[DEBUG] ❌ NO 'position' KEY FOUND in player data!")
			LogVerbose("[DEBUG] This means CS is not sending position data.")
			LogVerbose("[DEBUG] Make sure you are in GOTV/Demo playback mode, not free camera!")
		}

		// Pretty print the entire player data
		jsonBytes, err := json.MarshalIndent(playerMap, "", "  ")
		if err == nil {
			LogVerbose("[DEBUG] Full player data:\n%s", string(jsonBytes))
		}

		count++
	}

	LogVerbose("[DEBUG] ==========================================")
}

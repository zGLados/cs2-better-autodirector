package main

import (
	"sort"
	"time"

	"github.com/go-vgo/robotgo"
)

// SpectatorController controls the observer mode in CS
type SpectatorController struct {
	steamIDToSlot  map[string]int
	currentPlayers []PlayerInfo
}

// NewSpectatorController creates a new Spectator Controller
func NewSpectatorController() *SpectatorController {
	return &SpectatorController{
		steamIDToSlot:  make(map[string]int),
		currentPlayers: make([]PlayerInfo, 0),
	}
}

// UpdatePlayerSlots updates the mapping from SteamID to player slot
func (sc *SpectatorController) UpdatePlayerSlots(gameState map[string]interface{}) {
	allPlayers, ok := gameState["allplayers"].(map[string]interface{})
	if !ok {
		return
	}

	sc.steamIDToSlot = make(map[string]int)
	sc.currentPlayers = make([]PlayerInfo, 0)

	LogVerbose("[CONTROLLER] Checking for observer_slot in GSI data...")

	// First, try to get observer_slot from GSI data
	playersWithSlots := make(map[int]PlayerInfo)
	playersWithoutSlots := make([]PlayerInfo, 0)
	hasObserverSlot := false

	for steamID, playerData := range allPlayers {
		playerMap, ok := playerData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract basic player info
		name := getStringValue(playerMap, "name")
		if name == "" || name == "Unknown" {
			continue
		}

		team := getStringValue(playerMap, "team")
		state, _ := playerMap["state"].(map[string]interface{})
		health := getIntValue(state, "health")

		if health <= 0 {
			continue
		}

		matchStats, _ := playerMap["match_stats"].(map[string]interface{})
		kills := getIntValue(matchStats, "kills")

		weapons, _ := playerMap["weapons"].(map[string]interface{})
		equipValue := 0
		for _, weapon := range weapons {
			weaponMap, _ := weapon.(map[string]interface{})
			weaponType := getStringValue(weaponMap, "type")
			if weaponType != "Knife" && weaponType != "C4" && weaponType != "Grenade" {
				equipValue += getIntValue(weaponMap, "value")
			}
		}

		player := PlayerInfo{
			SteamID:        steamID,
			Name:           name,
			Team:           team,
			Health:         health,
			Kills:          kills,
			EquipmentValue: equipValue,
		}

		// Check if GSI provides observer_slot
		if observerSlot, ok := playerMap["observer_slot"]; ok {
			hasObserverSlot = true
			slot := int(observerSlot.(float64)) + 1 // GSI uses 0-based index
			playersWithSlots[slot] = player
			LogVerbose("[CONTROLLER] Player %s has observer_slot: %d", name, slot)
		} else {
			playersWithoutSlots = append(playersWithoutSlots, player)
		}
	}

	if hasObserverSlot {
		// Use observer_slot from GSI
		LogVerbose("[CONTROLLER] ✓ Using observer_slot from GSI data")
		LogVerbose("[CONTROLLER] Updating player slots for %d players:", len(playersWithSlots))

		for slot := 1; slot <= 10; slot++ {
			if player, exists := playersWithSlots[slot]; exists {
				sc.steamIDToSlot[player.SteamID] = slot
				sc.currentPlayers = append(sc.currentPlayers, player)
				LogVerbose("[CONTROLLER]   Slot %d: %s (%s) - HP:%d Eq:$%d K:%d",
					slot, player.Name, player.Team, player.Health, player.EquipmentValue, player.Kills)
			}
		}
	} else {
		// Fallback: Use custom sorting (old method)
		LogVerbose("[CONTROLLER] ⚠️  No observer_slot in GSI - using fallback sorting")

		players := playersWithoutSlots
		// Sort: CT first, then T, then alphabetically
		sort.Slice(players, func(i, j int) bool {
			if players[i].Team != players[j].Team {
				return players[i].Team == "CT"
			}
			return players[i].Name < players[j].Name
		})

		sc.currentPlayers = players
		LogVerbose("[CONTROLLER] Updating player slots for %d players:", len(players))

		for idx, player := range players {
			if idx >= 10 {
				break
			}
			slot := idx + 1
			sc.steamIDToSlot[player.SteamID] = slot
			LogVerbose("[CONTROLLER]   Slot %d: %s (%s) - HP:%d Eq:$%d K:%d",
				slot, player.Name, player.Team, player.Health, player.EquipmentValue, player.Kills)
		}
	}
}

// SwitchToPlayer switches to a player based on their SteamID
func (sc *SpectatorController) SwitchToPlayer(steamID string) bool {
	slot, ok := sc.steamIDToSlot[steamID]
	if !ok {
		LogVerbose("[CONTROLLER] ❌ SteamID %s not found in player slots", steamID)
		return false
	}

	// Find player name for logging
	playerName := "Unknown"
	for _, player := range sc.currentPlayers {
		if player.SteamID == steamID {
			playerName = player.Name
			break
		}
	}

	LogInfo("➡️  Switching to: %s (Slot %d)", playerName, slot)

	// Press the corresponding number key
	var key string
	if slot == 10 {
		key = "0"
	} else {
		key = string(rune('0' + slot))
	}

	// Use more reliable key press method
	// Hold key down briefly instead of just tapping
	LogVerbose("[CONTROLLER] Pressing key '%s' (hold method)...", key)

	// Method: KeyToggle with multiple repetitions for maximum reliability
	// This helps overcome focus issues and missed key presses

	// Repeat 3 times with delays (increased from 2x for better reliability)
	for i := 0; i < 3; i++ {
		robotgo.KeyToggle(key, "down")
		time.Sleep(150 * time.Millisecond) // Increased from 100ms
		robotgo.KeyToggle(key, "up")
		time.Sleep(150 * time.Millisecond) // Increased from 100ms

		if i < 2 {
			time.Sleep(50 * time.Millisecond) // Brief pause between repetitions
		}
	}

	time.Sleep(250 * time.Millisecond) // Final pause before next operation

	LogVerbose("[CONTROLLER] ✓ Key sequence completed (3x repetition)")

	return true
}

// EnableXRay activates X-Ray (wallhack in observer mode)
func (sc *SpectatorController) EnableXRay() {
	robotgo.KeyTap("x")
	time.Sleep(50 * time.Millisecond)
	LogVerbose("X-Ray toggled")
}

// CycleCameraMode switches the camera mode
func (sc *SpectatorController) CycleCameraMode() {
	robotgo.KeyTap("space")
	time.Sleep(50 * time.Millisecond)
	LogVerbose("Camera mode cycled")
}

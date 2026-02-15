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
	players := GetPlayerInfo(gameState)

	// Sort: CT first, then T, then alphabetically
	sort.Slice(players, func(i, j int) bool {
		if players[i].Team != players[j].Team {
			return players[i].Team == "CT"
		}
		return players[i].Name < players[j].Name
	})

	// Assign slots 1-10
	sc.steamIDToSlot = make(map[string]int)
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
	LogInfo("[CONTROLLER] ⚠️  If switch didn't work: Make sure CS2 window is in FOCUS!")

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

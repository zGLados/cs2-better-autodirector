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
		LogVerbose("SteamID %s not found in player slots", steamID)
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

	// Press key with robotgo
	robotgo.KeyTap(key)
	time.Sleep(50 * time.Millisecond)

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

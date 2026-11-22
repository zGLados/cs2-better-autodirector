package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-vgo/robotgo"
)

// SpectatorController controls the observer mode in CS
type SpectatorController struct {
	steamIDToSlot  map[string]int
	currentPlayers []PlayerInfo
	netcon         *NetconClient
	// GetSpectated returns the SteamID CS2 is currently spectating (from GSI).
	// Used to confirm key-press switches. Optional.
	GetSpectated func() string
}

// NewSpectatorController creates a new Spectator Controller
func NewSpectatorController() *SpectatorController {
	return &SpectatorController{
		steamIDToSlot:  make(map[string]int),
		currentPlayers: make([]PlayerInfo, 0),
		netcon:         NewNetconClient(NetconPort),
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
		activeWeapon := ""
		for _, weapon := range weapons {
			weaponMap, _ := weapon.(map[string]interface{})
			weaponType := getStringValue(weaponMap, "type")
			weaponName := getStringValue(weaponMap, "name")
			weaponState := getStringValue(weaponMap, "state")

			// Track active weapon
			if weaponState == "active" {
				activeWeapon = weaponName
			}

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
			ActiveWeapon:   activeWeapon,
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

	// Preferred path: send the same command the key binding runs
	// (spectator_bindings.cfg: key N -> spec_player_by_index N-1) via netcon.
	// Takes milliseconds instead of ~1.1s of key presses and needs no focus.
	if err := sc.netcon.Send(fmt.Sprintf("spec_player_by_index %d", slot-1)); err == nil {
		LogVerbose("[CONTROLLER] ✓ Switched via netcon (spec_player_by_index %d)", slot-1)
		return true
	}

	// Press the corresponding number key (spectator_bindings.cfg)
	var key string
	if slot == 10 {
		key = "0"
	} else {
		key = string(rune('0' + slot))
	}

	// FIX: the old method always pressed the key 3 times with 150ms holds and a
	// final pause (~1.15s per switch, blocking the main loop), so the camera
	// often arrived after the kill. Now we press once with a short hold and
	// confirm via GSI that CS2 is actually on the target player; only if it
	// isn't do we press again (max 3 attempts, same worst case as before).
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		robotgo.KeyToggle(key, "down")
		time.Sleep(40 * time.Millisecond)
		robotgo.KeyToggle(key, "up")

		if sc.GetSpectated == nil {
			// No way to confirm: keep a second press for reliability
			if attempt == 2 {
				break
			}
			time.Sleep(60 * time.Millisecond)
			continue
		}

		// Wait for GSI to report the new player (GSI throttle+buffer is ~0.2s)
		deadline := time.Now().Add(350 * time.Millisecond)
		reported := ""
		for time.Now().Before(deadline) {
			reported = sc.GetSpectated()
			if reported == steamID {
				LogVerbose("[CONTROLLER] ✓ Key '%s' confirmed by GSI (attempt %d)", key, attempt)
				return true
			}
			time.Sleep(25 * time.Millisecond)
		}
		if reported == "" {
			// GSI doesn't report the spectated player: can't confirm, stop here
			LogVerbose("[CONTROLLER] GSI does not report the spectated player - cannot confirm switch")
			return true
		}
		LogVerbose("[CONTROLLER] Key '%s' not confirmed yet (attempt %d/%d)", key, attempt, maxAttempts)
	}

	if sc.GetSpectated != nil {
		LogInfo("⚠️  Switch to %s not confirmed by GSI - is CS2 in the foreground and spectator_bindings active (F9)?", playerName)
	}
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

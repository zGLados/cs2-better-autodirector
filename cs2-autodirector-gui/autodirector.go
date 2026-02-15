package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// AutoDirector is the main application controller for GUI mode
type AutoDirector struct {
	gsiServer           *GSIServer
	playerAnalyzer      *PlayerAnalyzer
	spectatorController *SpectatorController
	running             bool
	switchCheckInterval time.Duration
	app                 *App // Reference back to App for events
	stopChan            chan bool
}

// Run starts the main AutoDirector loop
func (ad *AutoDirector) Run() {
	LogInfo("AutoDirector main loop starting...")
	LogInfo("Make sure you are in spectator mode in CS2!")

	ad.stopChan = make(chan bool)
	ticker := time.NewTicker(ad.switchCheckInterval)
	defer ticker.Stop()

	lastRoundPhase := ""

	// Signal handler for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ad.running = true

	for ad.running {
		select {
		case <-ad.stopChan:
			LogInfo("Received stop signal...")
			ad.running = false
			return

		case <-sigChan:
			LogInfo("Received interrupt signal, shutting down...")
			ad.running = false
			return

		case <-ticker.C:
			gameState := ad.gsiServer.GetCurrentGameState()

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
				LogVerbose("[AUTODIRECTOR] Round phase changed to: %s", roundPhase)
				ad.spectatorController.UpdatePlayerSlots(gameState)
				lastRoundPhase = roundPhase
			}

			// Extract players for GUI updates
			players := extractPlayers(gameState)
			if ad.app != nil {
				ad.app.UpdatePlayers(players)
			}

			// Extract and send encounters for GUI
			encounters := ad.playerAnalyzer.PredictEncounters(players)
			if ad.app != nil && len(encounters) > 0 {
				encounterInfos := convertEncountersToInfo(encounters, ad.playerAnalyzer.currentSpectatedID)
				ad.app.UpdateEncounters(encounterInfos)
			}

			// Sync statistics and game state from PlayerAnalyzer to App
			if ad.app != nil {
				ad.app.mu.Lock()
				ad.app.stats.SniperKills = ad.playerAnalyzer.TotalSniperKills
				ad.app.stats.UpsetVictories = ad.playerAnalyzer.TotalUpsetVictories
				ad.app.stats.DamageDetections = ad.playerAnalyzer.TotalDamageDetections
				ad.app.lastGameState = gameState // Store game state for GetStatus
				ad.app.mu.Unlock()
			}

			// Only switch automatically during gameplay phases
			if roundPhase == "live" || roundPhase == "freezetime" || roundPhase == "warmup" || roundPhase == "timeout" {
				LogVerbose("[AUTODIRECTOR] Analyzing game state (phase: %s)...", roundPhase)

				// Sync with actual spectated player from GSI
				actualSpectatedID := ad.gsiServer.GetCurrentlySpectatedPlayer()
				if actualSpectatedID != "" && actualSpectatedID != ad.playerAnalyzer.currentSpectatedID {
					LogInfo("🔄 Synced spectated player: CS is on different player than expected")
					ad.playerAnalyzer.SyncCurrentPlayer(actualSpectatedID)
				}

				bestSteamID := ad.playerAnalyzer.GetBestPlayerToSpectate(gameState, roundPhase)

				if bestSteamID != "" {
					ad.spectatorController.SwitchToPlayer(bestSteamID)

					// Record switch for GUI
					if ad.app != nil {
						ad.app.RecordSwitch()
					}
				} else {
					LogVerbose("[AUTODIRECTOR] No player switch needed (rate limiting or no encounters)")
				}
			} else {
				LogVerbose("[AUTODIRECTOR] Waiting for gameplay phase (current: %s)", roundPhase)
			}
		}
	}

	LogInfo("AutoDirector main loop stopped")
}

// extractPlayers extracts player info from game state for GUI display
func extractPlayers(gameState map[string]interface{}) []PlayerInfo {
	players := make([]PlayerInfo, 0)

	allPlayers, ok := gameState["allplayers"].(map[string]interface{})
	if !ok {
		return players
	}

	for steamID, playerData := range allPlayers {
		playerMap, ok := playerData.(map[string]interface{})
		if !ok {
			continue
		}

		// Only include alive players
		if state, ok := playerMap["state"].(map[string]interface{}); ok {
			if health, ok := state["health"].(float64); ok {
				if health <= 0 {
					continue // Skip dead players
				}

				player := PlayerInfo{
					SteamID: steamID,
					Health:  int(health),
				}

				// Observer slot (for stable sorting)
				if observerSlot, ok := playerMap["observer_slot"]; ok {
					if slotFloat, ok := observerSlot.(float64); ok {
						player.Slot = int(slotFloat) + 1 // GSI uses 0-based index
					}
				}

				// Name
				if name, ok := playerMap["name"].(string); ok {
					player.Name = name
				}

				// Team
				if team, ok := playerMap["team"].(string); ok {
					player.Team = team
				}

				// Armor
				if armor, ok := state["armor"].(float64); ok {
					player.Armor = int(armor)
				}

				// Money
				if money, ok := state["money"].(float64); ok {
					player.Money = int(money)
				}

				// Equipment value
				if equipValue, ok := state["equip_value"].(float64); ok {
					player.EquipmentValue = int(equipValue)
				}

				// Defuser
				if defuser, ok := state["defusekit"].(bool); ok {
					player.HasDefuser = defuser
				}

				// Match stats
				if matchStats, ok := playerMap["match_stats"].(map[string]interface{}); ok {
					if kills, ok := matchStats["kills"].(float64); ok {
						player.Kills = int(kills)
					}
					if deaths, ok := matchStats["deaths"].(float64); ok {
						player.Deaths = int(deaths)
					}
				}

				// Active weapon
				if weapons, ok := playerMap["weapons"].(map[string]interface{}); ok {
					for _, weaponData := range weapons {
						if weaponMap, ok := weaponData.(map[string]interface{}); ok {
							if state, ok := weaponMap["state"].(string); ok {
								if state == "active" {
									if name, ok := weaponMap["name"].(string); ok {
										player.ActiveWeapon = name
									}
									break
								}
							}
						}
					}
				}

				// Position
				if position, ok := playerMap["position"].(string); ok {
					player.Position = parsePositionString(position)
				}

				players = append(players, player)
			}
		}
	}

	return players
}

// convertEncountersToInfo converts Encounter structs to EncounterInfo for the frontend
func convertEncountersToInfo(encounters []Encounter, currentSpectatedID string) []EncounterInfo {
	infos := make([]EncounterInfo, 0, len(encounters))

	// Limit to top 10 encounters to avoid overwhelming the UI
	maxEncounters := 10
	if len(encounters) > maxEncounters {
		encounters = encounters[:maxEncounters]
	}

	for _, enc := range encounters {
		isCurrentPlayer := (enc.Player1.SteamID == currentSpectatedID || enc.Player2.SteamID == currentSpectatedID)

		info := EncounterInfo{
			Player1:         enc.Player1.Name,
			Player2:         enc.Player2.Name,
			Player1Team:     enc.Player1.Team,
			Player2Team:     enc.Player2.Team,
			Distance:        enc.Distance,
			Priority:        enc.Priority,
			IsCurrentPlayer: isCurrentPlayer,
		}

		infos = append(infos, info)
	}

	return infos
}

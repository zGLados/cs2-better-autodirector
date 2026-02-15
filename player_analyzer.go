package main

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PlayerInfo contains relevant player data
type PlayerInfo struct {
	SteamID        string
	Name           string
	Team           string
	Position       Position
	Health         int
	Armor          int
	Money          int
	Kills          int
	Deaths         int
	EquipmentValue int
	HasDefuser     bool
}

// Position represents a 3D position
type Position struct {
	X float64
	Y float64
	Z float64
}

// Encounter represents a potential encounter
type Encounter struct {
	Player1  PlayerInfo
	Player2  PlayerInfo
	Distance float64
	Priority float64
}

// PlayerAnalyzer analyzes the game state and finds the best action
type PlayerAnalyzer struct {
	lastSwitchTime          time.Time
	minSwitchInterval       time.Duration
	currentSpectatedID      string
	positionWarningShown    bool
	currentPlayerSwitchTime time.Time     // Track when we switched to current player
	maxStickyTime           time.Duration // Max time to stick with current player
}

// NewPlayerAnalyzer creates a new Player Analyzer
func NewPlayerAnalyzer() *PlayerAnalyzer {
	return &PlayerAnalyzer{
		lastSwitchTime:          time.Now(),
		minSwitchInterval:       2 * time.Second, // Reduced from 3s for more responsive switching
		positionWarningShown:    false,
		currentPlayerSwitchTime: time.Now(),
		maxStickyTime:           15 * time.Second, // Max 15s on one player
	}
}

// CalculateDistance2D calculates the 2D distance between two positions
func CalculateDistance2D(pos1, pos2 Position) float64 {
	dx := pos2.X - pos1.X
	dy := pos2.Y - pos1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// CalculateDistance3D calculates the 3D distance between two positions
func CalculateDistance3D(pos1, pos2 Position) float64 {
	dx := pos2.X - pos1.X
	dy := pos2.Y - pos1.Y
	dz := pos2.Z - pos1.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// parsePositionString parses position from string format "X, Y, Z"
// CS2 sends position as string like "-1520.06, 430.89, -63.97"
func parsePositionString(posStr string) Position {
	parts := strings.Split(posStr, ",")
	if len(parts) != 3 {
		LogVerbose("[ANALYZER] ⚠️  Invalid position string format: '%s'", posStr)
		return Position{X: 0, Y: 0, Z: 0}
	}

	x, errX := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	y, errY := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	z, errZ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)

	if errX != nil || errY != nil || errZ != nil {
		LogVerbose("[ANALYZER] ⚠️  Error parsing position string '%s': X=%v, Y=%v, Z=%v",
			posStr, errX, errY, errZ)
		return Position{X: 0, Y: 0, Z: 0}
	}

	return Position{X: x, Y: y, Z: z}
}

// GetPlayerInfo extracts player data from the game state
func GetPlayerInfo(gameState map[string]interface{}) []PlayerInfo {
	var players []PlayerInfo

	allPlayers, ok := gameState["allplayers"].(map[string]interface{})
	if !ok {
		return players
	}

	for steamID, playerData := range allPlayers {
		playerMap, ok := playerData.(map[string]interface{})
		if !ok {
			continue
		}

		// Only living players
		state, _ := playerMap["state"].(map[string]interface{})
		health := getIntValue(state, "health")
		if health <= 0 {
			continue
		}

		// Extract position
		// CS2 sends position as STRING in format "X, Y, Z"
		var position Position
		if posData, hasPos := playerMap["position"]; hasPos {
			// Try as string first (CS2 format)
			if posStr, ok := posData.(string); ok {
				position = parsePositionString(posStr)

				if VerboseMode {
					if position.X == 0 && position.Y == 0 && position.Z == 0 {
						LogVerbose("[ANALYZER] ⚠️  Player %s: position string is '%s' but parsed as ZERO!",
							getStringValue(playerMap, "name"), posStr)
					} else {
						LogVerbose("[ANALYZER] ✓ Player %s position: X=%.1f Y=%.1f Z=%.1f (from string '%s')",
							getStringValue(playerMap, "name"), position.X, position.Y, position.Z, posStr)
					}
				}
			} else if posMap, ok := posData.(map[string]interface{}); ok {
				// Fallback: try as map (old format or other implementations)
				position = Position{
					X: getFloatValue(posMap, "x"),
					Y: getFloatValue(posMap, "y"),
					Z: getFloatValue(posMap, "z"),
				}

				if VerboseMode {
					LogVerbose("[ANALYZER] ✓ Player %s position from map: X=%.1f Y=%.1f Z=%.1f",
						getStringValue(playerMap, "name"), position.X, position.Y, position.Z)
				}
			} else {
				if VerboseMode {
					LogVerbose("[ANALYZER] ⚠️  Player %s: position has unexpected type %T",
						getStringValue(playerMap, "name"), posData)
				}
				position = Position{X: 0, Y: 0, Z: 0}
			}
		} else {
			if VerboseMode {
				LogVerbose("[ANALYZER] ❌ Player %s: NO 'position' key in GSI data!",
					getStringValue(playerMap, "name"))
			}
			position = Position{X: 0, Y: 0, Z: 0}
		}

		matchStats, _ := playerMap["match_stats"].(map[string]interface{})

		player := PlayerInfo{
			SteamID:        steamID,
			Name:           getStringValue(playerMap, "name"),
			Team:           getStringValue(playerMap, "team"),
			Position:       position,
			Health:         health,
			Armor:          getIntValue(state, "armor"),
			Money:          getIntValue(state, "money"),
			Kills:          getIntValue(matchStats, "kills"),
			Deaths:         getIntValue(matchStats, "deaths"),
			EquipmentValue: getIntValue(state, "equip_value"),
			HasDefuser:     getBoolValue(state, "defusekit"),
		}

		players = append(players, player)
	}

	return players
}

// PredictEncounters finds potential encounters between players
func PredictEncounters(players []PlayerInfo) []Encounter {
	var encounters []Encounter

	// Check if we have position data
	hasPositionData := false
	for _, p := range players {
		if p.Position.X != 0 || p.Position.Y != 0 || p.Position.Z != 0 {
			hasPositionData = true
			break
		}
	}

	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			p1 := players[i]
			p2 := players[j]

			// Only players from different teams
			if p1.Team == p2.Team || p1.Team == "Unknown" || p2.Team == "Unknown" {
				continue
			}

			var distance float64
			if hasPositionData {
				distance = CalculateDistance2D(p1.Position, p2.Position)
				// Focus on close to medium range encounters (where kills happen)
				// 0-500: Very close (grenades, shotguns)
				// 500-1500: Medium range (rifles)
				// 1500-2000: Long range (AWP, etc.)
				if distance >= 2000 {
					continue
				}
			} else {
				// No position data - use 0 as placeholder and rely on other factors
				distance = 0
			}

			priority := CalculateEncounterPriority(p1, p2, distance, hasPositionData)
			encounters = append(encounters, Encounter{
				Player1:  p1,
				Player2:  p2,
				Distance: distance,
				Priority: priority,
			})
		}
	}

	// Sort by priority (highest first)
	sort.Slice(encounters, func(i, j int) bool {
		return encounters[i].Priority > encounters[j].Priority
	})

	return encounters
}

// CalculateEncounterPriority calculates the priority of an encounter
func CalculateEncounterPriority(p1, p2 PlayerInfo, distance float64, hasPositionData bool) float64 {
	priority := 0.0

	// 1. Distance-based priority (HEAVILY weighted - this is where action happens!)
	if hasPositionData {
		if distance < 300 {
			priority += 150 // VERY high priority - imminent fight!
		} else if distance < 600 {
			priority += 120 // High priority - close combat
		} else if distance < 1000 {
			priority += 80 // Medium-close range
		} else if distance < 1500 {
			priority += 50 // Medium range
		} else {
			priority += 20 // Long range - less likely to result in kills
		}
	} else {
		// Without position data, give base priority for any matchup
		priority += 20
	}

	// 2. Equipment-Value (increased weight when no position data)
	avgEquipment := float64(p1.EquipmentValue+p2.EquipmentValue) / 2.0
	if hasPositionData {
		priority += avgEquipment / 200.0
	} else {
		priority += avgEquipment / 100.0 // Double weight without position
	}

	// 3. Skill level (Kills) (increased weight when no position data)
	avgKills := float64(p1.Kills+p2.Kills) / 2.0
	if hasPositionData {
		priority += avgKills * 3.0
	} else {
		priority += avgKills * 5.0 // More weight without position
	}

	// 4. Health status (low HP = more exciting, but also more likely to die soon)
	if p1.Health < 50 || p2.Health < 50 {
		priority += 10 // Reduced from 15 - low HP is risky for spectating
	}
	if p1.Health < 30 || p2.Health < 30 {
		priority += 10 // Reduced from 15 - very risky
	}

	// 5. Defuser bonus
	if p1.HasDefuser || p2.HasDefuser {
		priority += 20
	}

	return priority
}

// GetBestPlayerToSpectate finds the best player to observe
func (pa *PlayerAnalyzer) GetBestPlayerToSpectate(gameState map[string]interface{}) string {
	players := GetPlayerInfo(gameState)

	// PRIORITY: Check if currently spectated player is dead
	if pa.currentSpectatedID != "" {
		currentPlayerDead := true
		for _, p := range players {
			if p.SteamID == pa.currentSpectatedID {
				currentPlayerDead = false
				break
			}
		}

		if currentPlayerDead {
			LogInfo("☠️  Currently spectated player DIED - forcing immediate switch!")
			pa.currentSpectatedID = "" // Reset to force switch
			// Skip rate limiting when player dies
		} else {
			// Normal rate limiting for alive player
			if time.Since(pa.lastSwitchTime) < pa.minSwitchInterval {
				timeSinceLastSwitch := time.Since(pa.lastSwitchTime).Seconds()
				LogVerbose("[ANALYZER] Rate limiting active - waited %.1fs of %.0fs", timeSinceLastSwitch, pa.minSwitchInterval.Seconds())
				return ""
			}
		}
	} else {
		// No one spectated yet, check normal rate limit
		if time.Since(pa.lastSwitchTime) < pa.minSwitchInterval {
			timeSinceLastSwitch := time.Since(pa.lastSwitchTime).Seconds()
			LogVerbose("[ANALYZER] Rate limiting active - waited %.1fs of %.0fs", timeSinceLastSwitch, pa.minSwitchInterval.Seconds())
			return ""
		}
	}

	if len(players) == 0 {
		LogVerbose("[ANALYZER] No players found in game state")
		return ""
	}

	LogVerbose("[ANALYZER] Found %d alive players", len(players))

	// Check for position data and warn once
	if !pa.positionWarningShown {
		hasPositionData := false
		positionKeyExists := false

		for _, p := range players {
			if p.Position.X != 0 || p.Position.Y != 0 || p.Position.Z != 0 {
				hasPositionData = true
				break
			}
		}

		// Check if position key exists in game state (as string or map)
		if allPlayers, ok := gameState["allplayers"].(map[string]interface{}); ok {
			for _, playerData := range allPlayers {
				if playerMap, ok := playerData.(map[string]interface{}); ok {
					if posData, hasPos := playerMap["position"]; hasPos {
						// Position can be string or map
						if posStr, ok := posData.(string); ok && posStr != "" {
							positionKeyExists = true
							break
						} else if _, ok := posData.(map[string]interface{}); ok {
							positionKeyExists = true
							break
						}
					}
				}
			}
		}

		if !positionKeyExists {
			LogInfo("❌ CRITICAL: NO POSITION DATA IN GSI!")
			LogInfo("   The 'position' key is missing from GSI data")
			LogInfo("   Check that you are in GOTV/Demo mode, not spectator")
			LogInfo("   Or CS2 is not sending position data at all")
			LogInfo("   → Using FALLBACK: equipment/kills only")
			pa.positionWarningShown = true
		} else if !hasPositionData {
			LogInfo("⚠️  Position coordinates are all ZERO")
			LogInfo("   Position key exists but coords are 0,0,0")
			LogInfo("   This might be normal during freezetime/warmup")
			LogInfo("   → Using FALLBACK: equipment/kills only")
			pa.positionWarningShown = true
		} else {
			LogInfo("✅ Position data is available and working!")
			pa.positionWarningShown = true
		}
	}

	// Find encounters
	encounters := PredictEncounters(players)
	LogVerbose("[ANALYZER] Detected %d potential encounters", len(encounters))

	if len(encounters) > 0 {
		// Apply bonus to encounters involving the currently spectated player
		// BUT only if:
		// 1. The encounter has decent priority (>80) to begin with
		// 2. We haven't been stuck on this player too long (>15s)
		//
		// EXCEPTIONS for sticky time limit:
		// - Less than 4 players alive (clutch situation) → unlimited sticky time
		// - Current player in active encounter (Priority >80) → unlimited sticky time
		timeOnCurrentPlayer := time.Since(pa.currentPlayerSwitchTime).Seconds()
		baseIsStuckTooLong := timeOnCurrentPlayer > pa.maxStickyTime.Seconds()

		// Exception 1: Clutch situation (few players alive)
		isClutchSituation := len(players) < 4
		if isClutchSituation {
			LogVerbose("[ANALYZER] 🎯 Clutch situation (%d players alive) - sticky time limit disabled", len(players))
		}

		for i := range encounters {
			if pa.currentSpectatedID != "" {
				if encounters[i].Player1.SteamID == pa.currentSpectatedID ||
					encounters[i].Player2.SteamID == pa.currentSpectatedID {
					originalPriority := encounters[i].Priority

					// Check if sticky time limit applies
					isStuckTooLong := baseIsStuckTooLong

					// Exception 1: Clutch situation - no time limit
					if isClutchSituation {
						isStuckTooLong = false
					}

					// Exception 2: Active encounter (Priority >80) - current player is in action
					if originalPriority > 80.0 {
						if baseIsStuckTooLong {
							LogVerbose("[ANALYZER] 🔥 Active encounter detected - sticky time limit disabled for this fight")
						}
						isStuckTooLong = false
					}

					// Only give bonus if base priority is good enough (>80 = close encounter)
					// AND we haven't been on this player too long (unless exceptions apply)
					if originalPriority > 80.0 && !isStuckTooLong {
						encounters[i].Priority += 30.0 // Reduced from 100 to 30
						LogVerbose("[ANALYZER] ⭐ Current player in encounter: +30 priority (%.1f → %.1f) [%.1fs on player]",
							originalPriority, encounters[i].Priority, timeOnCurrentPlayer)
					} else if isStuckTooLong {
						LogVerbose("[ANALYZER] ⏱️  Sticky time exceeded (%.1fs > %.0fs) - no bonus applied",
							timeOnCurrentPlayer, pa.maxStickyTime.Seconds())
					} else {
						LogVerbose("[ANALYZER] 📉 Priority too low (%.1f < 80) - no bonus applied", originalPriority)
					}
				}
			}
		}

		// Re-sort after applying bonuses
		sort.Slice(encounters, func(i, j int) bool {
			return encounters[i].Priority > encounters[j].Priority
		})

		// Take the most important encounter
		encounter := encounters[0]

		LogInfo("⚔️  ENCOUNTER: %s (%s) vs %s (%s) | Distance: %.0f units | Priority: %.1f",
			encounter.Player1.Name, encounter.Player1.Team,
			encounter.Player2.Name, encounter.Player2.Team,
			encounter.Distance, encounter.Priority)

		// Choose the better player from the encounter
		var bestPlayerID string
		var bestPlayerName string

		// If currently spectating one of the players in this encounter, stay with them
		if pa.currentSpectatedID == encounter.Player1.SteamID {
			bestPlayerID = encounter.Player1.SteamID
			bestPlayerName = encounter.Player1.Name
			LogVerbose("[ANALYZER] ✓ Staying with current player %s in this encounter", bestPlayerName)
		} else if pa.currentSpectatedID == encounter.Player2.SteamID {
			bestPlayerID = encounter.Player2.SteamID
			bestPlayerName = encounter.Player2.Name
			LogVerbose("[ANALYZER] ✓ Staying with current player %s in this encounter", bestPlayerName)
		} else {
			// Choose based on equipment/kills/health
			p1Score := float64(encounter.Player1.EquipmentValue)/100 + float64(encounter.Player1.Kills)*10
			p2Score := float64(encounter.Player2.EquipmentValue)/100 + float64(encounter.Player2.Kills)*10

			// Bonus for more health
			if encounter.Player1.Health > 50 {
				p1Score += 5
			}
			if encounter.Player2.Health > 50 {
				p2Score += 5
			}

			if p1Score > p2Score {
				bestPlayerID = encounter.Player1.SteamID
				bestPlayerName = encounter.Player1.Name
				LogVerbose("[ANALYZER] → Switching to %s (score: %.1f vs %.1f)", bestPlayerName, p1Score, p2Score)
			} else {
				bestPlayerID = encounter.Player2.SteamID
				bestPlayerName = encounter.Player2.Name
				LogVerbose("[ANALYZER] → Switching to %s (score: %.1f vs %.1f)", bestPlayerName, p2Score, p1Score)
			}
		}

		// Verify player is still alive before switching
		playerIsAlive := false
		for _, p := range players {
			if p.SteamID == bestPlayerID {
				playerIsAlive = true
				break
			}
		}

		if !playerIsAlive {
			LogVerbose("[ANALYZER] ⚠️  Target player %s is DEAD, skipping switch", bestPlayerName)
			return ""
		}

		// Only switch if it's a different player
		if bestPlayerID != pa.currentSpectatedID {
			pa.lastSwitchTime = time.Now()
			pa.currentSpectatedID = bestPlayerID
			pa.currentPlayerSwitchTime = time.Now() // Reset sticky timer
			return bestPlayerID
		} else {
			LogVerbose("[ANALYZER] Already spectating this player, no switch needed")
		}
	} else {
		LogVerbose("[ANALYZER] No encounters detected, using fallback (highest equipment/kills)")
		// Fallback: Player with highest equipment/kills
		sort.Slice(players, func(i, j int) bool {
			if players[i].EquipmentValue != players[j].EquipmentValue {
				return players[i].EquipmentValue > players[j].EquipmentValue
			}
			return players[i].Kills > players[j].Kills
		})

		if len(players) > 0 {
			bestPlayerID := players[0].SteamID
			if bestPlayerID != pa.currentSpectatedID {
				pa.lastSwitchTime = time.Now()
				pa.currentSpectatedID = bestPlayerID
				pa.currentPlayerSwitchTime = time.Now() // Reset sticky timer
				return bestPlayerID
			}
		}
	}

	return ""
}

// SyncCurrentPlayer updates the analyzer's internal state to match the actual spectated player from GSI
// This is called when we detect that CS is spectating a different player than we think
func (pa *PlayerAnalyzer) SyncCurrentPlayer(steamID string) {
	pa.currentSpectatedID = steamID
	pa.currentPlayerSwitchTime = time.Now() // Reset sticky timer to prevent immediate switch
	LogVerbose("[ANALYZER] Synced to actual spectated player: %s", steamID)
}

// Helper functions for safe type assertions
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return "Unknown"
}

func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0.0
}

func getBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

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
	ActiveWeapon   string // e.g. "weapon_awp", "weapon_ak47"
}

// Position represents a 3D position
type Position struct {
	X float64
	Y float64
	Z float64
}

// isLongRangeWeapon checks if a weapon is effective at long range
func isLongRangeWeapon(weapon string) bool {
	longRangeWeapons := []string{
		"weapon_awp",    // AWP Sniper
		"weapon_ssg08",  // Scout
		"weapon_aug",    // AUG (scoped rifle)
		"weapon_sg556",  // SG553/SG556 (scoped rifle)
		"weapon_deagle", // Desert Eagle (accurate at range)
	}
	for _, w := range longRangeWeapons {
		if weapon == w {
			return true
		}
	}
	return false
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
	previousSpectatedID     string // Track previous player to avoid immediate back-switching
	positionWarningShown    bool
	currentPlayerSwitchTime time.Time            // Track when we switched to current player
	maxStickyTime           time.Duration        // Max time to stick with current player
	lastEncounterPlayers    []string             // SteamIDs of last encounter (to detect upset victories)
	combatWinnerBonus       map[string]float64   // Temporary bonus for upset victory winners
	combatWinnerTime        map[string]time.Time // When the bonus was given
}

// NewPlayerAnalyzer creates a new Player Analyzer
func NewPlayerAnalyzer() *PlayerAnalyzer {
	return &PlayerAnalyzer{
		lastSwitchTime:          time.Now(),
		minSwitchInterval:       2 * time.Second, // Reduced from 3s for more responsive switching
		positionWarningShown:    false,
		currentPlayerSwitchTime: time.Now(),
		maxStickyTime:           15 * time.Second, // Max 15s on one player
		lastEncounterPlayers:    make([]string, 0),
		combatWinnerBonus:       make(map[string]float64),
		combatWinnerTime:        make(map[string]time.Time),
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

				// Determine max encounter distance based on weapons
				// Normal: 1500 units (rifles)
				// Long-range (AWP/SSG/AUG/SG/Deagle): 2500 units
				maxDistance := 1500.0
				hasLongRange := isLongRangeWeapon(p1.ActiveWeapon) || isLongRangeWeapon(p2.ActiveWeapon)
				if hasLongRange {
					maxDistance = 2500.0
				}

				if distance >= maxDistance {
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

	// Check if long-range weapons are involved
	p1HasLongRange := isLongRangeWeapon(p1.ActiveWeapon)
	p2HasLongRange := isLongRangeWeapon(p2.ActiveWeapon)
	bothHaveLongRange := p1HasLongRange && p2HasLongRange
	eitherHasLongRange := p1HasLongRange || p2HasLongRange

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
		} else if distance < 2500 {
			// Long range (1500-2500) - typically only sniper fights
			if eitherHasLongRange {
				priority += 80 // Good priority for sniper duels
				if bothHaveLongRange {
					priority += 50 // Extra bonus for sniper vs sniper
				}
			} else {
				priority += 20 // Low priority if no long-range weapons
			}
		} else {
			priority += 10 // Very long range - unlikely to result in action
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
func (pa *PlayerAnalyzer) GetBestPlayerToSpectate(gameState map[string]interface{}, roundPhase string) string {
	players := GetPlayerInfo(gameState)

	// Check for upset victories: If one player from last encounter died, reward the survivor
	upsetVictoryOccurred := false
	upsetWinnerID := ""

	if len(pa.lastEncounterPlayers) == 2 {
		player1ID := pa.lastEncounterPlayers[0]
		player2ID := pa.lastEncounterPlayers[1]

		player1Alive := false
		player2Alive := false

		for _, p := range players {
			if p.SteamID == player1ID {
				player1Alive = true
			}
			if p.SteamID == player2ID {
				player2Alive = true
			}
		}

		// One died, one survived → combat concluded
		if player1Alive != player2Alive {
			winnerID := ""
			if player1Alive {
				winnerID = player1ID
			} else {
				winnerID = player2ID
			}

			// If the winner is NOT the player we were spectating → upset victory!
			if winnerID != pa.currentSpectatedID && pa.currentSpectatedID != "" {
				// Give significant bonus for upset victory
				pa.combatWinnerBonus[winnerID] = 100.0
				pa.combatWinnerTime[winnerID] = time.Now()
				upsetVictoryOccurred = true
				upsetWinnerID = winnerID
				LogInfo("🏆 UPSET VICTORY! %s won the fight - forcing immediate switch!", getPlayerNameByID(winnerID, players))
			}

			// Clear last encounter
			pa.lastEncounterPlayers = make([]string, 0)
		}
	}

	// Clean up expired combat winner bonuses (older than 10 seconds)
	for steamID, bonusTime := range pa.combatWinnerTime {
		if time.Since(bonusTime) > 10*time.Second {
			delete(pa.combatWinnerBonus, steamID)
			delete(pa.combatWinnerTime, steamID)
		}
	}

	// PRIORITY: Check if currently spectated player is dead OR upset victory occurred
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
		} else if upsetVictoryOccurred {
			// Skip rate limiting for upset victory - we want immediate switch to winner
			LogVerbose("[ANALYZER] Upset victory detected - bypassing rate limit for immediate switch")
		} else {
			// Phase-specific switching logic
			switchInterval := pa.minSwitchInterval
			maxTimeOnPlayer := pa.maxStickyTime.Seconds()

			switch roundPhase {
			case "warmup", "freezetime":
				// Warmup/Freezetime: max 5 seconds per player for dynamic viewing
				maxTimeOnPlayer = 5.0
				switchInterval = 2 * time.Second // 2 second rate limit for better pacing
				LogVerbose("[ANALYZER] 🔄 %s mode - max %.0fs per player (dynamic)", roundPhase, maxTimeOnPlayer)
			case "timeout":
				// Timeout: max 10 seconds per player
				maxTimeOnPlayer = 10.0
				switchInterval = 2 * time.Second
				LogVerbose("[ANALYZER] ⏸️  Timeout mode - max %.0fs per player", maxTimeOnPlayer)
			default:
				// Normal game: clutch situation check
				isClutchSituation := len(players) <= 4
				if isClutchSituation {
					switchInterval = 1 * time.Second // Faster switching in clutch
				}
			}

			// Check if we've been on current player too long (forced switch)
			timeOnCurrentPlayer := time.Since(pa.currentPlayerSwitchTime).Seconds()
			if timeOnCurrentPlayer >= maxTimeOnPlayer {
				// Force immediate switch by bypassing rate limit (don't log yet, wait until we actually switch)
			} else if time.Since(pa.lastSwitchTime) < switchInterval {
				// Normal rate limiting for alive player
				timeSinceLastSwitch := time.Since(pa.lastSwitchTime).Seconds()
				LogVerbose("[ANALYZER] Rate limiting active - waited %.1fs of %.0fs", timeSinceLastSwitch, switchInterval.Seconds())
				return ""
			}
		}
	} else {
		// No one spectated yet, check normal rate limit
		// BUT: Skip if upset victory occurred
		if !upsetVictoryOccurred {
			switchInterval := pa.minSwitchInterval

			switch roundPhase {
			case "warmup", "freezetime":
				switchInterval = 2 * time.Second // 2 second for better pacing
			case "timeout":
				switchInterval = 2 * time.Second
			default:
				isClutchSituation := len(players) <= 4
				if isClutchSituation {
					switchInterval = 1 * time.Second // Faster switching in clutch
				}
			}

			if time.Since(pa.lastSwitchTime) < switchInterval {
				timeSinceLastSwitch := time.Since(pa.lastSwitchTime).Seconds()
				LogVerbose("[ANALYZER] Rate limiting active - waited %.1fs of %.0fs", timeSinceLastSwitch, switchInterval.Seconds())
				return ""
			}
		}
	}

	if len(players) == 0 {
		LogVerbose("[ANALYZER] No players found in game state")
		return ""
	}

	LogVerbose("[ANALYZER] Found %d alive players", len(players))

	// PRIORITY: If upset victory just occurred, switch to winner immediately
	if upsetVictoryOccurred && upsetWinnerID != "" {
		// Verify winner is still alive
		winnerAlive := false
		for _, p := range players {
			if p.SteamID == upsetWinnerID {
				winnerAlive = true
				break
			}
		}

		if winnerAlive && upsetWinnerID != pa.currentSpectatedID {
			LogInfo("🎯 Switching to upset victory winner immediately!")
			pa.lastSwitchTime = time.Now()
			pa.currentSpectatedID = upsetWinnerID
			pa.currentPlayerSwitchTime = time.Now()
			return upsetWinnerID
		}
	}

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
		// 1. The encounter has decent priority (>140) to begin with
		// 2. We haven't been stuck on this player too long
		//
		// TIME LIMITS:
		// - Normal: 15s limit
		// - Clutch (<4 players) or Active Encounter (Priority >140): 25s limit
		timeOnCurrentPlayer := time.Since(pa.currentPlayerSwitchTime).Seconds()
		baseIsStuckTooLong := timeOnCurrentPlayer > pa.maxStickyTime.Seconds()
		const maxExtendedStickyTime = 25.0 // Absolute maximum for any situation

		// Exception 1: Clutch situation (few players alive)
		isClutchSituation := len(players) <= 4
		if isClutchSituation {
			LogVerbose("[ANALYZER] 🎯 Clutch situation (%d players alive) - faster switching enabled (1s rate limit, +10 bonus)", len(players))
		}

		for i := range encounters {
			if pa.currentSpectatedID != "" {
				if encounters[i].Player1.SteamID == pa.currentSpectatedID ||
					encounters[i].Player2.SteamID == pa.currentSpectatedID {
					originalPriority := encounters[i].Priority

					// Check if sticky time limit applies
					isStuckTooLong := baseIsStuckTooLong
					absoluteMaxExceeded := timeOnCurrentPlayer > maxExtendedStickyTime

					// Exception 1: Clutch situation - extended time limit (25s instead of 15s)
					if isClutchSituation && !absoluteMaxExceeded {
						isStuckTooLong = false
					}

					// Exception 2: Active encounter (Priority >140) - extended time limit (25s instead of 15s)
					if originalPriority > 140.0 && !absoluteMaxExceeded {
						if baseIsStuckTooLong {
							LogVerbose("[ANALYZER] 🔥 Active encounter detected - extended sticky time limit (25s)")
						}
						isStuckTooLong = false
					}

					// Check if absolute maximum is exceeded
					if absoluteMaxExceeded {
						isStuckTooLong = true
					}

					// Only give bonus if base priority is good enough (>140 = close encounter)
					// AND we haven't been on this player too long (unless exceptions apply)
					if originalPriority > 140.0 && !isStuckTooLong {
						// In clutch situations, give smaller bonus to encourage more switching
						bonusAmount := 30.0
						if isClutchSituation {
							bonusAmount = 10.0 // Reduced bonus in clutch for more dynamic switching
						}

						encounters[i].Priority += bonusAmount
						LogVerbose("[ANALYZER] ⭐ Current player in encounter: +%.0f priority (%.1f → %.1f) [%.1fs on player]%s",
							bonusAmount, originalPriority, encounters[i].Priority, timeOnCurrentPlayer,
							func() string {
								if isClutchSituation {
									return " [CLUTCH]"
								}
								return ""
							}())
					} else if isStuckTooLong {
						timeLimit := pa.maxStickyTime.Seconds()
						if absoluteMaxExceeded {
							timeLimit = maxExtendedStickyTime
						}
						LogVerbose("[ANALYZER] ⏱️  Sticky time exceeded (%.1fs > %.0fs) - no bonus applied",
							timeOnCurrentPlayer, timeLimit)
					} else {
						LogVerbose("[ANALYZER] 📉 Priority too low (%.1f < 140) - no bonus applied", originalPriority)
					}
				}
			}
		}

		// Apply combat winner bonus (for upset victories)
		for i := range encounters {
			bonus1 := pa.combatWinnerBonus[encounters[i].Player1.SteamID]
			bonus2 := pa.combatWinnerBonus[encounters[i].Player2.SteamID]

			if bonus1 > 0 {
				encounters[i].Priority += bonus1
				LogVerbose("[ANALYZER] 🏆 Combat winner bonus for %s: +%.0f (upset victory)",
					encounters[i].Player1.Name, bonus1)
			}
			if bonus2 > 0 {
				encounters[i].Priority += bonus2
				LogVerbose("[ANALYZER] 🏆 Combat winner bonus for %s: +%.0f (upset victory)",
					encounters[i].Player2.Name, bonus2)
			}
		}

		// Re-sort after applying bonuses
		sort.Slice(encounters, func(i, j int) bool {
			return encounters[i].Priority > encounters[j].Priority
		})

		// Take the most important encounter
		encounter := encounters[0]

		// Track this encounter for upset victory detection
		pa.lastEncounterPlayers = []string{encounter.Player1.SteamID, encounter.Player2.SteamID}

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
			// Already on this player - check if we need to force switch in warmup/freezetime/timeout
			timeOnCurrent := time.Since(pa.currentPlayerSwitchTime).Seconds()
			var maxTime float64
			forceSwitch := false

			switch roundPhase {
			case "warmup", "freezetime":
				maxTime = 5.0
				forceSwitch = true
			case "timeout":
				maxTime = 10.0
				forceSwitch = true
			default:
				maxTime = pa.maxStickyTime.Seconds()
			}

			if timeOnCurrent >= maxTime {
				if forceSwitch {
					// In warmup/freezetime/timeout: switch to the other player in encounter for variety
					var alternativeID string
					if encounter.Player1.SteamID == pa.currentSpectatedID {
						alternativeID = encounter.Player2.SteamID
					} else {
						alternativeID = encounter.Player1.SteamID
					}

					LogInfo("⏱️  Max time (%.0fs) exceeded in %s - switching for variety", maxTime, roundPhase)
					pa.lastSwitchTime = time.Now()
					pa.previousSpectatedID = pa.currentSpectatedID
					pa.currentSpectatedID = alternativeID
					pa.currentPlayerSwitchTime = time.Now()
					return alternativeID
				} else {
					// Normal game: reset timer with offset to retry later
					LogVerbose("[ANALYZER] Max time exceeded but already on best player - will retry in 3s")
					pa.currentPlayerSwitchTime = time.Now().Add(-time.Duration(maxTime-3) * time.Second)
				}
			} else {
				LogVerbose("[ANALYZER] Already spectating this player, no switch needed")
			}
		}
	} else {
		LogVerbose("[ANALYZER] No encounters detected, using fallback (highest equipment/kills)")

		// Clear last encounter since no active fight
		pa.lastEncounterPlayers = make([]string, 0)

		// Apply combat winner bonus even in fallback mode
		for i := range players {
			if bonus, exists := pa.combatWinnerBonus[players[i].SteamID]; exists {
				// Add to equipment value for sorting (scaled appropriately)
				players[i].EquipmentValue += int(bonus * 100) // Scale bonus to equipment value range
				LogVerbose("[ANALYZER] 🏆 Combat winner bonus for %s in fallback mode", players[i].Name)
			}
		}

		// Fallback: Player with highest equipment/kills
		sort.Slice(players, func(i, j int) bool {
			if players[i].EquipmentValue != players[j].EquipmentValue {
				return players[i].EquipmentValue > players[j].EquipmentValue
			}
			return players[i].Kills > players[j].Kills
		})

		if len(players) > 0 {
			// In freezetime/warmup: ONLY switch on maxTime (every 5s), not based on "best player"
			// In other phases: Switch to best player immediately (after rate limit)
			if roundPhase == "warmup" || roundPhase == "freezetime" {
				// Force switch mode: only switch every 5 seconds for consistent pacing
				timeOnCurrent := time.Since(pa.currentPlayerSwitchTime).Seconds()
				if timeOnCurrent >= 5.0 {
					// Time to switch: find next player (not current, not previous)
					var nextPlayerID string
					var currentTeam string

					// Find current player's team
					for _, p := range players {
						if p.SteamID == pa.currentSpectatedID {
							currentTeam = p.Team
							break
						}
					}

					// First try to find player from opposite team (but not the previous player)
					for _, p := range players {
						if p.SteamID != pa.currentSpectatedID && p.SteamID != pa.previousSpectatedID && p.Team != currentTeam && p.Team != "Unknown" {
							nextPlayerID = p.SteamID
							break
						}
					}
					// If still no player found, try opposite team without previous player restriction
					if nextPlayerID == "" {
						for _, p := range players {
							if p.SteamID != pa.currentSpectatedID && p.Team != currentTeam && p.Team != "Unknown" {
								nextPlayerID = p.SteamID
								break
							}
						}
					}
					// If no opposite team player found, just take next different player
					if nextPlayerID == "" {
						for _, p := range players {
							if p.SteamID != pa.currentSpectatedID {
								nextPlayerID = p.SteamID
								break
							}
						}
					}

					if nextPlayerID != "" {
						LogInfo("⏱️  Max time (5s) exceeded in %s - switching for variety", roundPhase)
						pa.lastSwitchTime = time.Now()
						pa.previousSpectatedID = pa.currentSpectatedID
						pa.currentSpectatedID = nextPlayerID
						pa.currentPlayerSwitchTime = time.Now()
						return nextPlayerID
					}
				}
				// Not time to switch yet in freezetime/warmup
				return ""
			} else {
				// Normal game phases: switch to best player
				bestPlayerID := players[0].SteamID
				if bestPlayerID != pa.currentSpectatedID {
					pa.lastSwitchTime = time.Now()
					pa.previousSpectatedID = pa.currentSpectatedID
					pa.currentSpectatedID = bestPlayerID
					pa.currentPlayerSwitchTime = time.Now() // Reset sticky timer
					return bestPlayerID
				} else {
					// Already on best player (fallback) - check if we need to force switch in timeout
					timeOnCurrent := time.Since(pa.currentPlayerSwitchTime).Seconds()

					if roundPhase == "timeout" && timeOnCurrent >= 10.0 && len(players) > 1 {
						// In timeout: switch to different player for variety
						var nextPlayerID string
						for _, p := range players {
							if p.SteamID != pa.currentSpectatedID {
								nextPlayerID = p.SteamID
								break
							}
						}

						if nextPlayerID != "" {
							LogInfo("⏱️  Max time (10s) exceeded in timeout - switching for variety")
							pa.lastSwitchTime = time.Now()
							pa.previousSpectatedID = pa.currentSpectatedID
							pa.currentSpectatedID = nextPlayerID
							pa.currentPlayerSwitchTime = time.Now()
							return nextPlayerID
						}
					}
				}
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

func getPlayerNameByID(steamID string, players []PlayerInfo) string {
	for _, p := range players {
		if p.SteamID == steamID {
			return p.Name
		}
	}
	return "Unknown Player"
}

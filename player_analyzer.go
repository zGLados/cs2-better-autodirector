package main

import (
	"math"
	"sort"
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
	lastSwitchTime     time.Time
	minSwitchInterval  time.Duration
	currentSpectatedID string
}

// NewPlayerAnalyzer creates a new Player Analyzer
func NewPlayerAnalyzer() *PlayerAnalyzer {
	return &PlayerAnalyzer{
		lastSwitchTime:    time.Now(),
		minSwitchInterval: 3 * time.Second,
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
		posMap, _ := playerMap["position"].(map[string]interface{})
		position := Position{
			X: getFloatValue(posMap, "x"),
			Y: getFloatValue(posMap, "y"),
			Z: getFloatValue(posMap, "z"),
		}

		// Debug: Log if position is zero (potential issue)
		if VerboseMode && position.X == 0 && position.Y == 0 && position.Z == 0 {
			LogVerbose("[ANALYZER] ⚠️  Player %s has ZERO position - GSI not sending position data!",
				getStringValue(playerMap, "name"))
		} else if VerboseMode {
			LogVerbose("[ANALYZER] Player %s position: X=%.1f Y=%.1f Z=%.1f",
				getStringValue(playerMap, "name"), position.X, position.Y, position.Z)
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

	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			p1 := players[i]
			p2 := players[j]

			// Only players from different teams
			if p1.Team == p2.Team || p1.Team == "Unknown" || p2.Team == "Unknown" {
				continue
			}

			distance := CalculateDistance2D(p1.Position, p2.Position)

			// Only encounters under 3000 units
			if distance < 3000 {
				priority := CalculateEncounterPriority(p1, p2, distance)
				encounters = append(encounters, Encounter{
					Player1:  p1,
					Player2:  p2,
					Distance: distance,
					Priority: priority,
				})
			}
		}
	}

	// Sort by priority (highest first)
	sort.Slice(encounters, func(i, j int) bool {
		return encounters[i].Priority > encounters[j].Priority
	})

	return encounters
}

// CalculateEncounterPriority calculates the priority of an encounter
func CalculateEncounterPriority(p1, p2 PlayerInfo, distance float64) float64 {
	priority := 0.0

	// 1. Distance-based priority
	if distance < 500 {
		priority += 100 // Very high risk
	} else if distance < 1000 {
		priority += 70
	} else if distance < 1500 {
		priority += 50
	} else if distance < 2000 {
		priority += 30
	} else {
		priority += 10
	}

	// 2. Equipment-Value
	avgEquipment := float64(p1.EquipmentValue+p2.EquipmentValue) / 2.0
	priority += avgEquipment / 200.0

	// 3. Skill level (Kills)
	avgKills := float64(p1.Kills+p2.Kills) / 2.0
	priority += avgKills * 3.0

	// 4. Health status (low HP = more exciting)
	if p1.Health < 50 || p2.Health < 50 {
		priority += 15
	}
	if p1.Health < 30 || p2.Health < 30 {
		priority += 15
	}

	// 5. Defuser bonus
	if p1.HasDefuser || p2.HasDefuser {
		priority += 20
	}

	return priority
}

// GetBestPlayerToSpectate finds the best player to observe
func (pa *PlayerAnalyzer) GetBestPlayerToSpectate(gameState map[string]interface{}) string {
	// Rate limiting
	if time.Since(pa.lastSwitchTime) < pa.minSwitchInterval {
		timeSinceLastSwitch := time.Since(pa.lastSwitchTime).Seconds()
		LogVerbose("[ANALYZER] Rate limiting active - waited %.1fs of %.0fs", timeSinceLastSwitch, pa.minSwitchInterval.Seconds())
		return ""
	}

	players := GetPlayerInfo(gameState)
	if len(players) == 0 {
		LogVerbose("[ANALYZER] No players found in game state")
		return ""
	}

	LogVerbose("[ANALYZER] Found %d players", len(players))

	// Find encounters
	encounters := PredictEncounters(players)
	LogVerbose("[ANALYZER] Detected %d potential encounters", len(encounters))

	if len(encounters) > 0 {
		// Take the most important encounter
		encounter := encounters[0]

		LogInfo("⚔️  ENCOUNTER: %s (%s) vs %s (%s) | Distance: %.0f units | Priority: %.1f",
			encounter.Player1.Name, encounter.Player1.Team,
			encounter.Player2.Name, encounter.Player2.Team,
			encounter.Distance, encounter.Priority)

		// Choose the better player
		var bestPlayerID string
		if encounter.Player1.EquipmentValue > encounter.Player2.EquipmentValue ||
			encounter.Player1.Kills > encounter.Player2.Kills {
			bestPlayerID = encounter.Player1.SteamID
			LogVerbose("[ANALYZER] → Switching to %s (better equipment/kills)", encounter.Player1.Name)
		} else {
			bestPlayerID = encounter.Player2.SteamID
			LogVerbose("[ANALYZER] → Switching to %s (better equipment/kills)", encounter.Player2.Name)
		}

		// Only switch if it's a different player
		if bestPlayerID != pa.currentSpectatedID {
			pa.lastSwitchTime = time.Now()
			pa.currentSpectatedID = bestPlayerID
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
				return bestPlayerID
			}
		}
	}

	return ""
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

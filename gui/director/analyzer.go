package director

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
	Slot           int // Observer slot (1-10) for stable sorting
	Position       Position
	Health         int
	Armor          int
	Money          int
	Kills          int
	Deaths         int
	EquipmentValue int
	HasDefuser     bool
	ActiveWeapon   string   // e.g. "weapon_awp", "weapon_ak47"
	Forward        Position // Aim direction (unit vector) from GSI "forward"
	HasForward     bool     // True if GSI sent a usable forward vector
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

// isSniperRifle checks if a weapon is a true sniper rifle (AWP/Scout)
func isSniperRifle(weapon string) bool {
	return weapon == "weapon_awp" || weapon == "weapon_ssg08"
}

// isAWP checks if weapon is AWP (highest tier sniper)
func isAWP(weapon string) bool {
	return weapon == "weapon_awp"
}

// getSniperWeaponBonus returns bonus points for sniper weapons (AWP > Scout)
func (pa *PlayerAnalyzer) getSniperWeaponBonus(weapon string) float64 {
	if weapon == "weapon_awp" {
		return pa.Settings.AWPBonus
	} else if weapon == "weapon_ssg08" {
		return pa.Settings.ScoutBonus
	}
	return 0.0
}

// getRifleWeaponBonus returns bonus points for rifles (AK-47 priority over M4s)
func (pa *PlayerAnalyzer) getRifleWeaponBonus(weapon string) float64 {
	if weapon == "weapon_ak47" {
		return pa.Settings.AK47Bonus
	} else if weapon == "weapon_m4a1" || weapon == "weapon_m4a1_silencer" {
		return 0.0 // M4s are baseline
	}
	return 0.0
}

// isShootingWeapon checks if weapon can deal instant damage (not grenade/knife/bomb)
func isShootingWeapon(weapon string) bool {
	// Exclude grenades, knives, bomb, etc.
	excludedWeapons := []string{
		"weapon_hegrenade", "weapon_molotov", "weapon_incgrenade", "weapon_flashbang",
		"weapon_smokegrenade", "weapon_decoy", "weapon_knife", "weapon_c4", "weapon_taser",
	}
	for _, excluded := range excludedWeapons {
		if weapon == excluded || strings.Contains(weapon, "knife") {
			return false
		}
	}
	// Empty weapon or unknown -> assume not shooting
	if weapon == "" || weapon == "weapon_none" {
		return false
	}
	return true
}

// Aim thresholds (cosine of the angle between where a player looks and the
// direction to the other player)
const (
	aimOnTargetCos = 0.94 // ~20°: crosshair is on/near the enemy
	aimAwayCos     = 0.5  // ~60°: clearly looking somewhere else
)

// Kill-follow constants (any weapon kill)
const (
	killFollowBonus    = 120.0           // Priority bonus for a player who just got a kill
	killFollowDuration = 6 * time.Second // How long the bonus lasts
	minEventSwitchGap  = 750 * time.Millisecond
)

// Anti ping-pong (hysteresis) for encounter-based switches. Event switches
// (kill, death, damage, AWP kill, upset) are not affected.
const (
	minDwellTime      = 4 * time.Second // Min time on a player before an encounter-based switch
	switchMargin      = 40.0            // New encounter must beat the current one by this much
	returnMargin      = 80.0            // ...and by more when going back to the previous player
	returnWindow      = 8 * time.Second // What counts as "going back"
	actionProtectTime = 2 * time.Second // Don't leave a player who just dealt damage/killed for a damage event
)

// currentInAction reports whether the spectated player dealt damage or got a
// kill within d (we are probably watching the important fight already).
func (pa *PlayerAnalyzer) currentInAction(d time.Duration) bool {
	id := pa.currentSpectatedID
	if id == "" {
		return false
	}
	if t, ok := pa.damageDealtTime[id]; ok && since(t) < d {
		return true
	}
	if t, ok := pa.killTime[id]; ok && since(t) < d {
		return true
	}
	return false
}

// aimCos returns the cosine between from's aim direction and the direction
// from from's eyes to to's body. ok=false if aim/position data is missing.
func aimCos(from, to PlayerInfo) (float64, bool) {
	if !from.HasForward {
		return 0, false
	}
	if (from.Position == Position{}) || (to.Position == Position{}) {
		return 0, false
	}
	dx := to.Position.X - from.Position.X
	dy := to.Position.Y - from.Position.Y
	dz := (to.Position.Z + 40) - (from.Position.Z + 64) // target chest vs eye height
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	f := from.Forward
	flen := math.Sqrt(f.X*f.X + f.Y*f.Y + f.Z*f.Z)
	if dist < 1 || flen < 0.001 {
		return 0, false
	}
	return (dx*f.X + dy*f.Y + dz*f.Z) / (dist * flen), true
}

// isAimingAt reports whether from is looking at to (within ~20°)
func isAimingAt(from, to PlayerInfo) bool {
	c, ok := aimCos(from, to)
	return ok && c >= aimOnTargetCos
}

// hasEnemyNearby reports whether p has a living enemy within maxDist
func hasEnemyNearby(p PlayerInfo, players []PlayerInfo, maxDist float64) bool {
	for _, o := range players {
		if o.Team == p.Team || o.Team == "Unknown" {
			continue
		}
		if CalculateDistance3D(p.Position, o.Position) <= maxDist {
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
	playerKills             map[string]int       // Track kills per player for kill detection
	sniperKillBonus         map[string]float64   // Temporary bonus for sniper kills
	sniperKillTime          map[string]time.Time // When sniper kill bonus was given
	playerHealth            map[string]int       // Track health per player for damage detection
	damageDealtBonus        map[string]float64   // Temporary bonus for damage dealers
	damageDealtTime         map[string]time.Time // When damage dealt bonus was given
	killBonus               map[string]float64   // Temporary bonus for players who just got a kill (any weapon)
	killTime                map[string]time.Time // When kill bonus was given

	// Public stats counters (for GUI)
	TotalSniperKills      int
	TotalUpsetVictories   int
	TotalDamageDetections int

	// Settings
	Settings *Settings

	// Predictive camera (live rounds): model + switching rules
	Model        *Model
	Predict      PredictConfig
	featureState *FeatureState
	probs        map[string]float64
}

// NewPlayerAnalyzer creates a new Player Analyzer
func NewPlayerAnalyzer() *PlayerAnalyzer {
	return &PlayerAnalyzer{
		lastSwitchTime:          Now(),
		minSwitchInterval:       2 * time.Second, // Reduced from 3s for more responsive switching
		positionWarningShown:    false,
		currentPlayerSwitchTime: Now(),
		Settings:                LoadSettings(), // Load settings from file or use defaults
		Model:                   DefaultModel(),
		Predict:                 DefaultPredictConfig(),
		maxStickyTime:           15 * time.Second, // Max 15s on one player
		lastEncounterPlayers:    make([]string, 0),
		combatWinnerBonus:       make(map[string]float64),
		combatWinnerTime:        make(map[string]time.Time),
		playerKills:             make(map[string]int),
		sniperKillBonus:         make(map[string]float64),
		sniperKillTime:          make(map[string]time.Time),
		playerHealth:            make(map[string]int),
		damageDealtBonus:        make(map[string]float64),
		damageDealtTime:         make(map[string]time.Time),
		killBonus:               make(map[string]float64),
		killTime:                make(map[string]time.Time),
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
func ParsePositionString(posStr string) Position {
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
		health := GetIntValue(state, "health")
		if health <= 0 {
			continue
		}

		// Extract position
		// CS2 sends position as STRING in format "X, Y, Z"
		var position Position
		if posData, hasPos := playerMap["position"]; hasPos {
			// Try as string first (CS2 format)
			if posStr, ok := posData.(string); ok {
				position = ParsePositionString(posStr)

				if LogLevel >= 1 {
					if position.X == 0 && position.Y == 0 && position.Z == 0 {
						LogVerbose("[ANALYZER] ⚠️  Player %s: position string is '%s' but parsed as ZERO!",
							GetStringValue(playerMap, "name"), posStr)
					} else {
						LogVerbose("[ANALYZER] ✓ Player %s position: X=%.1f Y=%.1f Z=%.1f (from string '%s')",
							GetStringValue(playerMap, "name"), position.X, position.Y, position.Z, posStr)
					}
				}
			} else if posMap, ok := posData.(map[string]interface{}); ok {
				// Fallback: try as map (old format or other implementations)
				position = Position{
					X: GetFloatValue(posMap, "x"),
					Y: GetFloatValue(posMap, "y"),
					Z: GetFloatValue(posMap, "z"),
				}

				if LogLevel >= 1 {
					LogVerbose("[ANALYZER] ✓ Player %s position from map: X=%.1f Y=%.1f Z=%.1f",
						GetStringValue(playerMap, "name"), position.X, position.Y, position.Z)
				}
			} else {
				if LogLevel >= 1 {
					LogVerbose("[ANALYZER] ⚠️  Player %s: position has unexpected type %T",
						GetStringValue(playerMap, "name"), posData)
				}
				position = Position{X: 0, Y: 0, Z: 0}
			}
		} else {
			if LogLevel >= 1 {
				LogVerbose("[ANALYZER] ❌ Player %s: NO 'position' key in GSI data!",
					GetStringValue(playerMap, "name"))
			}
			position = Position{X: 0, Y: 0, Z: 0}
		}

		matchStats, _ := playerMap["match_stats"].(map[string]interface{})

		player := PlayerInfo{
			SteamID:        steamID,
			Name:           GetStringValue(playerMap, "name"),
			Team:           GetStringValue(playerMap, "team"),
			Position:       position,
			Health:         health,
			Armor:          GetIntValue(state, "armor"),
			Money:          GetIntValue(state, "money"),
			Kills:          GetIntValue(matchStats, "kills"),
			Deaths:         GetIntValue(matchStats, "deaths"),
			EquipmentValue: GetIntValue(state, "equip_value"),
			HasDefuser:     GetBoolValue(state, "defusekit"),
			// FIX: ActiveWeapon was never filled here, which silently disabled
			// damage-dealer detection, sniper-kill detection and long-range logic.
			ActiveWeapon: getActiveWeapon(playerMap),
		}
		if slot, ok := playerMap["observer_slot"].(float64); ok {
			player.Slot = int(slot) + 1
		}
		// Aim direction: CS2 sends "forward" as a string "x, y, z" (unit vector)
		if fwdStr, ok := playerMap["forward"].(string); ok && fwdStr != "" {
			fwd := ParsePositionString(fwdStr)
			if fwd.X != 0 || fwd.Y != 0 || fwd.Z != 0 {
				player.Forward = fwd
				player.HasForward = true
			}
		}

		players = append(players, player)
	}

	return players
}

// PredictEncounters finds potential encounters between players
func (pa *PlayerAnalyzer) PredictEncounters(players []PlayerInfo) []Encounter {
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
				// Check vertical distance - skip if players are on different floors
				// (e.g., different levels in Nuke, Vertigo, etc.)
				// Note: Can be adjusted for specific maps if needed
				zDiff := math.Abs(p1.Position.Z - p2.Position.Z)
				if zDiff > pa.Settings.VerticalDiffThreshold {
					// Players likely separated by walls/floors - not a real encounter
					continue
				}

				// Use 3D distance (including height) for more accurate encounter detection
				distance = CalculateDistance3D(p1.Position, p2.Position)

				// Determine max encounter distance based on weapons
				// Normal: configurable (default 1500 units)
				// Long-range (AWP/SSG/AUG/SG/Deagle): configurable + 1000 units
				maxDistance := pa.Settings.MaxEncounterDistance
				hasLongRange := isLongRangeWeapon(p1.ActiveWeapon) || isLongRangeWeapon(p2.ActiveWeapon)
				if hasLongRange {
					maxDistance = pa.Settings.MaxEncounterDistance + 1000.0
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
					priority += 100 // INCREASED: Strong bonus for sniper vs sniper (was 50)
					// Extra bonus if AWP is involved
					if isAWP(p1.ActiveWeapon) || isAWP(p2.ActiveWeapon) {
						priority += 30 // AWP duel bonus
					}
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

	// 6. Aim direction: distance alone can't see walls. If someone is aiming at
	// the other player a fight is likely; if neither is looking at the other,
	// they are probably separated by a wall/floor.
	c1, ok1 := aimCos(p1, p2)
	c2, ok2 := aimCos(p2, p1)
	if ok1 && ok2 {
		a1 := c1 >= aimOnTargetCos
		a2 := c2 >= aimOnTargetCos
		switch {
		case a1 && a2:
			priority += 70 // Both looking at each other: duel
		case a1 || a2:
			priority += 40 // One is aiming at the other
		case c1 < aimAwayCos && c2 < aimAwayCos:
			priority -= 30 // Neither looking at the other
		}
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
				pa.combatWinnerBonus[winnerID] = pa.Settings.UpsetVictoryBonus
				pa.combatWinnerTime[winnerID] = Now()
				pa.TotalUpsetVictories++ // Increment stats counter
				upsetVictoryOccurred = true
				upsetWinnerID = winnerID
				LogInfo("🏆 UPSET VICTORY! %s won the fight - forcing immediate switch!", getPlayerNameByID(winnerID, players))
			}

			// Clear last encounter
			pa.lastEncounterPlayers = make([]string, 0)
		}
	}

	// Clean up expired combat winner bonuses (older than configured duration)
	for steamID, bonusTime := range pa.combatWinnerTime {
		if since(bonusTime) > time.Duration(pa.Settings.UpsetVictoryDuration)*time.Second {
			delete(pa.combatWinnerBonus, steamID)
			delete(pa.combatWinnerTime, steamID)
		}
	}

	// Clean up expired sniper kill bonuses (older than configured duration)
	for steamID, bonusTime := range pa.sniperKillTime {
		if since(bonusTime) > time.Duration(pa.Settings.SniperKillDuration)*time.Second {
			delete(pa.sniperKillBonus, steamID)
			delete(pa.sniperKillTime, steamID)
		}
	}

	// Clean up expired kill-follow bonuses
	for steamID, bonusTime := range pa.killTime {
		if since(bonusTime) > killFollowDuration {
			delete(pa.killBonus, steamID)
			delete(pa.killTime, steamID)
		}
	}

	// KILL DETECTION: Track kills and detect sniper kills
	sniperKillDetected := false
	sniperKillerID := ""
	killDetected := false
	killerID := ""

	// Demo seeks / reconnects make several kill counters jump in the same
	// update. Those aren't real live kills, so ignore kill-follow then.
	changedKillers := 0
	for _, p := range players {
		if old, ok := pa.playerKills[p.SteamID]; ok && p.Kills > old {
			changedKillers++
		}
	}
	killFollowAllowed := roundPhase == "live" && changedKillers <= 2

	for _, p := range players {
		oldKills, exists := pa.playerKills[p.SteamID]
		currentKills := p.Kills

		// Update kill count
		pa.playerKills[p.SteamID] = currentKills

		// Detect kill (kills increased)
		if exists && currentKills > oldKills {
			killsGained := currentKills - oldKills
			LogVerbose("[ANALYZER] 💀 %s got %d kill(s) (weapon: %s)", p.Name, killsGained, p.ActiveWeapon)

			// Check if it was a sniper kill
			if isSniperRifle(p.ActiveWeapon) {
				// Give massive bonus for sniper kills
				bonusAmount := pa.Settings.SniperKillBonus
				if isAWP(p.ActiveWeapon) {
					bonusAmount = pa.Settings.SniperKillBonus * 1.5 // AWP kills get even higher priority
					LogInfo("🎯 AWP KILL by %s - immediate priority!", p.Name)
				} else {
					LogInfo("🎯 SNIPER KILL by %s - high priority!", p.Name)
				}

				pa.sniperKillBonus[p.SteamID] = bonusAmount
				pa.sniperKillTime[p.SteamID] = Now()
				pa.TotalSniperKills++ // Increment stats counter
				sniperKillDetected = true
				sniperKillerID = p.SteamID
			} else if strings.Contains(p.ActiveWeapon, "grenade") || strings.Contains(p.ActiveWeapon, "molotov") || strings.Contains(p.ActiveWeapon, "incendiary") || p.ActiveWeapon == "weapon_c4" {
				// Grenade/molotov kills - do NOT prioritize (player likely far away)
				LogVerbose("[ANALYZER] 💥 Grenade/fire kill by %s - not prioritizing", p.Name)
			} else if !killFollowAllowed {
				LogVerbose("[ANALYZER] Kill by %s ignored for kill-follow (phase %s, %d counters changed)", p.Name, roundPhase, changedKillers)
			} else {
				// NEW: any-weapon kill. The kill counter tells us exactly who
				// killed (no guessing). A player who just killed is likely to
				// keep fighting (entry frags, multi-kills), so follow them.
				pa.killBonus[p.SteamID] = killFollowBonus
				pa.killTime[p.SteamID] = Now()
				killDetected = true
				killerID = p.SteamID
				LogInfo("💀 KILL by %s (%s) - following the killer", p.Name, p.ActiveWeapon)
			}
		}
	}

	// Clean up expired damage dealt bonuses (older than configured duration)
	for steamID, bonusTime := range pa.damageDealtTime {
		if since(bonusTime) > time.Duration(pa.Settings.DamageDealtDuration)*time.Second {
			delete(pa.damageDealtBonus, steamID)
			delete(pa.damageDealtTime, steamID)
		}
	}

	// HP/DAMAGE DETECTION: Track health changes and find likely damage dealers
	damageDealtDetected := false
	damageDealerID := ""

	// Check if we have valid position data
	hasPositionData := false
	for _, p := range players {
		if p.Position.X != 0 || p.Position.Y != 0 || p.Position.Z != 0 {
			hasPositionData = true
			break
		}
	}

	for _, victim := range players {
		oldHealth, exists := pa.playerHealth[victim.SteamID]
		currentHealth := victim.Health

		// Update health tracking
		pa.playerHealth[victim.SteamID] = currentHealth

		// Detect significant HP loss (>20 damage)
		if exists && currentHealth < oldHealth {
			healthLost := oldHealth - currentHealth

			if healthLost >= pa.Settings.MinHealthLoss {
				// Only proceed with damage dealer identification if we have position data
				if !hasPositionData {
					LogVerbose("[ANALYZER] ⚠️  No position data - cannot identify damage dealer")
					continue
				}

				// Verify victim has valid position
				if victim.Position.X == 0 && victim.Position.Y == 0 && victim.Position.Z == 0 {
					LogVerbose("[ANALYZER] ⚠️  Victim %s has no position data", victim.Name)
					continue
				}

				// Find nearby enemies with shooting weapons
				var nearestAttacker *PlayerInfo
				nearestDistance := 99999.0
				// NEW: prefer the enemy whose crosshair is on the victim; the
				// nearest enemy is often behind a wall and not the shooter.
				var aimingAttacker *PlayerInfo
				aimingDistance := 0.0
				bestAim := 0.9 // ~25° minimum alignment

				for _, attacker := range players {
					// Skip same team
					if attacker.Team == victim.Team {
						continue
					}

					// Verify attacker has valid position
					if attacker.Position.X == 0 && attacker.Position.Y == 0 && attacker.Position.Z == 0 {
						continue
					}

					// Check if attacker has a shooting weapon (not grenade/knife)
					if !isShootingWeapon(attacker.ActiveWeapon) {
						continue
					}

					// Calculate distance (3D for accuracy)
					distance := CalculateDistance3D(victim.Position, attacker.Position)

					// Only consider enemies within reasonable range
					// Close range: Very likely
					// Mid range: Likely
					// Far range: Only if sniper rifle
					maxRange := pa.Settings.MaxDamageDistNormal
					if isSniperRifle(attacker.ActiveWeapon) {
						maxRange = pa.Settings.MaxDamageDistSniper // Snipers can damage from far away
					}

					if distance <= maxRange && distance < nearestDistance {
						nearestDistance = distance
						a := attacker
						nearestAttacker = &a
					}
					if c, ok := aimCos(attacker, victim); ok && distance <= maxRange && c > bestAim {
						bestAim = c
						a := attacker
						aimingAttacker = &a
						aimingDistance = distance
					}
				}
				if aimingAttacker != nil {
					nearestAttacker = aimingAttacker
					nearestDistance = aimingDistance
				}

				// If we found a likely attacker
				if nearestAttacker != nil {
					LogInfo("🎯 DAMAGE DEALT: %s likely hit %s (-%d HP, distance: %.0f)",
						nearestAttacker.Name, victim.Name, healthLost, nearestDistance)

					// Give damage dealer temporary bonus
					pa.damageDealtBonus[nearestAttacker.SteamID] = pa.Settings.DamageDealtBonus
					pa.damageDealtTime[nearestAttacker.SteamID] = Now()
					pa.TotalDamageDetections++ // Increment stats counter
					damageDealtDetected = true
					damageDealerID = nearestAttacker.SteamID
				}
			}
		}
	}

	// FIX: damage-dealer switches used to be blocked by the normal 2s rate limit,
	// so the camera often arrived after the kill. Let them bypass it, but keep a
	// short minimum gap so several hits in a row don't make the camera flicker.
	// PREDICTIVE MODE: during live rounds a trained model picks the player most
	// likely to get the next kill. The old reactive switches are turned off
	// unless Predict.EventSwitches is set.
	predictive := roundPhase == "live" && pa.Model != nil && pa.Predict.Enabled && hasPositionData
	if predictive && !pa.Predict.EventSwitches {
		upsetVictoryOccurred = false
		sniperKillDetected = false
		damageDealtDetected = false
		killDetected = false
	}

	const minDamageSwitchGap = 750 * time.Millisecond
	damageBypass := damageDealtDetected && damageDealerID != pa.currentSpectatedID &&
		since(pa.lastSwitchTime) >= minDamageSwitchGap

	// NEW: follow a killer immediately if they still have enemies around
	killSwitch := false
	if killDetected && killerID != pa.currentSpectatedID {
		for _, p := range players {
			if p.SteamID == killerID {
				killSwitch = hasEnemyNearby(p, players, pa.Settings.MaxEncounterDistance)
				break
			}
		}
	}
	killBypass := killSwitch && since(pa.lastSwitchTime) >= minEventSwitchGap

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
		} else if predictive || upsetVictoryOccurred || sniperKillDetected || damageBypass || killBypass {
			// Skip rate limiting for upset victory, sniper kill or damage dealt - we want immediate switch
			if upsetVictoryOccurred {
				LogVerbose("[ANALYZER] Upset victory detected - bypassing rate limit for immediate switch")
			}
			if sniperKillDetected {
				LogVerbose("[ANALYZER] Sniper kill detected - bypassing rate limit for immediate switch to killer")
			}
			if damageBypass {
				LogVerbose("[ANALYZER] Damage dealt detected - bypassing rate limit for immediate switch to shooter")
			}
			if killBypass {
				LogVerbose("[ANALYZER] Kill detected - bypassing rate limit for immediate switch to killer")
			}
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
			timeOnCurrentPlayer := since(pa.currentPlayerSwitchTime).Seconds()
			if timeOnCurrentPlayer >= maxTimeOnPlayer {
				// Force immediate switch by bypassing rate limit (don't log yet, wait until we actually switch)
			} else if since(pa.lastSwitchTime) < switchInterval {
				// Normal rate limiting for alive player
				timeSinceLastSwitch := since(pa.lastSwitchTime).Seconds()
				LogVerbose("[ANALYZER] Rate limiting active - waited %.1fs of %.0fs", timeSinceLastSwitch, switchInterval.Seconds())
				return ""
			}
		}
	} else {
		// No one spectated yet, check normal rate limit
		// BUT: Skip if upset victory or sniper kill occurred
		if !predictive && !upsetVictoryOccurred && !sniperKillDetected && !damageBypass && !killBypass {
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

			if since(pa.lastSwitchTime) < switchInterval {
				timeSinceLastSwitch := since(pa.lastSwitchTime).Seconds()
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
			pa.lastSwitchTime = Now()
			pa.currentSpectatedID = upsetWinnerID
			pa.currentPlayerSwitchTime = Now()
			return upsetWinnerID
		}
	}

	// PRIORITY: If sniper kill just occurred, switch to killer immediately
	// BUT only if it makes sense (Option A + C):
	// - Don't interrupt if current player has AWP/Sniper
	// - Only switch if killer has AWP (C) OR current player has no important action
	if sniperKillDetected && sniperKillerID != "" {
		// Verify killer is still alive
		killerAlive := false
		var killerWeapon string
		for _, p := range players {
			if p.SteamID == sniperKillerID {
				killerAlive = true
				killerWeapon = p.ActiveWeapon
				break
			}
		}

		if killerAlive && sniperKillerID != pa.currentSpectatedID {
			// Check current player's weapon
			currentPlayerHasSniper := false
			for _, p := range players {
				if p.SteamID == pa.currentSpectatedID {
					currentPlayerHasSniper = isSniperRifle(p.ActiveWeapon)
					break
				}
			}

			// Decision logic:
			// 1. If current player has AWP/Sniper → DON'T switch (keep watching sniper action)
			// 2. If killer has AWP → DO switch (AWP kills are high priority)
			// 3. Otherwise → DON'T switch immediately (use normal encounter logic)
			shouldSwitch := false
			reason := ""

			if currentPlayerHasSniper {
				// Don't interrupt sniper action
				LogVerbose("[ANALYZER] ⚠️  Not switching - current player has sniper weapon")
			} else if isAWP(killerWeapon) {
				// AWP kill always switches (unless watching AWP)
				shouldSwitch = true
				reason = "AWP kill detected"
			} else if isSniperRifle(killerWeapon) {
				// Scout kill only switches if not busy
				// We'll let the bonus handle this in encounter logic
				LogVerbose("[ANALYZER] Scout kill detected - using bonus instead of immediate switch")
			}

			if shouldSwitch {
				LogInfo("🎯 Switching to sniper killer immediately! (%s)", reason)
				pa.lastSwitchTime = Now()
				pa.currentSpectatedID = sniperKillerID
				pa.currentPlayerSwitchTime = Now()
				return sniperKillerID
			}
		}
	}

	// PRIORITY: If someone just got a kill and still has enemies around, follow them
	if killSwitch {
		LogInfo("💀 Switching to killer immediately!")
		pa.lastSwitchTime = Now()
		pa.previousSpectatedID = pa.currentSpectatedID
		pa.currentSpectatedID = killerID
		pa.currentPlayerSwitchTime = Now()
		return killerID
	}

	// PRIORITY: If damage was dealt, switch to damage dealer immediately
	if damageDealtDetected && damageDealerID != "" {
		// Verify damage dealer is still alive
		dealerAlive := false
		for _, p := range players {
			if p.SteamID == damageDealerID {
				dealerAlive = true
				break
			}
		}

		if dealerAlive && damageDealerID != pa.currentSpectatedID && pa.currentInAction(actionProtectTime) {
			LogVerbose("[ANALYZER] Damage elsewhere, but current player is in action - staying")
		} else if dealerAlive && damageDealerID != pa.currentSpectatedID {
			LogInfo("💥 Switching to damage dealer immediately!")
			pa.lastSwitchTime = Now()
			pa.currentSpectatedID = damageDealerID
			pa.currentPlayerSwitchTime = Now()
			return damageDealerID
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
			hasAim := false
			for _, p := range players {
				if p.HasForward {
					hasAim = true
					break
				}
			}
			if hasAim {
				LogInfo("✅ Aim direction (forward) is available - aim-based detection enabled")
			} else {
				LogInfo("⚠️  No aim direction (forward) in GSI - aim-based detection disabled")
			}
			pa.positionWarningShown = true
		}
	}

	if predictive {
		choice := pa.predictiveChoice(players)
		if choice != "" && choice != pa.currentSpectatedID {
			LogInfo("🔮 PREDICT: switching to %s (P=%.2f)", getPlayerNameByID(choice, players), pa.probs[choice])
			pa.lastSwitchTime = Now()
			pa.previousSpectatedID = pa.currentSpectatedID
			pa.currentSpectatedID = choice
			pa.currentPlayerSwitchTime = Now()
			return choice
		}
		return ""
	}

	// Find encounters
	encounters := pa.PredictEncounters(players)
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
		timeOnCurrentPlayer := since(pa.currentPlayerSwitchTime).Seconds()
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

		// Apply sniper kill bonus
		for i := range encounters {
			bonus1 := pa.sniperKillBonus[encounters[i].Player1.SteamID]
			bonus2 := pa.sniperKillBonus[encounters[i].Player2.SteamID]

			if bonus1 > 0 {
				encounters[i].Priority += bonus1
				weaponType := "Sniper"
				if isAWP(encounters[i].Player1.ActiveWeapon) {
					weaponType = "AWP"
				}
				LogVerbose("[ANALYZER] 🎯 %s kill bonus for %s: +%.0f",
					weaponType, encounters[i].Player1.Name, bonus1)
			}
			if bonus2 > 0 {
				encounters[i].Priority += bonus2
				weaponType := "Sniper"
				if isAWP(encounters[i].Player2.ActiveWeapon) {
					weaponType = "AWP"
				}
				LogVerbose("[ANALYZER] 🎯 %s kill bonus for %s: +%.0f",
					weaponType, encounters[i].Player2.Name, bonus2)
			}
		}

		// Apply damage dealt bonus
		for i := range encounters {
			bonus1 := pa.damageDealtBonus[encounters[i].Player1.SteamID]
			bonus2 := pa.damageDealtBonus[encounters[i].Player2.SteamID]

			if bonus1 > 0 {
				encounters[i].Priority += bonus1
				LogVerbose("[ANALYZER] 💥 Damage dealt bonus for %s: +%.0f",
					encounters[i].Player1.Name, bonus1)
			}
			if bonus2 > 0 {
				encounters[i].Priority += bonus2
				LogVerbose("[ANALYZER] 💥 Damage dealt bonus for %s: +%.0f",
					encounters[i].Player2.Name, bonus2)
			}
		}

		// Apply kill-follow bonus (any weapon)
		for i := range encounters {
			if b := pa.killBonus[encounters[i].Player1.SteamID]; b > 0 {
				encounters[i].Priority += b
				LogVerbose("[ANALYZER] 💀 Kill bonus for %s: +%.0f", encounters[i].Player1.Name, b)
			}
			if b := pa.killBonus[encounters[i].Player2.SteamID]; b > 0 {
				encounters[i].Priority += b
				LogVerbose("[ANALYZER] 💀 Kill bonus for %s: +%.0f", encounters[i].Player2.Name, b)
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
			// Choose based on equipment/kills/health/weapon
			p1Score := float64(encounter.Player1.EquipmentValue)/100 + float64(encounter.Player1.Kills)*10
			p2Score := float64(encounter.Player2.EquipmentValue)/100 + float64(encounter.Player2.Kills)*10

			// Bonus for more health
			if encounter.Player1.Health > 50 {
				p1Score += 5
			}
			if encounter.Player2.Health > 50 {
				p2Score += 5
			}

			// Bonus for sniper weapons (AWP > Scout)
			p1Score += pa.getSniperWeaponBonus(encounter.Player1.ActiveWeapon)
			p2Score += pa.getSniperWeaponBonus(encounter.Player2.ActiveWeapon)

			// Bonus for rifles (AK-47 > M4s)
			p1Score += pa.getRifleWeaponBonus(encounter.Player1.ActiveWeapon)
			p2Score += pa.getRifleWeaponBonus(encounter.Player2.ActiveWeapon)

			// NEW: prefer the POV of whoever is aiming at the other player
			if isAimingAt(encounter.Player1, encounter.Player2) {
				p1Score += 25
			}
			if isAimingAt(encounter.Player2, encounter.Player1) {
				p2Score += 25
			}

			// NEW: prefer a player who just got a kill
			p1Score += pa.killBonus[encounter.Player1.SteamID] / 4
			p2Score += pa.killBonus[encounter.Player2.SteamID] / 4

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

		// ANTI PING-PONG: an encounter-based switch away from a living player
		// must (1) wait a minimum dwell time and (2) clearly beat the current
		// player's own best encounter. Going straight back to the player we
		// just left needs an even bigger margin.
		if bestPlayerID != pa.currentSpectatedID && pa.currentSpectatedID != "" && roundPhase == "live" {
			timeOnCurrent := since(pa.currentPlayerSwitchTime)
			if timeOnCurrent < pa.maxStickyTime {
				currentBest := 0.0
				for _, e := range encounters {
					if e.Player1.SteamID == pa.currentSpectatedID || e.Player2.SteamID == pa.currentSpectatedID {
						if e.Priority > currentBest {
							currentBest = e.Priority
						}
					}
				}
				margin := switchMargin
				if bestPlayerID == pa.previousSpectatedID && since(pa.lastSwitchTime) < returnWindow {
					margin = returnMargin
				}
				if timeOnCurrent < minDwellTime {
					LogVerbose("[ANALYZER] ⏸️  Holding camera (%.1fs < %.0fs min dwell)", timeOnCurrent.Seconds(), minDwellTime.Seconds())
					return ""
				}
				if currentBest > 0 && encounter.Priority < currentBest+margin {
					LogVerbose("[ANALYZER] ⏸️  Holding camera: %.1f doesn't beat current %.1f by %.0f", encounter.Priority, currentBest, margin)
					return ""
				}
			}
		}

		// Only switch if it's a different player
		if bestPlayerID != pa.currentSpectatedID {
			pa.lastSwitchTime = Now()
			pa.previousSpectatedID = pa.currentSpectatedID
			pa.currentSpectatedID = bestPlayerID
			pa.currentPlayerSwitchTime = Now() // Reset sticky timer
			return bestPlayerID
		} else {
			// Already on this player - check if we need to force switch in warmup/freezetime/timeout
			timeOnCurrent := since(pa.currentPlayerSwitchTime).Seconds()
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
					pa.lastSwitchTime = Now()
					pa.previousSpectatedID = pa.currentSpectatedID
					pa.currentSpectatedID = alternativeID
					pa.currentPlayerSwitchTime = Now()
					return alternativeID
				} else {
					// Normal game: reset timer with offset to retry later
					LogVerbose("[ANALYZER] Max time exceeded but already on best player - will retry in 3s")
					pa.currentPlayerSwitchTime = Now().Add(-time.Duration(maxTime-3) * time.Second)
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

		// Apply sniper kill bonus even in fallback mode
		for i := range players {
			if bonus, exists := pa.sniperKillBonus[players[i].SteamID]; exists {
				// Add to equipment value for sorting (scaled appropriately)
				players[i].EquipmentValue += int(bonus * 150) // Sniper kills get even higher weight
				weaponType := "Sniper"
				if isAWP(players[i].ActiveWeapon) {
					weaponType = "AWP"
				}
				LogVerbose("[ANALYZER] 🎯 %s kill bonus for %s in fallback mode", weaponType, players[i].Name)
			}
		}

		// Apply damage dealt bonus even in fallback mode
		for i := range players {
			if bonus, exists := pa.damageDealtBonus[players[i].SteamID]; exists {
				// Add to equipment value for sorting (scaled appropriately)
				players[i].EquipmentValue += int(bonus * 100) // Similar to combat winner bonus
				LogVerbose("[ANALYZER] 💥 Damage dealt bonus for %s in fallback mode", players[i].Name)
			}
		}

		// Apply kill-follow bonus even in fallback mode
		for i := range players {
			if bonus, exists := pa.killBonus[players[i].SteamID]; exists {
				players[i].EquipmentValue += int(bonus * 100)
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
			// In freezetime/warmup: ONLY switch on maxTime (every 7.5s), not based on "best player"
			// In other phases: Switch to best player immediately (after rate limit)
			if roundPhase == "warmup" || roundPhase == "freezetime" {
				// Force switch mode: only switch every 7.5 seconds for consistent pacing
				timeOnCurrent := since(pa.currentPlayerSwitchTime).Seconds()
				if timeOnCurrent >= 7.5 {
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
						LogInfo("⏱️  Max time (7.5s) exceeded in %s - switching for variety", roundPhase)
						pa.lastSwitchTime = Now()
						pa.previousSpectatedID = pa.currentSpectatedID
						pa.currentSpectatedID = nextPlayerID
						pa.currentPlayerSwitchTime = Now()
						return nextPlayerID
					}
				}
				// Not time to switch yet in freezetime/warmup
				return ""
			} else {
				// Normal game phases (live/timeout): wait at least 5 seconds before switching in fallback
				timeOnCurrent := since(pa.currentPlayerSwitchTime).Seconds()
				var minTime float64

				if roundPhase == "timeout" {
					minTime = 10.0
				} else {
					minTime = 5.0 // Live phase: also 5 seconds minimum
				}

				// Only switch if we've been on current player long enough
				if timeOnCurrent >= minTime {
					bestPlayerID := players[0].SteamID
					if bestPlayerID != pa.currentSpectatedID && len(players) > 1 {
						// Switch to best player after minimum time
						LogInfo("⏱️  Fallback: %.0fs elapsed, switching to better player", timeOnCurrent)
						pa.lastSwitchTime = Now()
						pa.previousSpectatedID = pa.currentSpectatedID
						pa.currentSpectatedID = bestPlayerID
						pa.currentPlayerSwitchTime = Now()
						return bestPlayerID
					}
				}
				// Not enough time elapsed, stay with current player
				return ""
			}
		}
	}

	return ""
}

// SyncCurrentPlayer updates the analyzer's internal state to match the actual spectated player from GSI
// This is called when we detect that CS is spectating a different player than we think
func (pa *PlayerAnalyzer) SyncCurrentPlayer(steamID string) {
	pa.currentSpectatedID = steamID
	pa.currentPlayerSwitchTime = Now() // Reset sticky timer to prevent immediate switch
	LogVerbose("[ANALYZER] Synced to actual spectated player: %s", steamID)
}

// RecentlySwitched reports whether the analyzer issued a camera switch within d.
// Used to ignore stale GSI "spectated player" data right after a switch.
func (pa *PlayerAnalyzer) RecentlySwitched(d time.Duration) bool {
	return since(pa.lastSwitchTime) < d
}

// getActiveWeapon returns the name of the player's active weapon from GSI data
// ("weapons": {"weapon_0": {"name": "weapon_ak47", "state": "active"}, ...})
func getActiveWeapon(playerMap map[string]interface{}) string {
	weapons, ok := playerMap["weapons"].(map[string]interface{})
	if !ok {
		return ""
	}
	for _, weaponData := range weapons {
		weaponMap, ok := weaponData.(map[string]interface{})
		if !ok {
			continue
		}
		if state, _ := weaponMap["state"].(string); state == "active" {
			name, _ := weaponMap["name"].(string)
			return name
		}
	}
	return ""
}

// Helper functions for safe type assertions
func GetStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return "Unknown"
}

func GetIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func GetFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0.0
}

func GetBoolValue(m map[string]interface{}, key string) bool {
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

package director

// Settings represents all configurable parameters for the Auto Director
type Settings struct {
	// Camera Priority Bonuses
	AWPBonus          float64 `json:"awp_bonus"`           // Bonus points for AWP sniper rifle
	ScoutBonus        float64 `json:"scout_bonus"`         // Bonus points for Scout sniper rifle
	AK47Bonus         float64 `json:"ak47_bonus"`          // Bonus points for AK-47 rifle
	DamageDealtBonus  float64 `json:"damage_dealt_bonus"`  // Bonus for dealing damage
	UpsetVictoryBonus float64 `json:"upset_victory_bonus"` // Bonus for upset victories
	SniperKillBonus   float64 `json:"sniper_kill_bonus"`   // Bonus for sniper kills

	// Duration Settings (in seconds)
	DamageDealtDuration  int `json:"damage_dealt_duration"`  // How long damage dealt bonus lasts
	UpsetVictoryDuration int `json:"upset_victory_duration"` // How long upset victory bonus lasts
	SniperKillDuration   int `json:"sniper_kill_duration"`   // How long sniper kill bonus lasts

	// Distance Limits (in game units)
	MaxEncounterDistance float64 `json:"max_encounter_distance"` // Max distance to consider an encounter
	MaxDamageDistNormal  float64 `json:"max_damage_dist_normal"` // Max distance for damage detection (normal weapons)
	MaxDamageDistSniper  float64 `json:"max_damage_dist_sniper"` // Max distance for damage detection (snipers)

	// Advanced Settings
	MinHealthLoss         int     `json:"min_health_loss"`         // Minimum HP loss to trigger damage detection
	VerticalDiffThreshold float64 `json:"vertical_diff_threshold"` // Max vertical distance for encounters
}

// NewDefaultSettings returns settings with default values
func NewDefaultSettings() *Settings {
	return &Settings{
		// Camera Priority Bonuses
		AWPBonus:          30.0,
		ScoutBonus:        15.0,
		AK47Bonus:         5.0,
		DamageDealtBonus:  40.0,
		UpsetVictoryBonus: 150.0,
		SniperKillBonus:   100.0,

		// Duration Settings
		DamageDealtDuration:  5,
		UpsetVictoryDuration: 8,
		SniperKillDuration:   5,

		// Distance Limits
		MaxEncounterDistance: 1500.0,
		MaxDamageDistNormal:  1500.0,
		MaxDamageDistSniper:  3000.0,

		// Advanced Settings
		MinHealthLoss:         20,
		VerticalDiffThreshold: 250.0,
	}
}

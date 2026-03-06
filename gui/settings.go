package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Secrets holds sensitive data like API keys
type Secrets struct {
	FaceitAPIKey string `json:"faceit_api_key"` // FACEIT API Key
}

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

// GetSettings returns the current settings
func (a *App) GetSettings() *Settings {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// If settings not loaded yet, return defaults
	if a.autoDirector == nil || a.autoDirector.playerAnalyzer == nil {
		return NewDefaultSettings()
	}

	return a.autoDirector.playerAnalyzer.Settings
}

// SaveSettings saves settings to file and updates the analyzer
func (a *App) SaveSettings(settings *Settings) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Update analyzer settings
	if a.autoDirector != nil && a.autoDirector.playerAnalyzer != nil {
		a.autoDirector.playerAnalyzer.Settings = settings
	}

	// Save to file
	return saveSettingsToFile(settings)
}

// LoadSettings loads settings from file or returns defaults
func LoadSettings() *Settings {
	settings, err := loadSettingsFromFile()
	if err != nil {
		LogInfo("Using default settings (no saved config found)")
		return NewDefaultSettings()
	}
	return settings
}

// saveSettingsToFile saves settings to JSON file
func saveSettingsToFile(settings *Settings) error {
	configDir := filepath.Join(".", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(configDir, "settings.json")
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// loadSettingsFromFile loads settings from JSON file
func loadSettingsFromFile() (*Settings, error) {
	filePath := filepath.Join(".", "config", "settings.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// ResetSettings resets settings to defaults
func (a *App) ResetSettings() error {
	return a.SaveSettings(NewDefaultSettings())
}

// ========== Secrets Management ==========

// LoadSecrets loads secrets from file or returns empty struct
func LoadSecrets() *Secrets {
	secrets, err := loadSecretsFromFile()
	if err != nil {
		LogInfo("No secrets file found. Please create config/secrets.json")
		return &Secrets{}
	}
	return secrets
}

// loadSecretsFromFile loads secrets from JSON file
func loadSecretsFromFile() (*Secrets, error) {
	filePath := filepath.Join(".", "config", "secrets.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var secrets Secrets
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, err
	}

	return &secrets, nil
}

// saveSecretsToFile saves secrets to JSON file
func saveSecretsToFile(secrets *Secrets) error {
	configDir := filepath.Join(".", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(configDir, "secrets.json")
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// GetSecrets returns current secrets
func (a *App) GetSecrets() *Secrets {
	return LoadSecrets()
}

// SaveSecrets saves secrets to file
func (a *App) SaveSecrets(secrets *Secrets) error {
	return saveSecretsToFile(secrets)
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// getProjectRoot returns the project root directory
// If running from gui/, it goes one level up
func getProjectRoot() string {
	// Check if we're in the gui/ subdirectory
	if _, err := os.Stat("wails.json"); err == nil {
		// We're in gui/, go one level up
		return ".."
	}
	// We're already in project root
	return "."
}

// getConfigDir returns the directory where settings and logs should be stored.
// It uses %APPDATA% on Windows to avoid permission issues in Program Files.
func getConfigDir() string {
	// If we are in development mode (wails.json exists), use project root
	if _, err := os.Stat("wails.json"); err == nil {
		return "."
	}

	// Check if the current executable directory is writable (e.g., Admin install or Portable mode)
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		// Attempt to create a temporary test file to verify write permissions
		testFile := filepath.Join(exeDir, ".permission_test")
		if f, err := os.Create(testFile); err == nil {
			f.Close()
			os.Remove(testFile)
			return exeDir
		}
	}

	// Use OS specific config directory (e.g., AppData/Roaming on Windows)
	dir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	appConfigDir := filepath.Join(dir, "cs2-better-autodirector")
	os.MkdirAll(appConfigDir, 0755)
	return appConfigDir
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
	configDir := filepath.Join(getConfigDir(), "config")
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
	filePath := filepath.Join(getConfigDir(), "config", "settings.json")
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

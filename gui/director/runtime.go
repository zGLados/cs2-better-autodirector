package director

import "time"

// Now is the clock used by the director. The app uses the real clock; the
// demo replay tool replaces it with the demo's game time.
var Now = time.Now

func since(t time.Time) time.Duration { return Now().Sub(t) }

// LogLevel mirrors the app's log level (0=info, 1=verbose, 2=trace)
var LogLevel = 0

// Logging hooks, wired to the app's logger by package main. No-op by default.
var (
	LogInfo    = func(format string, args ...interface{}) {}
	LogVerbose = func(format string, args ...interface{}) {}
)

// LoadSettings returns the settings to use for a new analyzer. The app
// replaces it with its file-based loader.
var LoadSettings = func() *Settings { return NewDefaultSettings() }

// CurrentSpectatedID returns who the director believes is on camera
func (pa *PlayerAnalyzer) CurrentSpectatedID() string { return pa.currentSpectatedID }

// PredictionDisabled makes new analyzers use the legacy encounter logic
// (app flag -legacy).
var PredictionDisabled = false

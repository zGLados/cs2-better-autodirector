package main

import "cs2-autodirector-gui/director"

// The camera-decision logic lives in package director so it can also be run
// offline by the demo replay tool. These aliases keep the app code unchanged.
type (
	PlayerInfo     = director.PlayerInfo
	Position       = director.Position
	Encounter      = director.Encounter
	Settings       = director.Settings
	PlayerAnalyzer = director.PlayerAnalyzer
)

var (
	NewPlayerAnalyzer   = director.NewPlayerAnalyzer
	NewDefaultSettings  = director.NewDefaultSettings
	parsePositionString = director.ParsePositionString
	getStringValue      = director.GetStringValue
	getIntValue         = director.GetIntValue
)

func init() {
	director.LogInfo = LogInfo
	director.LogVerbose = LogVerbose
	director.LoadSettings = LoadSettings
}

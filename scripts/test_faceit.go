package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// Simple test program for FACEIT API integration
func main() {
	// Command line flags
	apiKey := flag.String("key", "", "FACEIT API Key")
	matchURL := flag.String("url", "", "FACEIT Match Room URL")
	flag.Parse()

	fmt.Println("====================================")
	fmt.Println("🎯 FACEIT API Test Tool")
	fmt.Println("====================================")
	fmt.Println()

	// Load API key from secrets.json if not provided
	if *apiKey == "" {
		fmt.Println("📂 Loading API key from config/secrets.json...")
		secrets := LoadSecrets()
		if secrets.FaceitAPIKey == "" || secrets.FaceitAPIKey == "YOUR_FACEIT_API_KEY_HERE" {
			fmt.Println("❌ Error: No API key found!")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  1. Add API key to config/secrets.json")
			fmt.Println("  2. Use -key flag: go run test_faceit.go -key YOUR_KEY -url MATCH_URL")
			os.Exit(1)
		}
		*apiKey = secrets.FaceitAPIKey
		fmt.Println("✅ API key loaded from secrets.json")
	} else {
		fmt.Println("✅ Using API key from command line")
	}
	fmt.Println()

	// Check for match URL
	if *matchURL == "" {
		fmt.Println("❌ Error: No match URL provided!")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run gui/test_faceit.go -url \"https://www.faceit.com/en/cs2/room/1-...\"")
		fmt.Println()
		fmt.Println("Or with custom API key:")
		fmt.Println("  go run gui/test_faceit.go -key YOUR_KEY -url \"https://www.faceit.com/en/cs2/room/1-...\"")
		os.Exit(1)
	}

	fmt.Printf("🔗 Match URL: %s\n", *matchURL)
	fmt.Println()

	// Initialize FACEIT client
	fmt.Println("🚀 Initializing FACEIT client...")
	client := NewFaceitClient(*apiKey)
	fmt.Println("✅ Client initialized")
	fmt.Println()

	// Fetch match data
	fmt.Println("📡 Fetching match data from FACEIT API...")
	fmt.Println("------------------------------------")
	matchData, err := client.GetMatchDataFromURL(*matchURL)
	if err != nil {
		fmt.Printf("❌ Error fetching match data: %v\n", err)
		fmt.Println()
		fmt.Println("Troubleshooting:")
		fmt.Println("  - Check if API key is valid")
		fmt.Println("  - Verify match URL format")
		fmt.Println("  - Ensure match ID is correct")
		os.Exit(1)
	}

	fmt.Println("✅ Match data fetched successfully!")
	fmt.Println()

	// Display match information
	fmt.Println("====================================")
	fmt.Println("📊 Match Information")
	fmt.Println("====================================")
	fmt.Println()

	fmt.Printf("🆔 Match ID:      %s\n", matchData.MatchID)
	fmt.Printf("📌 Status:        %s\n", matchData.Status)
	fmt.Printf("🏆 Competition:   %s\n", matchData.Competition)
	fmt.Printf("🎮 Game:          %s\n", matchData.MatchType)
	fmt.Println()

	fmt.Println("====================================")
	fmt.Println("🔵 Team 1")
	fmt.Println("====================================")
	fmt.Printf("Name:    %s\n", matchData.Team1.Name)
	fmt.Printf("Team ID: %s\n", matchData.Team1.TeamID)
	fmt.Printf("Score:   %d\n", matchData.Team1.Score)
	fmt.Printf("Logo:    %s\n", matchData.Team1.Logo)
	fmt.Println()

	fmt.Println("====================================")
	fmt.Println("🔴 Team 2")
	fmt.Println("====================================")
	fmt.Printf("Name:    %s\n", matchData.Team2.Name)
	fmt.Printf("Team ID: %s\n", matchData.Team2.TeamID)
	fmt.Printf("Score:   %d\n", matchData.Team2.Score)
	fmt.Printf("Logo:    %s\n", matchData.Team2.Logo)
	fmt.Println()

	fmt.Println("====================================")
	fmt.Println("📺 GOTV Connection")
	fmt.Println("====================================")
	if matchData.GotvLink != "" {
		fmt.Printf("Link:     %s\n", matchData.GotvLink)
		fmt.Printf("Command:  %s\n", client.FormatGotvLink(matchData.GotvLink))
		fmt.Println()
		if matchData.GotvLink == "Not available (match might not be live yet)" {
			fmt.Println("ℹ️  Note: GOTV link not available yet")
			fmt.Println("   This is normal for matches that haven't started")
		}
	} else {
		fmt.Println("❌ No GOTV link available")
	}
	fmt.Println()

	// Output JSON for debugging
	fmt.Println("====================================")
	fmt.Println("🔍 Full JSON Response (for debugging)")
	fmt.Println("====================================")
	jsonData, err := json.MarshalIndent(matchData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
	} else {
		fmt.Println(string(jsonData))
	}
	fmt.Println()

	fmt.Println("====================================")
	fmt.Println("✅ Test completed successfully!")
	fmt.Println("====================================")
}

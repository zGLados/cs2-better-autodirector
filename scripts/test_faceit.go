package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Secrets holds sensitive data like API keys
type Secrets struct {
	FaceitAPIKey string `json:"faceit_api_key"`
}

// FaceitClient handles FACEIT API interactions
type FaceitClient struct {
	apiKey     string
	httpClient *http.Client
}

// FaceitMatchData holds all relevant match information
type FaceitMatchData struct {
	MatchID     string         `json:"match_id"`
	GotvLink    string         `json:"gotv_link"`
	Team1       FaceitTeamData `json:"team1"`
	Team2       FaceitTeamData `json:"team2"`
	Status      string         `json:"status"`
	MatchType   string         `json:"match_type"`
	Competition string         `json:"competition"`
	StartedAt   int64          `json:"started_at"`
}

// FaceitTeamData holds team information
type FaceitTeamData struct {
	Name   string `json:"name"`
	Logo   string `json:"logo"`
	TeamID string `json:"team_id"`
	Score  int    `json:"score"`
}

// LoadSecrets loads secrets from file
func LoadSecrets() *Secrets {
	filePath := filepath.Join("..", "config", "secrets.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return &Secrets{}
	}

	var secrets Secrets
	if err := json.Unmarshal(data, &secrets); err != nil {
		return &Secrets{}
	}

	return &secrets
}

// NewFaceitClient creates a new FACEIT API client
func NewFaceitClient(apiKey string) *FaceitClient {
	return &FaceitClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetMatchDataFromURL extracts match data from a FACEIT match room URL
func (fc *FaceitClient) GetMatchDataFromURL(matchRoomURL string) (*FaceitMatchData, error) {
	matchID, err := fc.extractMatchID(matchRoomURL)
	if err != nil {
		return nil, err
	}

	return fc.getMatchDetails(matchID)
}

// extractMatchID extracts the match ID from a FACEIT match room URL
func (fc *FaceitClient) extractMatchID(url string) (string, error) {
	re := regexp.MustCompile(`/room/(1-[a-f0-9-]+)`)
	matches := re.FindStringSubmatch(url)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract match ID from URL: %s", url)
	}

	return matches[1], nil
}

// getMatchDetails fetches match details from FACEIT API
func (fc *FaceitClient) getMatchDetails(matchID string) (*FaceitMatchData, error) {
	url := fmt.Sprintf("https://open.faceit.com/data/v4/matches/%s", matchID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+fc.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := fc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FACEIT API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		MatchID         string `json:"match_id"`
		Status          string `json:"status"`
		StartedAt       int64  `json:"started_at"`
		CompetitionName string `json:"competition_name"`
		Game            string `json:"game"`
		DemoURL         []string `json:"demo_url"`
		Teams           map[string]struct {
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		} `json:"teams"`
		Results struct {
			Score map[string]int `json:"score"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	matchData := &FaceitMatchData{
		MatchID:     apiResponse.MatchID,
		Status:      apiResponse.Status,
		StartedAt:   apiResponse.StartedAt,
		Competition: apiResponse.CompetitionName,
		MatchType:   apiResponse.Game,
	}

	teamIndex := 0
	for faction, team := range apiResponse.Teams {
		teamData := FaceitTeamData{
			Name:   team.Name,
			Logo:   team.Avatar,
			TeamID: faction,
		}

		if score, ok := apiResponse.Results.Score[faction]; ok {
			teamData.Score = score
		}

		if teamIndex == 0 {
			matchData.Team1 = teamData
		} else {
			matchData.Team2 = teamData
		}
		teamIndex++
	}

	if len(apiResponse.DemoURL) > 0 {
		matchData.GotvLink = apiResponse.DemoURL[0]
	} else {
		matchData.GotvLink = "Not available (match might not be live yet)"
	}

	return matchData, nil
}

// FormatGotvLink formats the GOTV link for CS2 console
func (fc *FaceitClient) FormatGotvLink(gotvLink string) string {
	if strings.HasPrefix(gotvLink, "connect ") {
		return gotvLink
	}

	if strings.Contains(gotvLink, ":") {
		return fmt.Sprintf("connect %s", gotvLink)
	}

	return gotvLink
}

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

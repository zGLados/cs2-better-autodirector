package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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
	BestOf      int            `json:"best_of"` // BO1, BO3, BO5 etc
}

// FaceitTeamData holds team information
type FaceitTeamData struct {
	Name   string `json:"name"`
	Logo   string `json:"logo"`
	TeamID string `json:"team_id"`
	Score  int    `json:"score"`
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
	// Extract match ID from URL
	// Examples:
	// https://www.faceit.com/en/cs2/room/1-abc123-def456-ghi789
	// https://www.faceit.com/en/csgo/room/1-abc123-def456-ghi789
	matchID, err := fc.extractMatchID(matchRoomURL)
	if err != nil {
		return nil, err
	}

	LogInfo("Fetching FACEIT match data for ID: %s", matchID)

	// Get match details from FACEIT API
	matchData, err := fc.getMatchDetails(matchID)
	if err != nil {
		return nil, err
	}

	LogInfo("Successfully fetched match data: %s vs %s", matchData.Team1.Name, matchData.Team2.Name)

	return matchData, nil
}

// extractMatchID extracts the match ID from a FACEIT match room URL
func (fc *FaceitClient) extractMatchID(url string) (string, error) {
	// Pattern: /room/1-abc123-def456-ghi789
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

	// Parse the response
	var apiResponse struct {
		MatchID         string   `json:"match_id"`
		Status          string   `json:"status"`
		StartedAt       int64    `json:"started_at"`
		CompetitionName string   `json:"competition_name"`
		CompetitionType string   `json:"competition_type"` // tournament, hub, matchmaking, etc.
		OrganizerID     string   `json:"organizer_id"`     // FACEIT ID for official tournaments
		Game            string   `json:"game"`
		BestOf          int      `json:"best_of"` // BO1, BO3, BO5 etc
		DemoURL         []string `json:"demo_url"`
		Teams           map[string]struct {
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
			Roster []struct {
				PlayerID string `json:"player_id"`
				Nickname string `json:"nickname"`
			} `json:"roster"`
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
		BestOf:      apiResponse.BestOf,
	}

	// Extract team data (FACEIT API returns teams in "faction1" and "faction2")
	// Always assign faction1 to Team1 and faction2 to Team2 for stable ordering
	if team1, ok := apiResponse.Teams["faction1"]; ok {
		matchData.Team1 = FaceitTeamData{
			Name:   team1.Name,
			Logo:   team1.Avatar,
			TeamID: "faction1",
			Score:  apiResponse.Results.Score["faction1"],
		}
	}

	if team2, ok := apiResponse.Teams["faction2"]; ok {
		matchData.Team2 = FaceitTeamData{
			Name:   team2.Name,
			Logo:   team2.Avatar,
			TeamID: "faction2",
			Score:  apiResponse.Results.Score["faction2"],
		}
	}

	// Extract GOTV link from demo URLs
	if len(apiResponse.DemoURL) > 0 {
		// Demo URLs are typically provided after the match
		// For live matches, we need to construct or fetch the server info differently
		// For now, we'll store the first demo URL if available
		matchData.GotvLink = apiResponse.DemoURL[0]
	} else {
		// Try to get server info for live matches
		serverInfo, err := fc.getMatchServerInfo(matchID)
		if err != nil {
			LogDebug("Could not fetch server info: %v", err)

			// Check if this is a tournament/official match
			isTournament := fc.isTournamentMatch(apiResponse.CompetitionName, apiResponse.CompetitionType, apiResponse.OrganizerID)

			if isTournament {
				matchData.GotvLink = "Not available yet (check back when match is live)"
			} else {
				matchData.GotvLink = "Not available - Only public tournament matches provide GOTV access"
			}
		} else {
			matchData.GotvLink = serverInfo
		}
	}

	return matchData, nil
}

// getMatchServerInfo attempts to get the live server info for a match
func (fc *FaceitClient) getMatchServerInfo(matchID string) (string, error) {
	// FACEIT doesn't always expose GOTV server info directly via the public API
	// This would require either:
	// 1. Websocket connection to match updates
	// 2. Server info from a different endpoint
	// 3. Scraping the match room page

	// For now, we'll return a placeholder that indicates where to connect
	// In a production version, you might need to:
	// - Poll the match endpoint for server_id
	// - Use that to fetch server details
	// - Or fetch from the match room HTML page

	return "", fmt.Errorf("server info not available via API yet")
}

// isTournamentMatch checks if a match is an official tournament based on available indicators
func (fc *FaceitClient) isTournamentMatch(competitionName, competitionType, organizerID string) bool {
	// If there's an organizer ID, it's likely an official tournament
	if organizerID != "" {
		return true
	}

	// Check competition type
	tournamentTypes := []string{"tournament", "championship", "league", "qualifier", "major"}
	competitionTypeLower := strings.ToLower(competitionType)
	for _, tt := range tournamentTypes {
		if strings.Contains(competitionTypeLower, tt) {
			return true
		}
	}

	// Check competition name for tournament indicators
	if competitionName == "" {
		return false // No competition = regular match
	}

	competitionLower := strings.ToLower(competitionName)

	// These are typical non-tournament competition names
	nonTournamentNames := []string{"matchmaking", "5v5", "ranked", "unranked", "casual", "pug"}
	for _, nt := range nonTournamentNames {
		if strings.Contains(competitionLower, nt) {
			return false
		}
	}

	// Known tournament/league indicators
	tournamentIndicators := []string{
		"major", "qualifier", "rmr", "esl", "blast", "iem", "pgl",
		"championship", "league", "cup", "open", "fpl", "faceit pro",
		"esea", "esportal", "dreamhack", "eleague",
	}

	for _, indicator := range tournamentIndicators {
		if strings.Contains(competitionLower, indicator) {
			return true
		}
	}

	// If we have a proper competition name but no negative indicators, assume it's a tournament
	return len(competitionName) > 0
}

// FormatGotvLink formats the GOTV link for CS2 console
func (fc *FaceitClient) FormatGotvLink(gotvLink string) string {
	// If it's already a connect command or IP:port, return as is
	if strings.HasPrefix(gotvLink, "connect ") {
		return gotvLink
	}

	// If it contains "gotv" or is an IP:port
	if strings.Contains(gotvLink, ":") {
		return fmt.Sprintf("connect %s", gotvLink)
	}

	return gotvLink
}

// DownloadTeamLogo downloads a team logo from URL and saves it locally
func (fc *FaceitClient) DownloadTeamLogo(logoURL, teamName string) (string, error) {
	if logoURL == "" {
		return "", fmt.Errorf("no logo URL provided")
	}

	// Create temp directory for logos if it doesn't exist
	projectRoot := getProjectRoot()
	tempDir := filepath.Join(projectRoot, "temp", "logos")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download the logo
	resp, err := fc.httpClient.Get(logoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download logo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download logo (status %d)", resp.StatusCode)
	}

	// Determine file extension from URL or Content-Type
	ext := filepath.Ext(logoURL)
	if ext == "" {
		contentType := resp.Header.Get("Content-Type")
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/svg+xml":
			ext = ".svg"
		default:
			ext = ".png" // Default to PNG
		}
	}

	// Create safe filename from team name
	safeTeamName := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, teamName)

	filename := fmt.Sprintf("%s_%d%s", safeTeamName, time.Now().Unix(), ext)
	filePath := filepath.Join(tempDir, filename)

	// Save the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create logo file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save logo: %w", err)
	}

	LogInfo("Downloaded team logo: %s -> %s", logoURL, filePath)
	return filePath, nil
}

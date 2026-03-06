package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	MatchID      string         `json:"match_id"`
	GotvLink     string         `json:"gotv_link"`
	Team1        FaceitTeamData `json:"team1"`
	Team2        FaceitTeamData `json:"team2"`
	Status       string         `json:"status"`
	MatchType    string         `json:"match_type"`
	Competition  string         `json:"competition"`
	StartedAt    int64          `json:"started_at"`
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
		MatchID     string `json:"match_id"`
		Status      string `json:"status"`
		StartedAt   int64  `json:"started_at"`
		CompetitionName string `json:"competition_name"`
		Game        string `json:"game"`
		DemoURL     []string `json:"demo_url"`
		Teams       map[string]struct {
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
	}
	
	// Extract team data (FACEIT API returns teams in "faction1" and "faction2")
	teamIndex := 0
	for faction, team := range apiResponse.Teams {
		teamData := FaceitTeamData{
			Name:   team.Name,
			Logo:   team.Avatar,
			TeamID: faction,
		}
		
		// Get score if available
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
			matchData.GotvLink = "Not available (match might not be live yet)"
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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// OpenHudClient handles OpenHud API interactions
type OpenHudClient struct {
	baseURL    string
	httpClient *http.Client
}

// OpenHudTeam represents a team in OpenHud
type OpenHudTeam struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	ShortName   string `json:"shortName"`
	Logo        string `json:"logo"`
	LastUpdated int64  `json:"last_updated"`
	Extra       string `json:"extra"`
}

// OpenHudMatch represents a match in OpenHud
type OpenHudMatch struct {
	ID          string `json:"_id"`
	Team1       string `json:"team1"` // Team1 ID
	Team2       string `json:"team2"` // Team2 ID
	Score       [2]int `json:"score"`
	Status      string `json:"status"` // "waiting" | "in_progress" | "finished"
	LastUpdated int64  `json:"last_updated"`
}

// NewOpenHudClient creates a new OpenHud API client
func NewOpenHudClient(baseURL string) *OpenHudClient {
	if baseURL == "" {
		baseURL = "http://localhost:1349"
	}
	return &OpenHudClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateTeam creates a new team in OpenHud
func (c *OpenHudClient) CreateTeam(name, country, shortName string, logoPath string) (*OpenHudTeam, error) {
	url := fmt.Sprintf("%s/api/teams", c.baseURL)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("country", country)
	_ = writer.WriteField("shortName", shortName)
	_ = writer.WriteField("extra", "")

	// Add logo file if provided
	if logoPath != "" && fileExists(logoPath) {
		file, err := os.Open(logoPath)
		if err != nil {
			LogInfo("Failed to open logo file: %v", err)
		} else {
			defer file.Close()
			part, err := writer.CreateFormFile("logo", filepath.Base(logoPath))
			if err == nil {
				io.Copy(part, file)
			}
		}
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create team (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Response contains the team ID
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	LogInfo("Created team in OpenHud: %s (ID: %s)", name, result["_id"])

	return &OpenHudTeam{
		ID:          result["_id"],
		Name:        name,
		Country:     country,
		ShortName:   shortName,
		LastUpdated: time.Now().Unix(),
	}, nil
}

// GetAllTeams retrieves all teams from OpenHud
func (c *OpenHudClient) GetAllTeams() ([]OpenHudTeam, error) {
	url := fmt.Sprintf("%s/api/teams", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch teams (status %d)", resp.StatusCode)
	}

	var teams []OpenHudTeam
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return teams, nil
}

// FindTeamByName searches for a team by name
func (c *OpenHudClient) FindTeamByName(name string) (*OpenHudTeam, error) {
	teams, err := c.GetAllTeams()
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		if team.Name == name {
			return &team, nil
		}
	}

	return nil, nil // Not found
}

// UpdateTeam updates an existing team in OpenHud
func (c *OpenHudClient) UpdateTeam(teamID string, name, country, shortName string, logoPath string) error {
	url := fmt.Sprintf("%s/api/teams/%s", c.baseURL, teamID)

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	_ = writer.WriteField("_id", teamID)
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("country", country)
	_ = writer.WriteField("shortName", shortName)
	_ = writer.WriteField("extra", "")

	// Add logo file if provided
	if logoPath != "" && fileExists(logoPath) {
		file, err := os.Open(logoPath)
		if err != nil {
			LogInfo("Failed to open logo file: %v", err)
		} else {
			defer file.Close()
			part, err := writer.CreateFormFile("logo", filepath.Base(logoPath))
			if err == nil {
				io.Copy(part, file)
			}
		}
	}

	writer.Close()

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update team (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	LogInfo("Updated team in OpenHud: %s (ID: %s)", name, teamID)
	return nil
}

// CreateMatch creates a new match in OpenHud
func (c *OpenHudClient) CreateMatch(team1ID, team2ID string) (*OpenHudMatch, error) {
	url := fmt.Sprintf("%s/api/match", c.baseURL)

	matchData := map[string]interface{}{
		"team1":  team1ID,
		"team2":  team2ID,
		"score":  [2]int{0, 0},
		"status": "waiting",
	}

	jsonData, err := json.Marshal(matchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal match data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create match (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Response contains the match ID
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	LogInfo("Created match in OpenHud (ID: %s)", result["_id"])

	return &OpenHudMatch{
		ID:          result["_id"],
		Team1:       team1ID,
		Team2:       team2ID,
		Score:       [2]int{0, 0},
		Status:      "waiting",
		LastUpdated: time.Now().Unix(),
	}, nil
}

// SetCurrentMatch sets a match as the current active match
func (c *OpenHudClient) SetCurrentMatch(matchID string) error {
	url := fmt.Sprintf("%s/api/match/current/%s", c.baseURL, matchID)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set current match (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	LogInfo("Set current match in OpenHud (ID: %s)", matchID)
	return nil
}

// UpdateMatchScore updates the score of a match
func (c *OpenHudClient) UpdateMatchScore(matchID string, team1Score, team2Score int) error {
	url := fmt.Sprintf("%s/api/match/%s", c.baseURL, matchID)

	matchData := map[string]interface{}{
		"_id":   matchID,
		"score": [2]int{team1Score, team2Score},
	}

	jsonData, err := json.Marshal(matchData)
	if err != nil {
		return fmt.Errorf("failed to marshal match data: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update match score (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	LogInfo("Updated match score in OpenHud: %d - %d", team1Score, team2Score)
	return nil
}

// CheckConnection verifies if OpenHud is reachable
func (c *OpenHudClient) CheckConnection() error {
	url := fmt.Sprintf("%s/api/teams", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("OpenHud not reachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("OpenHud returned status %d", resp.StatusCode)
	}

	return nil
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

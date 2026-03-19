package mlb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/retr0h/freebie/services/api/internal/httputil"
)

const baseURL = "https://statsapi.mlb.com/api/v1"

// Client is an MLB Stats API client
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new MLB API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetSchedule fetches the schedule for a team on a specific date
func (c *Client) GetSchedule(ctx context.Context, teamID int, date time.Time) (*ScheduleResponse, error) {
	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf("%s/schedule?sportId=1&teamId=%d&date=%s", baseURL, teamID, dateStr)

	newReq := func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, "GET", url, nil)
	}

	resp, err := httputil.Do(ctx, c.httpClient, newReq, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var schedule ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &schedule, nil
}

// GetBoxScore fetches the boxscore for a specific game
func (c *Client) GetBoxScore(ctx context.Context, gamePk int) (*BoxScoreResponse, error) {
	url := fmt.Sprintf("%s/game/%d/boxscore", baseURL, gamePk)

	newReq := func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, "GET", url, nil)
	}

	resp, err := httputil.Do(ctx, c.httpClient, newReq, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching boxscore: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var boxscore BoxScoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&boxscore); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &boxscore, nil
}

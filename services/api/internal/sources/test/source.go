package test

import (
	"context"
	"fmt"
	"time"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

// Source implements the sources.Source interface for testing.
// It always returns a game with high stats so any trigger rule fires.
type Source struct{}

// NewSource creates a new test source
func NewSource() *Source {
	return &Source{}
}

// League returns "test"
func (s *Source) League() string {
	return "test"
}

// GetYesterdaysGame fetches stats from yesterday's game
func (s *Source) GetYesterdaysGame(ctx context.Context, teamID string) (*sources.GameStats, error) {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	return s.GetGameByDate(ctx, teamID, yesterday)
}

// GetGameByDate always returns a game with stats that fire any rule.
// Uses date-based game IDs so dedup creates one triggered event per day.
func (s *Source) GetGameByDate(_ context.Context, teamID string, date time.Time) (*sources.GameStats, error) {
	dateStr := date.Format("2006-01-02")

	return &sources.GameStats{
		GameID:   fmt.Sprintf("test-%s-%s", teamID, dateStr),
		GameDate: date,
		TeamID:   teamID,
		Opponent: "Test Opponent",
		HomeGame: true,
		Won:      true,
		Metrics: map[string]int{
			"score":       999,
			"strikeouts":  99,
			"runs":        99,
			"hits":        99,
			"home_runs":   99,
			"win":         1,
		},
	}, nil
}

// Register the test source on package init
func init() {
	sources.Register(NewSource())
}

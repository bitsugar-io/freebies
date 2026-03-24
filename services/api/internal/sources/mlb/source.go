package mlb

import (
	"context"
	"fmt"
	"time"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

// Source implements the sources.Source interface for MLB
type Source struct {
	client *Client
}

// NewSource creates a new MLB source
func NewSource() *Source {
	return &Source{
		client: NewClient(),
	}
}

// League returns "mlb"
func (s *Source) League() string {
	return "mlb"
}

// GetYesterdaysGame fetches stats from yesterday's game
func (s *Source) GetYesterdaysGame(ctx context.Context, teamID string) (*sources.GameStats, error) {
	// Use PT timezone since most deals are LA-based
	loc, _ := time.LoadLocation("America/Los_Angeles")
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	return s.GetGameByDate(ctx, teamID, yesterday)
}

// GetGameByDate fetches stats for a specific date
func (s *Source) GetGameByDate(ctx context.Context, teamID string, date time.Time) (*sources.GameStats, error) {
	// Look up MLB API team ID
	mlbTeamID, ok := TeamIDs[teamID]
	if !ok {
		return nil, fmt.Errorf("unknown team ID: %s", teamID)
	}

	// Fetch schedule for that date
	schedule, err := s.client.GetSchedule(ctx, mlbTeamID, date)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule: %w", err)
	}

	// Find a completed game
	var game *Game
	for _, d := range schedule.Dates {
		for _, g := range d.Games {
			if g.Status.AbstractGameState == "Final" {
				game = &g
				break
			}
		}
	}

	if game == nil {
		// No completed game on this date
		return nil, nil
	}

	// Fetch boxscore
	boxscore, err := s.client.GetBoxScore(ctx, game.GamePk)
	if err != nil {
		return nil, fmt.Errorf("fetching boxscore: %w", err)
	}

	// Determine if we're home or away
	isHome := game.Teams.Home.Team.ID == mlbTeamID
	var teamStats TeamBoxScore
	var opponentName string
	var won bool

	if isHome {
		teamStats = boxscore.Teams.Home
		opponentName = game.Teams.Away.Team.Name
		won = game.Teams.Home.Score > game.Teams.Away.Score
	} else {
		teamStats = boxscore.Teams.Away
		opponentName = game.Teams.Home.Team.Name
		won = game.Teams.Away.Score > game.Teams.Home.Score
	}

	// Build GameStats with all available metrics
	stats := &sources.GameStats{
		GameID:   fmt.Sprintf("mlb-%d", game.GamePk),
		GameDate: date,
		TeamID:   teamID,
		Opponent: opponentName,
		HomeGame: isHome,
		Won:      won,
		Metrics: map[string]int{
			// Pitching stats (what the team's pitchers did)
			"strikeouts":       teamStats.TeamStats.Pitching.StrikeOuts,
			"pitching_hits":    teamStats.TeamStats.Pitching.Hits,    // hits allowed
			"pitching_runs":    teamStats.TeamStats.Pitching.Runs,    // runs allowed
			"pitching_walks":   teamStats.TeamStats.Pitching.Walks,
			"pitching_homers":  teamStats.TeamStats.Pitching.HomeRuns, // HRs allowed

			// Batting stats (what the team's batters did)
			"runs":     teamStats.TeamStats.Batting.Runs,
			"hits":     teamStats.TeamStats.Batting.Hits,
			"home_runs": teamStats.TeamStats.Batting.HomeRuns,
			"rbi":      teamStats.TeamStats.Batting.RBI,

			// Win/loss as a metric (1 or 0)
			"win":      boolToInt(won),
			"home_win": boolToInt(won && isHome),
		},
	}

	return stats, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Register the MLB source on package init
func init() {
	sources.Register(NewSource())
}

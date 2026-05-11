package mlb

import (
	"context"
	"fmt"
	"time"

	mlbsdk "github.com/retr0h/mlb-sdk/pkg/mlb"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

// Source implements sources.Source for MLB using the public mlb-sdk client.
// All API parsing (schedule, boxscore, the FIELDING/DP free-text quirk)
// lives in mlb-sdk; this file does only the freebies-specific mapping
// from box-score fields into the flat metrics map.
type Source struct {
	client *mlbsdk.Client
}

// NewSource creates a new MLB source backed by mlb-sdk pointed at the
// public MLB Stats API. Tests use NewSourceWithClient to inject a client
// pointed at a fake server.
func NewSource() *Source {
	return &Source{client: mlbsdk.New()}
}

// NewSourceWithClient wraps an already-configured mlb-sdk client. Used by
// tests to point at an httptest fake.
func NewSourceWithClient(c *mlbsdk.Client) *Source {
	return &Source{client: c}
}

// League returns "mlb".
func (s *Source) League() string { return "mlb" }

// GetYesterdaysGame fetches stats from yesterday's game (PT, since deals are
// LA-anchored). Delegates to GetGameByDate.
func (s *Source) GetYesterdaysGame(
	ctx context.Context,
	teamID string,
) (*sources.GameStats, error) {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	return s.GetGameByDate(ctx, teamID, yesterday)
}

// GetGameByDate fetches a team's game for a specific date and returns the
// flat metrics map the trigger layer expects. Returns (nil, nil) when no
// completed game exists on that date.
func (s *Source) GetGameByDate(
	ctx context.Context,
	teamID string,
	date time.Time,
) (*sources.GameStats, error) {
	mlbTeamID, ok := TeamIDs[teamID]
	if !ok {
		return nil, fmt.Errorf("unknown team ID: %s", teamID)
	}

	games, err := s.client.Schedule(ctx, mlbsdk.ScheduleQuery{
		Team: mlbTeamID,
		On:   date,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching schedule: %w", err)
	}

	var game *mlbsdk.Game
	for i := range games {
		if games[i].Status == mlbsdk.StatusFinal {
			game = &games[i]
			break
		}
	}
	if game == nil {
		return nil, nil
	}

	box, err := s.client.Boxscore(ctx, game.GamePk)
	if err != nil {
		return nil, fmt.Errorf("fetching boxscore: %w", err)
	}

	team := box.Team(mlbTeamID)
	if team == nil {
		return nil, fmt.Errorf("team %s not present in boxscore for game %d", teamID, game.GamePk)
	}

	isHome := game.Home.ID == mlbTeamID
	var opponent string
	var won bool
	if isHome {
		opponent = game.Away.Name
		won = game.Home.Score > game.Away.Score
	} else {
		opponent = game.Home.Name
		won = game.Away.Score > game.Home.Score
	}

	doublePlays := team.DoublePlaysTurned()

	stats := &sources.GameStats{
		GameID:   fmt.Sprintf("mlb-%d", game.GamePk),
		GameDate: date,
		TeamID:   teamID,
		Opponent: opponent,
		HomeGame: isHome,
		Won:      won,
		Metrics: map[string]int{
			"strikeouts":      team.Pitching.Strikeouts,
			"pitching_hits":   team.Pitching.Hits,
			"pitching_runs":   team.Pitching.Runs,
			"pitching_walks":  team.Pitching.Walks,
			"pitching_homers": team.Pitching.HomeRuns,

			"runs":         team.Batting.Runs,
			"hits":         team.Batting.Hits,
			"home_runs":    team.Batting.HomeRuns,
			"rbi":          team.Batting.RBI,
			"stolen_bases": team.Batting.StolenBases,

			"double_plays": doublePlays,

			"home_runs_scored":  conditionalInt(isHome, team.Batting.Runs),
			"home_stolen_bases": conditionalInt(isHome, team.Batting.StolenBases),
			"home_double_plays": conditionalInt(isHome, doublePlays),

			"win":      boolToInt(won),
			"home_win": boolToInt(won && isHome),
		},
	}
	return stats, nil
}

func conditionalInt(cond bool, v int) int {
	if cond {
		return v
	}
	return 0
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func init() { sources.Register(NewSource()) }

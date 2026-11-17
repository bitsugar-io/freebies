package mlb

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"sort"
	"time"

	mlbsdk "github.com/retr0h/mlb-sdk/pkg/mlb"

	"github.com/retr0h/freebie/services/api/internal/httputil"
	"github.com/retr0h/freebie/services/api/internal/sources"
)

var _ sources.MultiGameSource = (*Source)(nil)

// Source implements sources.Source for MLB. Box scores (and the FIELDING/DP
// free-text quirk) are parsed by mlb-sdk; the schedule is read directly
// (statsapi.go) because the SDK drops gameType. This file does the
// freebies-specific mapping from box-score fields into the flat metrics map.
type Source struct {
	client  *mlbsdk.Client
	baseURL string
	http    *http.Client
}

// NewSource creates a new MLB source backed by mlb-sdk pointed at the
// public MLB Stats API. Tests use NewSourceWithClient to inject a client
// pointed at a fake server.
//
// Requests retry with backoff because the Stats API edge intermittently
// rejects requests from our cluster with 406, and trigger checks run only
// once a day, so a single failed call loses that day's deals.
func NewSource() *Source {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &httputil.RetryTransport{Options: &httputil.RetryOptions{
			RetryStatuses: []int{http.StatusNotAcceptable},
		}},
	}
	return NewSourceWithClient(mlbsdk.DefaultBaseURL, httpClient)
}

// NewSourceWithClient builds a source against the Stats API at baseURL using
// hc for every request. Tests point it at an httptest fake.
func NewSourceWithClient(baseURL string, hc *http.Client) *Source {
	return &Source{
		client:  mlbsdk.New(mlbsdk.WithBaseURL(baseURL), mlbsdk.WithHTTPClient(hc)),
		baseURL: baseURL,
		http:    hc,
	}
}

// League returns "mlb".
func (s *Source) League() string { return "mlb" }

// KnowsPostseason: the statsapi schedule carries gameType (F, D, L, W, C, P
// are postseason rounds).
func (s *Source) KnowsPostseason() bool { return true }

// Metrics lists the metrics the MLB source emits.
func (s *Source) Metrics() []string {
	return []string{
		"strikeouts", "pitching_hits", "pitching_runs", "pitching_walks", "pitching_homers",
		"runs", "hits", "home_runs", "rbi", "stolen_bases", "double_plays",
		"home_runs_scored", "home_stolen_bases", "home_double_plays",
		"win", "home_win",
	}
}

// Teams returns the team codes the source maps, sorted.
func (s *Source) Teams() []string {
	return slices.Sorted(maps.Keys(TeamIDs))
}

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

// GetGameByDate returns the team's first completed game on date (MLB's
// official date), or (nil, nil) when it has none. GetGamesByDate returns every
// game (a doubleheader has two).
func (s *Source) GetGameByDate(
	ctx context.Context,
	teamID string,
	date time.Time,
) (*sources.GameStats, error) {
	games, err := s.GetGamesByDate(ctx, teamID, date)
	if err != nil || len(games) == 0 {
		return nil, err
	}
	return games[0], nil
}

// GetGamesByDate returns every completed game the team played on date,
// earliest first, with the flat metrics map the trigger layer expects. Games
// that were postponed, canceled or suspended (MLB marks them Final), spring
// training and exhibitions are left out.
func (s *Source) GetGamesByDate(
	ctx context.Context,
	teamID string,
	date time.Time,
) ([]*sources.GameStats, error) {
	mlbTeamID, ok := TeamIDs[teamID]
	if !ok {
		return nil, fmt.Errorf("unknown team ID: %s", teamID)
	}

	games, err := s.schedule(ctx, mlbTeamID.Int(), date)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule: %w", err)
	}

	var played []scheduleGame
	for _, g := range games {
		if g.Status.AbstractGameState != string(mlbsdk.StatusFinal) ||
			skippedStatus(g.Status.DetailedState) || !g.counts() {
			continue
		}
		played = append(played, g)
	}
	sort.SliceStable(played, func(a, b int) bool { return played[a].GameDate.Before(played[b].GameDate) })

	out := make([]*sources.GameStats, 0, len(played))
	for i := range played {
		st, err := s.gameStats(ctx, &played[i], mlbTeamID, teamID, date)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, nil
}

// gameStats reads game's box score from mlbTeamID's side.
func (s *Source) gameStats(
	ctx context.Context,
	game *scheduleGame,
	mlbTeamID mlbsdk.TeamID,
	teamID string,
	date time.Time,
) (*sources.GameStats, error) {
	box, err := s.client.Boxscore(ctx, game.GamePk)
	if err != nil {
		return nil, fmt.Errorf("fetching boxscore: %w", err)
	}

	team := box.Team(mlbTeamID)
	if team == nil {
		return nil, fmt.Errorf("team %s not present in boxscore for game %d", teamID, game.GamePk)
	}

	home, away := game.Teams.Home, game.Teams.Away
	isHome := mlbsdk.TeamID(home.Team.ID) == mlbTeamID
	var opponent string
	var won bool
	if isHome {
		opponent = away.Team.Name
		won = home.Score > away.Score
	} else {
		opponent = home.Team.Name
		won = away.Score > home.Score
	}

	doublePlays := team.DoublePlaysTurned()

	stats := &sources.GameStats{
		GameID:     fmt.Sprintf("mlb-%d", game.GamePk),
		GameDate:   date,
		TeamID:     teamID,
		Opponent:   opponent,
		HomeGame:   isHome,
		Won:        won,
		Postseason: game.postseason(),
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

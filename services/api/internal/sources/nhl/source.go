// Package nhl implements sources.Source for the NHL.
//
// IMPORTANT: this uses the NHL's public web API (https://api-web.nhle.com),
// which is unofficial and undocumented. It needs no key, but its terms
// restrict commercial automated use and the shape may change without notice.
// Budget a licensed data feed before monetizing the app.
//
// Metrics emitted for the deal's team in the game:
//
//	win, home_win        1 if the team won (at home), else 0
//	goals                goals scored by the team; shootout goals are not goals
//	                     (the scoreboard credits the shootout winner one)
//	home_goals_scored    goals if the team was home, else 0
//	shutout              1 if the opponent scored 0 (never in a shootout, where the
//	                     winner is credited one goal on the scoreboard)
//	overtime             1 if the game was decided in overtime
//	shootout             1 if the game was decided in a shootout
//	p2_last_minute_goals goals the team scored in the last minute of the 2nd
//	                     period (19:00 or later on the period clock); the Kings
//	                     "McFlurry Minute"
//
// An NHL game cannot end tied, so a win is a win however it was decided: OT
// and shootout wins count as wins. The Ducks McGriddles promotion page
// (https://www.nhl.com/ducks/fans/partner-promotions, read 2026-11-15) says
// "If the Ducks win at home, get $2 ... McGriddles the next day"; it makes no
// exception for overtime or shootout wins.
package nhl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"sort"
	"time"

	"github.com/retr0h/freebie/services/api/internal/httputil"
	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/ttlcache"
)

const (
	defaultBaseURL = "https://api-web.nhle.com/v1"
	userAgent      = "freebies-game-checker/1.0"
	cacheTTL       = 10 * time.Minute
	cacheMax       = 64 // dates kept
)

var _ sources.MultiGameSource = (*Source)(nil)

// Source implements sources.Source for the NHL.
type Source struct {
	baseURL string
	client  *http.Client

	// cache holds each date's games (keyed YYYY-MM-DD) for cacheTTL, at most
	// cacheMax dates, sharing one in-flight fetch per date.
	cache *ttlcache.Cache[[]game]
}

// NewSource creates an NHL source pointed at the public web API. Requests
// retry with backoff on network errors, 429 and 5xx, since trigger checks
// run once a day and a single failed call loses that day's deals.
func NewSource() *Source {
	return NewSourceWithClient(defaultBaseURL, &http.Client{
		Timeout:   30 * time.Second,
		Transport: &httputil.RetryTransport{Options: &httputil.RetryOptions{}},
	})
}

// NewSourceWithClient builds a source against baseURL using client. Tests
// point it at an httptest server.
func NewSourceWithClient(baseURL string, client *http.Client) *Source {
	return &Source{baseURL: baseURL, client: client, cache: ttlcache.New[[]game](cacheTTL, cacheMax)}
}

// League returns "nhl".
func (s *Source) League() string { return "nhl" }

// KnowsPostseason: the score feed carries gameType (3 = playoffs).
func (s *Source) KnowsPostseason() bool { return true }

// Metrics lists the metrics the NHL source emits.
func (s *Source) Metrics() []string {
	return []string{
		"win", "home_win", "goals", "home_goals_scored", "shutout", "overtime", "shootout",
		"p2_last_minute_goals",
	}
}

// Teams returns the team codes the source maps, sorted.
func (s *Source) Teams() []string {
	return slices.Sorted(maps.Keys(TeamIDs))
}

// GetYesterdaysGame fetches yesterday's game (America/Los_Angeles, since
// deals are LA-anchored).
func (s *Source) GetYesterdaysGame(ctx context.Context, teamID string) (*sources.GameStats, error) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil, fmt.Errorf("loading timezone: %w", err)
	}
	return s.GetGameByDate(ctx, teamID, time.Now().In(loc).AddDate(0, 0, -1))
}

// GetGameByDate returns the team's first finished game on the Pacific date,
// or (nil, nil) when the team has no finished game that day (none scheduled,
// not yet final, postponed, preseason). GetGamesByDate returns every game.
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

// GetGamesByDate returns every finished, counting game the team started on
// the given Pacific date (only its Y/M/D are used), earliest first.
//
// The NHL files a game under its local date, which is never earlier than the
// Pacific date but can be a day later: a matinee in Europe or Asia starts the
// evening before in Los Angeles. So the date and the next one are fetched and
// games are picked by the Pacific date of startTimeUTC (gameDate when the
// start time is missing).
func (s *Source) GetGamesByDate(
	ctx context.Context,
	teamID string,
	date time.Time,
) ([]*sources.GameStats, error) {
	abbrev, ok := TeamIDs[teamID]
	if !ok {
		return nil, fmt.Errorf("unknown team ID: %s", teamID)
	}

	day := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)
	want := day.Format("2006-01-02")
	var picked []*game
	seen := map[int64]bool{}
	for _, d := range []time.Time{day, day.AddDate(0, 0, 1)} {
		games, err := s.games(ctx, d.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		for i := range games {
			g := &games[i]
			if seen[g.ID] || pacificDate(g) != want {
				continue
			}
			if g.HomeTeam.Abbrev != abbrev && g.AwayTeam.Abbrev != abbrev {
				continue
			}
			if !counts(g) || !isFinal(g) {
				continue
			}
			seen[g.ID] = true
			picked = append(picked, g)
		}
	}
	sort.SliceStable(picked, func(a, b int) bool { return picked[a].StartTimeUTC < picked[b].StartTimeUTC })

	out := make([]*sources.GameStats, 0, len(picked))
	for _, g := range picked {
		st, err := buildStats(g, teamID, abbrev, date)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, nil
}

// counts reports whether a game counts for deals: regular season (gameType 2)
// and playoffs (3). Preseason (1), the All-Star game (4) and international
// events (the 4 Nations Face-Off was 19) do not. A missing gameType (0, as in
// trimmed fixtures) counts as regular season.
func counts(g *game) bool {
	switch g.GameType {
	case 0, 2, 3:
		return true
	}
	return false
}

// pacificDate is the America/Los_Angeles date (YYYY-MM-DD) the game starts
// on, or its NHL gameDate when startTimeUTC is missing or unreadable.
func pacificDate(g *game) string {
	t, err := time.Parse(time.RFC3339, g.StartTimeUTC)
	if err != nil {
		return g.GameDate
	}
	return t.In(pacific).Format("2006-01-02")
}

var pacific = func() *time.Location {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.FixedZone("PT", -8*3600)
	}
	return loc
}()

// isFinal reports whether the game is over and played. FINAL is the state
// right after the horn; OFF is the settled official state. LIVE, CRIT, PRE,
// FUT and postponed/cancelled games (gameScheduleState PPD, CNCL, SUSP) are
// not counted.
func isFinal(g *game) bool {
	if g.GameState != "FINAL" && g.GameState != "OFF" {
		return false
	}
	if g.GameScheduleState != "" && g.GameScheduleState != "OK" {
		return false
	}
	return g.HomeTeam.Score != nil && g.AwayTeam.Score != nil
}

func buildStats(g *game, teamID, abbrev string, date time.Time) (*sources.GameStats, error) {
	isHome := g.HomeTeam.Abbrev == abbrev
	us, them := g.AwayTeam, g.HomeTeam
	if isHome {
		us, them = g.HomeTeam, g.AwayTeam
	}
	// The scoreboard score decides the winner (a shootout winner is credited
	// one extra goal).
	won := *us.Score > *them.Score

	lastPeriod := g.GameOutcome.LastPeriodType
	if lastPeriod == "" {
		lastPeriod = g.PeriodDescriptor.PeriodType
	}

	// The goal list is the only record of when goals came; a scoring game
	// without one is an error, never a silent 0.
	if (*us.Score > 0 || *them.Score > 0) && len(g.Goals) == 0 {
		return nil, fmt.Errorf("NHL game %d has a score but no goal list", g.ID)
	}
	// Goals are real goals only: the goal list without the shootout (periodType
	// "SO"), whose winner the scoreboard credits with one goal.
	ours, theirs, lateP2 := 0, 0, 0
	for _, gl := range g.Goals {
		if gl.PeriodDescriptor.PeriodType == "SO" {
			continue
		}
		if gl.TeamAbbrev == abbrev {
			ours++
		} else {
			theirs++
		}
		if gl.TeamAbbrev == abbrev && gl.Period == 2 && gl.PeriodDescriptor.PeriodType == "REG" &&
			gl.TimeInPeriod >= "19:00" {
			lateP2++
		}
	}
	// Outside a shootout the goal list must add up to the score, or the
	// goal metrics would be guessed.
	if lastPeriod != "SO" && (ours != *us.Score || theirs != *them.Score) {
		return nil, fmt.Errorf("NHL game %d: goal list (%d-%d) does not match the score (%d-%d)",
			g.ID, ours, theirs, *us.Score, *them.Score)
	}

	return &sources.GameStats{
		GameID:     fmt.Sprintf("nhl-%d", g.ID),
		GameDate:   date,
		TeamID:     teamID,
		Opponent:   them.Name.Default,
		HomeGame:   isHome,
		Won:        won,
		Postseason: g.GameType == 3,
		Metrics: map[string]int{
			"win":                  boolToInt(won),
			"home_win":             boolToInt(won && isHome),
			"goals":                ours,
			"home_goals_scored":    conditionalInt(isHome, ours),
			"shutout":              boolToInt(theirs == 0 && lastPeriod != "SO"),
			"overtime":             boolToInt(lastPeriod == "OT"),
			"shootout":             boolToInt(lastPeriod == "SO"),
			"p2_last_minute_goals": lateP2,
		},
	}, nil
}

// games returns the day's games, cached briefly so one check run over many
// events does not refetch the same date.
func (s *Source) games(ctx context.Context, day string) ([]game, error) {
	return s.cache.Get(ctx, day, func(ctx context.Context) ([]game, error) {
		return s.fetchGames(ctx, day)
	})
}

func (s *Source) fetchGames(ctx context.Context, day string) ([]game, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/score/"+day, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching NHL scores for %s: %w", day, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching NHL scores for %s: status %d", day, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("reading NHL scores for %s: %w", day, err)
	}
	var sr scoreResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("decoding NHL scores for %s: %w", day, err)
	}
	// A real day without games still has "games": []. An error body or an
	// empty object is an error, not "no games" (and is not cached).
	if sr.Games == nil {
		return nil, fmt.Errorf("NHL scores for %s have no games list: %.200s", day, body)
	}
	return *sr.Games, nil
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

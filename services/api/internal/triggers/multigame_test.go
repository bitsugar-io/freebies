package triggers_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

// multiSource returns whatever games the test sets, through GetGamesByDate,
// like an MLB doubleheader or NCAA football and basketball on one day.
type multiSource struct {
	mu    sync.Mutex
	games []*sources.GameStats
	err   error
}

func (s *multiSource) League() string        { return "multitest" }
func (s *multiSource) Metrics() []string     { return []string{"runs", "win"} }
func (s *multiSource) KnowsPostseason() bool { return true }

func (s *multiSource) GetYesterdaysGame(ctx context.Context, teamID string) (*sources.GameStats, error) {
	return s.GetGameByDate(ctx, teamID, time.Now())
}

func (s *multiSource) GetGameByDate(ctx context.Context, teamID string, date time.Time) (*sources.GameStats, error) {
	games, err := s.GetGamesByDate(ctx, teamID, date)
	if err != nil || len(games) == 0 {
		return nil, err
	}
	return games[0], nil
}

func (s *multiSource) GetGamesByDate(context.Context, string, time.Time) ([]*sources.GameStats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.games, s.err
}

func (s *multiSource) set(err error, games ...*sources.GameStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games, s.err = games, err
}

var multi = &multiSource{}

func init() { sources.Register(multi) }

func game(id string, runs int, postseason bool) *sources.GameStats {
	return &sources.GameStats{
		GameID: id, TeamID: "LAD", Postseason: postseason,
		Metrics: map[string]int{"runs": runs, "win": 1},
	}
}

// multiChecker is a checker over a fresh database whose only active event is
// a multitest deal with rule.
func multiChecker(t *testing.T, rule string) (*triggers.Checker, *sql.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE events SET is_active = 0`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO events (id, offer_id, team_id, team_name, league,
		partner_name, offer_name, offer_description, trigger_condition, trigger_rule, region_code, is_active)
		VALUES ('event-multi', 'multi', 'LAD', 'Dodgers', 'multitest', 'P', 'O', 'D', 'C', ?, 'us-ca-la', 1)`,
		rule); err != nil {
		t.Fatal(err)
	}
	pt, _ := time.LoadLocation("America/Los_Angeles")
	c := triggers.NewChecker(db.New(database),
		triggers.WithClock(func() time.Time { return time.Date(2026, 7, 20, 7, 0, 0, 0, pt) }))
	return c, database
}

func triggeredGames(t *testing.T, database *sql.DB) []string {
	t.Helper()
	rows, err := database.Query(`SELECT game_id FROM triggered_events WHERE event_id = 'event-multi' ORDER BY game_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		out = append(out, id)
	}
	return out
}

const runsRule = `{"metric":"runs","operator":">=","value":6,"redemption_window":"same_day"}`

var gameDay = time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC)

// Doubleheader: game 1 misses, game 2 meets the rule. The deal fires on game 2.
func TestCheckerEvaluatesEveryGameOfTheDay(t *testing.T) {
	c, database := multiChecker(t, runsRule)
	multi.set(nil, game("g1", 2, false), game("g2", 8, false))

	results, _, err := c.CheckAllForDate(context.Background(), gameDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].Triggered || results[0].Error != nil {
		t.Fatalf("results = %+v, want one triggered result", results)
	}
	if results[0].Stats.GameID != "g2" {
		t.Errorf("fired on %s, want g2", results[0].Stats.GameID)
	}
	if got := triggeredGames(t, database); len(got) != 1 || got[0] != "g2" {
		t.Errorf("triggered_events = %v, want [g2]", got)
	}
}

// Both games qualify: the deal fires once for the date (sponsors give one a
// day), on the first game, and a rerun, even after a stat correction makes the
// other game the first to qualify, adds nothing.
func TestCheckerFiresAtMostOncePerDate(t *testing.T) {
	c, database := multiChecker(t, runsRule)
	multi.set(nil, game("g1", 2, false), game("g2", 8, false))
	if _, _, err := c.CheckAllForDate(context.Background(), gameDay); err != nil {
		t.Fatal(err)
	}

	multi.set(nil, game("g1", 7, false), game("g2", 8, false))
	results, _, err := c.CheckAllForDate(context.Background(), gameDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].TriggeredEventID != "" || results[0].Error != nil {
		t.Errorf("rerun results = %+v, want nothing new", results)
	}
	if got := triggeredGames(t, database); len(got) != 1 || got[0] != "g2" {
		t.Errorf("triggered_events = %v, want only [g2]", got)
	}

	// A fresh database where both qualify from the start: one deal, game 1.
	c, database = multiChecker(t, runsRule)
	multi.set(nil, game("g1", 7, false), game("g2", 8, false))
	if _, _, err := c.CheckAllForDate(context.Background(), gameDay); err != nil {
		t.Fatal(err)
	}
	if got := triggeredGames(t, database); len(got) != 1 || got[0] != "g1" {
		t.Errorf("triggered_events = %v, want [g1]", got)
	}
}

// A regular-season-only deal skips a playoff game but still fires on a
// regular-season game the same day.
func TestCheckerSeasonLimitPerGame(t *testing.T) {
	c, database := multiChecker(t,
		`{"metric":"runs","operator":">=","value":6,"redemption_window":"same_day","season":"regular"}`)
	multi.set(nil, game("g1", 9, true), game("g2", 6, false))
	if _, _, err := c.CheckAllForDate(context.Background(), gameDay); err != nil {
		t.Fatal(err)
	}
	if got := triggeredGames(t, database); len(got) != 1 || got[0] != "g2" {
		t.Errorf("triggered_events = %v, want [g2]", got)
	}
}

// No game, or a failing source: nothing fires, and a failure is an error.
func TestCheckerNoGamesAndErrors(t *testing.T) {
	c, database := multiChecker(t, runsRule)
	multi.set(nil)
	results, _, err := c.CheckAllForDate(context.Background(), gameDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Triggered || results[0].Stats != nil {
		t.Errorf("results = %+v, want no game", results)
	}

	multi.set(errors.New("feed down"))
	results, _, err = c.CheckAllForDate(context.Background(), gameDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Error == nil {
		t.Errorf("results = %+v, want an error", results)
	}
	if got := triggeredGames(t, database); len(got) != 0 {
		t.Errorf("triggered_events = %v, want none", got)
	}
}

// One feed failed (NCAA basketball down, football read): the games that were
// read still fire the deal, but when none of them meets the rule the day is
// incomplete, so the result is an error, never a quiet "no game" that would
// hide a missed deal.
func TestCheckerPartialFeedFailure(t *testing.T) {
	partial := &sources.PartialError{Err: errors.New("basketball feed: status 500")}

	// No game read, one feed down: an error, not "no game".
	c, database := multiChecker(t, runsRule)
	multi.set(partial)
	results, _, err := c.CheckAllForDate(context.Background(), gameDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Error == nil || results[0].Triggered {
		t.Errorf("results = %+v, want an error and no trigger", results)
	}

	// A game was read but misses the rule: still an error.
	multi.set(partial, game("g1", 2, false))
	results, _, err = c.CheckAllForDate(context.Background(), gameDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Error == nil || results[0].Triggered {
		t.Errorf("results = %+v, want an error and no trigger", results)
	}
	if got := triggeredGames(t, database); len(got) != 0 {
		t.Errorf("triggered_events = %v, want none", got)
	}

	// A game that was read meets the rule: the deal fires, no error.
	multi.set(partial, game("g1", 8, false))
	results, _, err = c.CheckAllForDate(context.Background(), gameDay)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].Triggered || results[0].Error != nil {
		t.Errorf("results = %+v, want one clean trigger", results)
	}
	if got := triggeredGames(t, database); len(got) != 1 || got[0] != "g1" {
		t.Errorf("triggered_events = %v, want [g1]", got)
	}
}

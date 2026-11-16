package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
	"github.com/retr0h/freebie/services/api/internal/sources"
)

// fakeScheduler returns fixed games (or an error) and counts calls.
type fakeScheduler struct {
	games []sources.ScheduledGame
	err   error
	calls atomic.Int32
	date  atomic.Value // last date asked for
}

func (f *fakeScheduler) GamesOn(_ context.Context, date time.Time) ([]sources.ScheduledGame, error) {
	f.calls.Add(1)
	f.date.Store(date.Format("2006-01-02"))
	return f.games, f.err
}

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func getToday(t *testing.T, h *handlers.Handler, query string) (handlers.TodayResponse, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/today"+query, nil)
	w := httptest.NewRecorder()
	h.GetToday(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var resp handlers.TodayResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp, w.Body.String()
}

func TestGetToday(t *testing.T) {
	f := newClaimedFixture(t)
	// Seeded active events include event-nhl-ana-mcdonalds (ANA) and
	// event-nfl-lar-carlsjr (LAR). The Kings' deals are switched off here so
	// their game shows the filter: a team with no active deal is left out.
	f.exec(t, `UPDATE events SET is_active = 0 WHERE team_id = 'LAK'`)
	f.exec(t, `INSERT INTO triggered_events (id, event_id, expires_at, available_at)
		VALUES ('te-ducks', 'event-nhl-ana-mcdonalds', datetime('now', '+2 days'), NULL)`)
	f.exec(t, `INSERT INTO triggered_events (id, event_id, expires_at)
		VALUES ('te-rams', 'event-nfl-lar-carlsjr', datetime('now', '+1 day'))`)
	f.exec(t, `INSERT INTO triggered_events (id, event_id, expires_at)
		VALUES ('te-old', 'event-nhl-ana-mcdonalds', datetime('now', '-1 hour'))`)

	nhl := &fakeScheduler{games: []sources.ScheduledGame{
		{TeamID: "LAK", Opponent: "Oilers", StartsAt: mustTime(t, "2026-11-16T03:30:00Z"), Home: true, State: "pre"},
		{TeamID: "ANA", Opponent: "Flames", StartsAt: mustTime(t, "2026-11-16T03:00:00Z"), Home: true, State: "pre"},
		{TeamID: "CGY", Opponent: "Ducks", StartsAt: mustTime(t, "2026-11-16T03:00:00Z"), Home: false, State: "pre"},
	}}
	nfl := &fakeScheduler{err: errors.New("espn is down")}
	f.h.SetSchedulerLookup(func(league string) (sources.Scheduler, bool) {
		switch league {
		case "nhl":
			return nhl, true
		case "nfl":
			return nfl, true
		}
		return nil, false
	})

	resp, body := getToday(t, f.h, "?at=2026-11-15")
	t.Logf("GET /api/v1/today?at=2026-11-15\n%s", body)

	if resp.Date != "2026-11-15" {
		t.Errorf("date = %q", resp.Date)
	}
	if got := nhl.date.Load(); got != "2026-11-15" {
		t.Errorf("scheduler asked for %v", got)
	}

	// live: unexpired only, expiresAt ascending, sprites attached.
	if len(resp.Live) != 2 || resp.Live[0].ID != "te-rams" || resp.Live[1].ID != "te-ducks" {
		t.Fatalf("live = %+v", resp.Live)
	}
	for _, d := range resp.Live {
		if d.ExpiresAt == nil || d.Event.Sprite == nil || d.Event.ID != d.EventID {
			t.Errorf("live deal %s incomplete: %+v", d.ID, d)
		}
	}

	// games: only ANA (Kings have no active deal, CGY isn't ours); the NFL
	// failure is skipped rather than failing the response.
	if len(resp.Games) != 1 {
		t.Fatalf("games = %+v", resp.Games)
	}
	g := resp.Games[0]
	if g.TeamID != "ANA" || g.TeamName != "Ducks" || g.League != "nhl" || g.Opponent != "Flames" ||
		g.StartsAt != "2026-11-16T03:00:00Z" || !g.Home || g.State != "pre" {
		t.Errorf("game = %+v", g)
	}
	// Every active ANA event rides on the game, each with its sprite.
	var mcgriddles bool
	for _, d := range g.Deals {
		if d.TeamID != "ANA" || !d.IsActive || d.Sprite == nil {
			t.Errorf("deal %s: team %s active %v sprite %v", d.ID, d.TeamID, d.IsActive, d.Sprite != nil)
		}
		mcgriddles = mcgriddles || d.ID == "event-nhl-ana-mcdonalds"
	}
	if !mcgriddles {
		t.Errorf("deals = %+v", g.Deals)
	}

	// Cached per league and date; an error is not cached.
	getToday(t, f.h, "?at=2026-11-15")
	if nhl.calls.Load() != 1 {
		t.Errorf("nhl schedule fetched %d times, want 1", nhl.calls.Load())
	}
	if nfl.calls.Load() != 2 {
		t.Errorf("failed nfl schedule fetched %d times, want 2", nfl.calls.Load())
	}
	getToday(t, f.h, "?at=2026-11-16")
	if nhl.calls.Load() != 2 {
		t.Errorf("a new date should refetch, got %d calls", nhl.calls.Load())
	}
}

func TestGetTodayEmpty(t *testing.T) {
	f := newClaimedFixture(t)
	f.h.SetSchedulerLookup(func(string) (sources.Scheduler, bool) { return nil, false })

	_, body := getToday(t, f.h, "")
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["live"]) != "[]" || string(raw["games"]) != "[]" {
		t.Errorf("want empty arrays, got %s", body)
	}
	want := time.Now().In(mustLA(t)).Format("2006-01-02")
	if string(raw["date"]) != `"`+want+`"` {
		t.Errorf("date = %s, want %s", raw["date"], want)
	}
}

func TestGetTodayBadDate(t *testing.T) {
	f := newClaimedFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/today?at=nope", nil)
	w := httptest.NewRecorder()
	f.h.GetToday(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func mustLA(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

// On a basketball night, UCLA's football-only deal (football_home_points)
// must not ride on the game; the basketball deal does.
func TestTodayDealsMatchTheGamesSport(t *testing.T) {
	f := newClaimedFixture(t)
	ncaa := &fakeScheduler{games: []sources.ScheduledGame{
		{TeamID: "UCLA", Opponent: "Long Beach St", StartsAt: mustTime(t, "2026-11-17T03:00:00Z"), Home: true, State: "pre", Sport: "basketball"},
	}}
	f.h.SetSchedulerLookup(func(league string) (sources.Scheduler, bool) {
		if league == "ncaa" {
			return ncaa, true
		}
		return nil, false
	})
	resp, _ := getToday(t, f.h, "?at=2026-11-16")
	if len(resp.Games) != 1 {
		t.Fatalf("games = %+v", resp.Games)
	}
	g := resp.Games[0]
	if g.TeamName != "UCLA" {
		t.Errorf("teamName = %q, want the short name UCLA", g.TeamName)
	}
	if len(g.Deals) != 1 || g.Deals[0].ID != "event-ncaa-ucla-fatsals-hoops" {
		ids := []string{}
		for _, d := range g.Deals {
			ids = append(ids, d.ID)
		}
		t.Errorf("deals = %v, want only the basketball deal", ids)
	}
}

package mlb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixtureDir = "../../../test/fixtures"

// fixtureServer serves real Stats API responses, trimmed:
//
//	/api/v1/schedule?teamId=T&date=D  from mlb_schedule_{T}_{D}.json
//	/api/v1/game/{pk}/boxscore        from mlb_boxscore_{pk}.json
//
// A schedule with no fixture is an empty day; a boxscore with none is a 404.
func fixtureServer(t *testing.T) *Source {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var name string
		switch {
		case strings.HasSuffix(r.URL.Path, "/schedule"):
			q := r.URL.Query()
			assert.Equal(t, "1", q.Get("sportId"))
			name = "mlb_schedule_" + q.Get("teamId") + "_" + q.Get("date") + ".json"
		case strings.HasSuffix(r.URL.Path, "/boxscore"):
			pk := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/game/"), "/boxscore")
			name = "mlb_boxscore_" + pk + ".json"
		}
		b, err := os.ReadFile(filepath.Join(fixtureDir, name))
		if err != nil {
			if strings.HasPrefix(name, "mlb_schedule_") {
				b = []byte(`{"dates":[]}`)
			} else {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	}))
	t.Cleanup(srv.Close)
	return NewSourceWithClient(srv.URL, srv.Client())
}

func date(s string) time.Time {
	d, _ := time.Parse("2006-01-02", s)
	return d
}

// Spring training (gameType "S") never counts: Angels at Dodgers, 2026-03-01,
// Final.
func TestSpringTrainingNeverCounts(t *testing.T) {
	got, err := fixtureServer(t).GetGameByDate(context.Background(), "LAD", date("2026-03-01"))
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestExhibitionsNeverCountInSchedule(t *testing.T) {
	const schedule = `{"dates": [{"games": [
		{"gamePk": 1, "gameType": "S", "gameDate": "2026-03-01T20:05:00Z",
		 "status": {"abstractGameState": "Final", "detailedState": "Final"},
		 "teams": {"home": {"team": {"id": 119, "name": "Los Angeles Dodgers"}},
		           "away": {"team": {"id": 108, "name": "Los Angeles Angels"}}}},
		{"gamePk": 2, "gameType": "E", "gameDate": "2026-03-01T22:05:00Z",
		 "status": {"abstractGameState": "Final", "detailedState": "Final"},
		 "teams": {"home": {"team": {"id": 108, "name": "Los Angeles Angels"}},
		           "away": {"team": {"id": 119, "name": "Los Angeles Dodgers"}}}}
	]}]}`
	srv := fakeMLB{scheduleStatus: http.StatusOK, scheduleBody: schedule}.server()
	t.Cleanup(srv.Close)
	src := NewSourceWithClient(srv.URL, srv.Client())
	got, err := src.GamesOn(context.Background(), date("2026-03-01"))
	require.NoError(t, err)
	assert.Empty(t, got, "spring training and exhibitions are not games today")
	games, err := src.GetGamesByDate(context.Background(), "LAD", date("2026-03-01"))
	require.NoError(t, err)
	assert.Empty(t, games)
}

// The Stats API lists Dodgers at Yankees (gamePk 823523) on 2026-07-18 as
// abstractGameState "Final", detailedState "Postponed"; it was played the next
// day. It is not a game on the 18th, and its boxscore (the win on the 19th)
// must not be read for the 18th.
func TestPostponedGameIsSkipped(t *testing.T) {
	got, err := fixtureServer(t).GetGameByDate(context.Background(), "LAD", date("2026-07-18"))
	require.NoError(t, err)
	assert.Nil(t, got)
}

// A postponed game listed first does not hide the game that was played.
func TestPostponedGameDoesNotHidePlayedGame(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(fixtureDir, "mlb_boxscore_823521.json"))
	require.NoError(t, err)
	const schedule = `{"dates": [{"games": [
		{"gamePk": 823523, "gameType": "R", "gameDate": "2026-07-19T16:35:00Z",
		 "status": {"abstractGameState": "Final", "detailedState": "Postponed"},
		 "teams": {"away": {"team": {"id": 119, "name": "Los Angeles Dodgers"}},
		           "home": {"team": {"id": 147, "name": "New York Yankees"}}}},
		{"gamePk": 823521, "gameType": "R", "gameDate": "2026-07-19T23:20:00Z",
		 "status": {"abstractGameState": "Final", "detailedState": "Final"},
		 "teams": {"away": {"team": {"id": 119, "name": "Los Angeles Dodgers"}, "score": 1},
		           "home": {"team": {"id": 147, "name": "New York Yankees"}, "score": 2}}}
	]}]}`
	srv := fakeMLB{
		scheduleStatus: http.StatusOK, scheduleBody: schedule,
		boxscoreStatus: http.StatusOK, boxscoreBody: string(b),
	}.server()
	t.Cleanup(srv.Close)
	got, err := NewSourceWithClient(srv.URL, srv.Client()).GetGameByDate(context.Background(), "LAD", date("2026-07-19"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "mlb-823521", got.GameID)
}

// The Dodgers' 2026-07-19 doubleheader at the Yankees: won game 1 8-2, lost
// game 2 1-2. Both games are returned, in order.
func TestDoubleheaderReturnsBothGames(t *testing.T) {
	src := fixtureServer(t)
	games, err := src.GetGamesByDate(context.Background(), "LAD", date("2026-07-19"))
	require.NoError(t, err)
	require.Len(t, games, 2)

	assert.Equal(t, "mlb-823523", games[0].GameID)
	assert.True(t, games[0].Won)
	assert.Equal(t, 8, games[0].Metrics["runs"])
	assert.Equal(t, 1, games[0].Metrics["double_plays"])

	assert.Equal(t, "mlb-823521", games[1].GameID)
	assert.False(t, games[1].Won)
	assert.Equal(t, 1, games[1].Metrics["runs"])

	first, err := src.GetGameByDate(context.Background(), "LAD", date("2026-07-19"))
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, "mlb-823523", first.GameID)
}

// A 200 that is not a schedule is an error, not "no games".
func TestBadScheduleBodyIsAnError(t *testing.T) {
	for _, bad := range []string{`{}`, `{"messageNumber":11,"message":"Object not found"}`} {
		srv := fakeMLB{scheduleStatus: http.StatusOK, scheduleBody: bad}.server()
		t.Cleanup(srv.Close)
		src := NewSourceWithClient(srv.URL, srv.Client())
		_, err := src.GetGameByDate(context.Background(), "LAD", date("2026-07-19"))
		require.Error(t, err, bad)
		_, err = src.GamesOn(context.Background(), date("2026-07-19"))
		require.Error(t, err, bad)
	}
}

package nhl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

func TestGamesOnFixture(t *testing.T) {
	var hits atomic.Int32
	src := fixtureServer(t, &hits)

	got, err := src.GamesOn(context.Background(), day(t, "2026-11-15"))
	require.NoError(t, err)
	// Five games, both sides mapped.
	require.Len(t, got, 10)

	var ducks *sources.ScheduledGame
	for i := range got {
		if got[i].TeamID == "ANA" {
			ducks = &got[i]
		}
	}
	require.NotNil(t, ducks)
	assert.Equal(t, "Flames", ducks.Opponent)
	assert.True(t, ducks.Home)
	assert.Equal(t, "pre", ducks.State)
	// The trimmed fixture has no startTimeUTC: midnight Pacific.
	assert.Equal(t, "2026-11-15T08:00:00Z", ducks.StartsAt.Format(time.RFC3339))

	// Shares GetGameByDate's per-date cache: only the next date is fetched.
	_, err = src.GetGameByDate(context.Background(), "ANA", day(t, "2026-11-15"))
	require.NoError(t, err)
	assert.Equal(t, int32(2), hits.Load())
}

func TestGamesOnStates(t *testing.T) {
	const body = `{"games":[
	 {"id":1,"gameDate":"2026-11-20","startTimeUTC":"2026-11-21T03:00:00Z","gameState":"LIVE","gameScheduleState":"OK",
	  "homeTeam":{"abbrev":"LAK","name":{"default":"Kings"}},"awayTeam":{"abbrev":"ANA","name":{"default":"Ducks"}}},
	 {"id":2,"gameDate":"2026-11-20","startTimeUTC":"2026-11-21T00:00:00Z","gameState":"OFF","gameScheduleState":"OK",
	  "homeTeam":{"abbrev":"BOS","name":{"default":"Bruins"}},"awayTeam":{"abbrev":"TOR","name":{"default":"Maple Leafs"}}},
	 {"id":3,"gameDate":"2026-11-20","startTimeUTC":"2026-11-21T01:00:00Z","gameState":"FUT","gameScheduleState":"PPD",
	  "homeTeam":{"abbrev":"SJS","name":{"default":"Sharks"}},"awayTeam":{"abbrev":"SEA","name":{"default":"Kraken"}}},
	 {"id":4,"gameDate":"2026-11-20","startTimeUTC":"2026-11-21T02:00:00Z","gameState":"FUT","gameScheduleState":"CNCL",
	  "homeTeam":{"abbrev":"VGK","name":{"default":"Golden Knights"}},"awayTeam":{"abbrev":"UTA","name":{"default":"Utah"}}}
	]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/score/2026-11-20", r.URL.Path)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	src := NewSourceWithClient(srv.URL, srv.Client())

	got, err := src.GamesOn(context.Background(), day(t, "2026-11-20"))
	require.NoError(t, err)
	start := func(s string) time.Time {
		ts, err := time.Parse(time.RFC3339, s)
		require.NoError(t, err)
		return ts
	}
	assert.Equal(t, []sources.ScheduledGame{
		{TeamID: "BOS", Opponent: "Maple Leafs", StartsAt: start("2026-11-21T00:00:00Z"), Home: true, State: "post"},
		{TeamID: "TOR", Opponent: "Bruins", StartsAt: start("2026-11-21T00:00:00Z"), Home: false, State: "post"},
		{TeamID: "LAK", Opponent: "Ducks", StartsAt: start("2026-11-21T03:00:00Z"), Home: true, State: "in"},
		{TeamID: "ANA", Opponent: "Kings", StartsAt: start("2026-11-21T03:00:00Z"), Home: false, State: "in"},
	}, got, "postponed and canceled games are left out")
}

func TestScheduleState(t *testing.T) {
	for in, want := range map[string]string{
		"FUT": "pre", "PRE": "pre", "LIVE": "in", "CRIT": "in", "FINAL": "post", "OFF": "post",
	} {
		got, ok := scheduleState(in)
		assert.True(t, ok, in)
		assert.Equal(t, want, got, in)
	}
	_, ok := scheduleState("WAT")
	assert.False(t, ok)
}

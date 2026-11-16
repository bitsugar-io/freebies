package espn_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/httputil"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
	"github.com/retr0h/freebie/services/api/internal/sources/nba"
)

func day(t *testing.T, s string) time.Time { return espntest.Day(t, s) }

// board renders a one-game scoreboard. start is the UTC kickoff.
func board(id, start, status string, completed bool, home, away string, extra string) string {
	state := "pre"
	if completed {
		state = "post"
	}
	return fmt.Sprintf(`{"events":[{"id":%q,"date":%q,"competitions":[{
		"status":{"type":{"name":%q,"state":%q,"completed":%t}},%s
		"competitors":[
		 {"homeAway":"home","score":"100","team":{"id":"13","displayName":"Home"}},
		 {"homeAway":"away","score":"90","team":{"id":"99","displayName":"Away"}}]}]}]}`,
		id, start, status, state, completed, extra)
}

// byDate serves a different body per ESPN day.
func byDate(t *testing.T, hits *atomic.Int32, bodies map[string]string) *espn.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			hits.Add(1)
		}
		b, ok := bodies[r.URL.Query().Get("dates")]
		if !ok {
			b = `{"events":[]}`
		}
		_, _ = w.Write([]byte(b))
	}))
	t.Cleanup(srv.Close)
	return espn.NewClientWith(srv.URL, srv.Client())
}

func nbaSource(c *espn.Client) *espn.Source {
	return espn.NewSource(espn.Config{
		League: "nba", Teams: map[string]string{"LAL": "13"},
		Feeds: []espn.FeedConfig{{
			Feed:  espn.Feed{Path: "basketball/nba"},
			Build: espn.PointsMetrics,
		}},
	}, c)
}

func TestLateGameFiledUnderNextDay(t *testing.T) {
	// 9pm PT on Nov 7 is 05:00Z Nov 8; ESPN files it under its Nov 8 day.
	c := byDate(t, nil, map[string]string{
		"20261108": board("1", "2026-11-08T05:00Z", "STATUS_FINAL", true, "LAL", "X", ""),
	})
	src := nbaSource(c)

	got, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-07"))
	require.NoError(t, err)
	require.NotNil(t, got, "a 9pm Pacific game belongs to the Pacific date it started on")

	got, err = src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-08"))
	require.NoError(t, err)
	assert.Nil(t, got, "and not to the next Pacific date")
}

func TestEarlyUTCGameFiledUnderPreviousDay(t *testing.T) {
	// 4pm PT on Nov 6 is 00:00Z Nov 7, but ESPN files it under Nov 6.
	c := byDate(t, nil, map[string]string{
		"20261106": board("1", "2026-11-07T00:00Z", "STATUS_FINAL", true, "LAL", "X", ""),
	})
	src := nbaSource(c)
	got, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-07"))
	require.NoError(t, err)
	assert.Nil(t, got, "the game is on Pacific Nov 6, not Nov 7")
	got, err = src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-06"))
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestNotCompletedGamesAreIgnored(t *testing.T) {
	for _, status := range []string{"STATUS_POSTPONED", "STATUS_CANCELED", "STATUS_IN_PROGRESS", "STATUS_SCHEDULED"} {
		t.Run(status, func(t *testing.T) {
			c := byDate(t, nil, map[string]string{
				"20261107": board("1", "2026-11-07T20:00Z", status, false, "LAL", "X", ""),
			})
			got, err := nbaSource(c).GetGameByDate(context.Background(), "LAL", day(t, "2026-11-07"))
			require.NoError(t, err)
			assert.Nil(t, got)
		})
	}
}

func TestCachesPerDate(t *testing.T) {
	var hits atomic.Int32
	srv := espntest.Server(t, &hits)
	src := nba.NewSourceWithClient(srv.URL, srv.Client())

	for _, team := range []string{"LAL", "LAC", "LAL"} {
		_, err := src.GetGameByDate(context.Background(), team, day(t, "2026-11-06"))
		require.NoError(t, err)
	}
	// Two scoreboards (the ESPN day and the next) plus the Lakers' home-game
	// summary (play-by-play), each fetched once.
	assert.Equal(t, int32(3), hits.Load(), "one date's pages are fetched once")

	_, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-09"))
	require.NoError(t, err)
	assert.Equal(t, int32(5), hits.Load(), "an away game adds only its two scoreboards")
}

func TestRetriesTransientErrors(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"events":[]}`))
	}))
	defer srv.Close()

	// Production wiring (RetryTransport) with a fast backoff.
	c := espn.NewClientWith(srv.URL, &http.Client{Transport: &httputil.RetryTransport{
		Options: &httputil.RetryOptions{MaxRetries: 3, BaseDelay: time.Millisecond},
	}})
	_, err := nbaSource(c).GetGameByDate(context.Background(), "LAL", day(t, "2026-11-06"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, hits.Load(), int32(2))
}

func TestServerErrorSurfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	src := nba.NewSourceWithClient(srv.URL, srv.Client())
	got, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-06"))
	require.Error(t, err, "an upstream failure is an error, never 'no game'")
	assert.Nil(t, got)
}

func TestUnreadableScoreIsAnError(t *testing.T) {
	b := `{"events":[{"id":"1","date":"2026-11-07T20:00Z","competitions":[{
		"status":{"type":{"name":"STATUS_FINAL","state":"post","completed":true}},
		"competitors":[
		 {"homeAway":"home","score":"","team":{"id":"13"}},
		 {"homeAway":"away","score":"90","team":{"id":"99"}}]}]}]}`
	c := byDate(t, nil, map[string]string{"20261107": b})
	_, err := nbaSource(c).GetGameByDate(context.Background(), "LAL", day(t, "2026-11-07"))
	require.Error(t, err)
}

func TestSoccerGoalsMustMatchScore(t *testing.T) {
	// The score says 100-90 but the events list no goals at all: refuse to
	// guess a first-half metric.
	c := byDate(t, nil, map[string]string{
		"20261107": board("1", "2026-11-07T20:00Z", "STATUS_FULL_TIME", true, "LAFC", "X", `"details":[],`),
	})
	src := espn.NewSource(espn.Config{
		League: "mls", Teams: map[string]string{"LAFC": "13"},
		Feeds: []espn.FeedConfig{{Feed: espn.Feed{Path: "soccer/usa.1"}, Build: espn.SoccerMetrics}},
	}, c)
	_, err := src.GetGameByDate(context.Background(), "LAFC", day(t, "2026-11-07"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "do not match the final score")
}

func TestSoccerPenaltyShootoutIsADraw(t *testing.T) {
	// 1-1 after extra time, decided on penalties: win = 0, draw = 1.
	b := `{"events":[{"id":"1","date":"2026-11-07T20:00Z","competitions":[{
		"status":{"type":{"name":"STATUS_FINAL_PEN","state":"post","completed":true}},
		"details":[
		 {"scoringPlay":true,"clock":{"value":600,"displayValue":"10'"},"team":{"id":"13"}},
		 {"scoringPlay":true,"clock":{"value":4000,"displayValue":"67'"},"team":{"id":"99"}},
		 {"scoringPlay":true,"shootout":true,"clock":{"value":0,"displayValue":"0'"},"team":{"id":"13"}}],
		"competitors":[
		 {"homeAway":"home","score":"1","winner":true,"team":{"id":"13","displayName":"Home"}},
		 {"homeAway":"away","score":"1","team":{"id":"99","displayName":"Away"}}]}]}]}`
	c := byDate(t, nil, map[string]string{"20261107": b})
	s := espn.NewSource(espn.Config{
		League: "mls", Teams: map[string]string{"LAFC": "13"},
		Feeds: []espn.FeedConfig{{Feed: espn.Feed{Path: "soccer/usa.1"}, Build: espn.SoccerMetrics}},
	}, c)
	got, err := s.GetGameByDate(context.Background(), "LAFC", day(t, "2026-11-07"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 0, got.Metrics["win"])
	assert.Equal(t, 1, got.Metrics["draw"])
	assert.Equal(t, 1, got.Metrics["home_scored_first_in_first_half"])
}

package espn_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
	"github.com/retr0h/freebie/services/api/internal/sources/nba"
	"github.com/retr0h/freebie/services/api/internal/sources/ncaa"
	"github.com/retr0h/freebie/services/api/internal/sources/nfl"
	"github.com/retr0h/freebie/services/api/internal/sources/nwsl"
)

// Preseason games never count: real ESPN responses, trimmed.
func TestPreseasonGamesNeverCount(t *testing.T) {
	srv := espntest.Server(t, nil)

	// LA Clippers at Toronto Raptors, 2026-10-10, season.type 1, STATUS_FINAL.
	got, err := nba.NewSourceWithClient(srv.URL, srv.Client()).
		GetGameByDate(context.Background(), "LAC", day(t, "2026-10-10"))
	require.NoError(t, err)
	assert.Nil(t, got, "an NBA preseason game is not a game for deals")

	// Los Angeles Rams at Kansas City Chiefs, 2026-08-15, season.type 1.
	got, err = nfl.NewSourceWithClient(srv.URL, srv.Client()).
		GetGameByDate(context.Background(), "LAR", day(t, "2026-08-15"))
	require.NoError(t, err)
	assert.Nil(t, got, "an NFL preseason game is not a game for deals")
}

// Today's schedule leaves preseason games out too.
func TestGamesOnSkipsPreseason(t *testing.T) {
	srv := espntest.Server(t, nil)
	got, err := nba.NewSourceWithClient(srv.URL, srv.Client()).GamesOn(context.Background(), day(t, "2026-10-10"))
	require.NoError(t, err)
	assert.Empty(t, got, "the Clippers' preseason game is not a game today")
}

func allStarBoard() string {
	return `{"events":[{"id":"1","date":"2026-02-15T20:00Z","season":{"type":2,"slug":"regular-season"},
	"competitions":[{"type":{"abbreviation":"ALLSTAR"},
	 "status":{"type":{"name":"STATUS_FINAL","state":"post","completed":true}},
	 "competitors":[
	  {"homeAway":"home","score":"100","team":{"id":"13","displayName":"Home"}},
	  {"homeAway":"away","score":"90","team":{"id":"99","displayName":"Away"}}]}]}]}`
}

// ESPN files the All-Star game as season.type 2; the competition type says
// what it is.
func TestAllStarGameNeverCounts(t *testing.T) {
	c := byDate(t, nil, map[string]string{"20260215": allStarBoard()})
	src := nbaSource(c)
	got, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-02-15"))
	require.NoError(t, err)
	assert.Nil(t, got)
	sched, err := src.GamesOn(context.Background(), day(t, "2026-02-15"))
	require.NoError(t, err)
	assert.Empty(t, sched)
}

// The NBA play-in (season.type 5) is not regular season.
func TestPlayInIsPostseason(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := espn.NewSource(espn.Config{
		League: "nba", Teams: map[string]string{"LAC": "12"},
		Feeds: []espn.FeedConfig{{Feed: espn.Feed{Path: "basketball/nba"}, Build: espn.PointsMetrics}},
	}, espn.NewClientWith(srv.URL, srv.Client()))
	// Golden State Warriors at LA Clippers, 2026-04-15 7pm PT, play-in.
	got, err := src.GetGameByDate(context.Background(), "LAC", day(t, "2026-04-15"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, got.Postseason)
}

// Soccer feeds use league-specific season type ids (NWSL 14023); the slug
// says it is a playoff game.
func TestSoccerPlayoffIsPostseason(t *testing.T) {
	srv := espntest.Server(t, nil)
	got, err := nwsl.NewSourceWithClient(srv.URL, srv.Client()).
		GetGameByDate(context.Background(), "ACFC", day(t, "2026-11-07"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, got.Postseason, "playoffs---quarterfinals is a playoff game")
}

// A 200 whose body is an error or has no events key is an error, never "no
// games", and it is not cached.
func TestBadScoreboardBodyIsAnError(t *testing.T) {
	for _, bad := range []string{`{}`, `{"code":400,"message":"Failed to get events endpoint."}`, `[]`} {
		t.Run(bad, func(t *testing.T) {
			var hits atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if hits.Add(1) == 1 {
					_, _ = w.Write([]byte(bad))
					return
				}
				_, _ = w.Write([]byte(`{"events":[]}`))
			}))
			t.Cleanup(srv.Close)
			src := nbaSource(espn.NewClientWith(srv.URL, srv.Client()))
			_, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-06"))
			require.Error(t, err)

			// The bad body was not cached: the next call refetches and works.
			got, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-06"))
			require.NoError(t, err)
			assert.Nil(t, got)
		})
	}
}

func twoGameBoard() string {
	game := func(id, start, us, them string) string {
		return fmt.Sprintf(`{"id":%q,"date":%q,"season":{"type":2},"competitions":[{
		"status":{"type":{"name":"STATUS_FINAL","state":"post","completed":true}},
		"competitors":[
		 {"homeAway":"home","score":%q,"team":{"id":"13","displayName":"Home"}},
		 {"homeAway":"away","score":%q,"team":{"id":"99","displayName":"Away"}}]}]}`, id, start, us, them)
	}
	return `{"events":[` + game("2", "2026-11-07T23:00Z", "100", "90") + "," +
		game("1", "2026-11-07T18:00Z", "80", "95") + `]}`
}

// Every finished game of the day is returned, earliest first; GetGameByDate
// stays the first.
func TestGetGamesByDateReturnsEveryGame(t *testing.T) {
	src := nbaSource(byDate(t, nil, map[string]string{"20261107": twoGameBoard()}))
	games, err := src.GetGamesByDate(context.Background(), "LAL", day(t, "2026-11-07"))
	require.NoError(t, err)
	require.Len(t, games, 2)
	assert.Equal(t, "espn-nba-1", games[0].GameID)
	assert.Equal(t, 0, games[0].Metrics["win"])
	assert.Equal(t, "espn-nba-2", games[1].GameID)
	assert.Equal(t, 1, games[1].Metrics["win"])

	first, err := src.GetGameByDate(context.Background(), "LAL", day(t, "2026-11-07"))
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, "espn-nba-1", first.GameID)
}

// failing serves the ESPN fixtures but answers 500 for any path containing
// one of fail.
func failing(t *testing.T, fail ...string) *httptest.Server {
	t.Helper()
	fixtures := espntest.Server(t, nil)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, f := range fail {
			if strings.Contains(r.URL.Path, f) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		resp, err := fixtures.Client().Get(fixtures.URL + r.URL.RequestURI())
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// NCAA reads two feeds; one failing must not hide the other's games.
func TestNCAAFeedFailureIsIsolated(t *testing.T) {
	// USC basketball home win on 2026-11-02 with the football feed down.
	srv := failing(t, "college-football")
	src := ncaa.NewSourceWithClient(srv.URL, srv.Client())
	games, err := src.GetGamesByDate(context.Background(), "USC", day(t, "2026-11-02"))
	// The basketball game comes back, flagged as an incomplete day so a
	// football deal reports an error instead of "no game".
	var partial *sources.PartialError
	require.ErrorAs(t, err, &partial)
	assert.Contains(t, err.Error(), "college-football")
	require.Len(t, games, 1)
	assert.Contains(t, games[0].GameID, "basketball")

	// GetGameByDate still answers with the game it could read.
	first, err := src.GetGameByDate(context.Background(), "USC", day(t, "2026-11-02"))
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Contains(t, first.GameID, "basketball")

	sched, err := src.GamesOn(context.Background(), day(t, "2026-11-02"))
	require.NoError(t, err)
	assert.NotEmpty(t, sched, "basketball games still listed")

	// USC football home win on 2026-10-31 with the basketball feed down.
	srv = failing(t, "mens-college-basketball")
	src = ncaa.NewSourceWithClient(srv.URL, srv.Client())
	games, err = src.GetGamesByDate(context.Background(), "USC", day(t, "2026-10-31"))
	require.ErrorAs(t, err, &partial)
	require.Len(t, games, 1)
	assert.Contains(t, games[0].GameID, "football")

	// Both down: an error, never "no games".
	srv = failing(t, "college-football", "mens-college-basketball")
	src = ncaa.NewSourceWithClient(srv.URL, srv.Client())
	_, err = src.GetGamesByDate(context.Background(), "USC", day(t, "2026-10-31"))
	require.Error(t, err)
	_, err = src.GamesOn(context.Background(), day(t, "2026-10-31"))
	require.Error(t, err)
}

package nhl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixtureDir = "../../../test/fixtures"

// fixtureServer serves test/fixtures/nhl_score_{date}.json for
// /score/{date}, an empty day ({"games":[]}) when there is no fixture, and
// counts requests.
func fixtureServer(t *testing.T, hits *atomic.Int32) *Source {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			hits.Add(1)
		}
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		day := strings.TrimPrefix(r.URL.Path, "/score/")
		b, err := os.ReadFile(filepath.Join(fixtureDir, "nhl_score_"+day+".json"))
		if err != nil {
			b = []byte(`{"games":[]}`)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	}))
	t.Cleanup(srv.Close)
	return NewSourceWithClient(srv.URL, srv.Client())
}

func day(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	require.NoError(t, err)
	return d
}

func TestGetGameByDate(t *testing.T) {
	tests := []struct {
		name     string
		team     string
		date     string
		wantNil  bool
		opponent string
		home     bool
		won      bool
		metrics  map[string]int
	}{
		{
			name: "home win in regulation", team: "ANA", date: "2026-11-08",
			opponent: "Penguins", home: true, won: true,
			metrics: map[string]int{
				"win": 1, "home_win": 1, "goals": 5, "home_goals_scored": 5,
				"shutout": 0, "overtime": 0, "shootout": 0, "p2_last_minute_goals": 0,
			},
		},
		{
			name: "away win is not a home win", team: "ANA", date: "2026-10-09",
			opponent: "Jets", home: false, won: true,
			metrics: map[string]int{
				"win": 1, "home_win": 0, "goals": 5, "home_goals_scored": 0,
				"shutout": 0, "overtime": 0, "shootout": 0, "p2_last_minute_goals": 0,
			},
		},
		{
			name: "home overtime win", team: "ANA", date: "2026-10-04",
			opponent: "Panthers", home: true, won: true,
			metrics: map[string]int{
				"win": 1, "home_win": 1, "goals": 3, "home_goals_scored": 3,
				"shutout": 0, "overtime": 1, "shootout": 0, "p2_last_minute_goals": 0,
			},
		},
		{
			name: "away overtime loss is not a win", team: "LAK", date: "2026-10-10",
			opponent: "Golden Knights", home: false, won: false,
			metrics: map[string]int{
				"win": 0, "home_win": 0, "goals": 1, "home_goals_scored": 0,
				"shutout": 0, "overtime": 1, "shootout": 0, "p2_last_minute_goals": 0,
			},
		},
		{
			name: "shootout loss", team: "DET", date: "2026-10-10",
			opponent: "Canadiens", home: false, won: false,
			metrics: map[string]int{
				"win": 0, "home_win": 0, "goals": 0, "home_goals_scored": 0,
				"shutout": 0, "overtime": 0, "shootout": 1, "p2_last_minute_goals": 0,
			},
		},
		{
			// 0-0 after overtime; the shootout goal the scoreboard credits is
			// not a goal.
			name: "shootout win is a home win, not a shutout", team: "MTL", date: "2026-10-10",
			opponent: "Red Wings", home: true, won: true,
			metrics: map[string]int{
				"win": 1, "home_win": 1, "goals": 0, "home_goals_scored": 0,
				"shutout": 0, "overtime": 0, "shootout": 1, "p2_last_minute_goals": 0,
			},
		},
		{
			name: "home loss", team: "ANA", date: "2026-11-14",
			opponent: "Islanders", home: true, won: false,
			metrics: map[string]int{
				"win": 0, "home_win": 0, "goals": 2, "home_goals_scored": 2,
				"shutout": 0, "overtime": 0, "shootout": 0, "p2_last_minute_goals": 0,
			},
		},
		{
			name: "goal in the last minute of the 2nd (19:53)", team: "NYI", date: "2026-10-10",
			opponent: "Lightning", home: true, won: false,
			metrics: map[string]int{
				"win": 0, "home_win": 0, "goals": 3, "home_goals_scored": 3,
				"shutout": 0, "overtime": 0, "shootout": 0, "p2_last_minute_goals": 1,
			},
		},
		{name: "game not final", team: "ANA", date: "2026-11-15", wantNil: true},
		{name: "no game that day", team: "ANA", date: "2026-11-13", wantNil: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src := fixtureServer(t, nil)
			got, err := src.GetGameByDate(context.Background(), tc.team, day(t, tc.date))
			require.NoError(t, err)
			if tc.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tc.team, got.TeamID)
			assert.Equal(t, tc.opponent, got.Opponent)
			assert.Equal(t, tc.home, got.HomeGame)
			assert.Equal(t, tc.won, got.Won)
			assert.True(t, strings.HasPrefix(got.GameID, "nhl-"))
			assert.Equal(t, tc.metrics, got.Metrics)
		})
	}
}

func TestGetGameByDate_Postponed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"games":[{"id":1,"gameState":"OFF","gameScheduleState":"PPD",
			"awayTeam":{"abbrev":"ANA","score":0},"homeTeam":{"abbrev":"SJS","score":0}}]}`))
	}))
	defer srv.Close()
	src := NewSourceWithClient(srv.URL, srv.Client())
	got, err := src.GetGameByDate(context.Background(), "ANA", day(t, "2026-11-01"))
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetGameByDate_UnknownTeam(t *testing.T) {
	src := fixtureServer(t, nil)
	got, err := src.GetGameByDate(context.Background(), "XXX", day(t, "2026-11-08"))
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "unknown team ID")
}

func TestGetGameByDate_CachesPerDate(t *testing.T) {
	var hits atomic.Int32
	src := fixtureServer(t, &hits)
	for _, team := range []string{"ANA", "LAK", "DET"} {
		_, err := src.GetGameByDate(context.Background(), team, day(t, "2026-11-14"))
		require.NoError(t, err)
	}
	// The date and the next one (for games filed under a later local date).
	assert.Equal(t, int32(2), hits.Load())

	_, err := src.GetGameByDate(context.Background(), "ANA", day(t, "2026-11-08"))
	require.NoError(t, err)
	assert.Equal(t, int32(4), hits.Load())
}

func TestGetGameByDate_RetriesTransientErrors(t *testing.T) {
	var hits atomic.Int32
	b, err := os.ReadFile(filepath.Join(fixtureDir, "nhl_score_2026-11-08.json"))
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write(b)
	}))
	defer srv.Close()

	// Reuse production wiring but with a fast backoff.
	src := NewSource()
	src.baseURL = srv.URL
	src.client.Transport = fastRetry()
	got, err := src.GetGameByDate(context.Background(), "ANA", day(t, "2026-11-08"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, got.Metrics["home_win"])
	assert.Equal(t, int32(3), hits.Load(), "one retry, then the next date")
}

func TestGetGameByDate_ServerErrorSurfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	src := NewSourceWithClient(srv.URL, srv.Client())
	_, err := src.GetGameByDate(context.Background(), "ANA", day(t, "2026-11-08"))
	require.Error(t, err)
}

func TestLeagueAndRegistration(t *testing.T) {
	assert.Equal(t, "nhl", NewSource().League())
}

// Concurrent lookups of one date share a single request, and a slow date does
// not block another.
func TestGamesSharedAndConcurrent(t *testing.T) {
	release := make(chan struct{})
	var slow, other atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "2026-11-14") {
			slow.Add(1)
			<-release
		} else {
			other.Add(1)
		}
		_, _ = w.Write([]byte(`{"games":[]}`))
	}))
	t.Cleanup(srv.Close)
	s := NewSourceWithClient(srv.URL, srv.Client())

	done := make(chan struct{}, 4)
	for range 4 {
		go func() {
			_, err := s.games(t.Context(), "2026-11-14")
			assert.NoError(t, err)
			done <- struct{}{}
		}()
	}
	require.Eventually(t, func() bool { return slow.Load() == 1 }, 2*time.Second, 5*time.Millisecond)
	_, err := s.games(t.Context(), "2026-11-20") // returns while the slow one is held
	require.NoError(t, err)
	close(release)
	for range 4 {
		<-done
	}
	assert.EqualValues(t, 1, slow.Load())
	assert.EqualValues(t, 1, other.Load())
}

// Preseason games never count: UTA at LAK on 2026-09-22 is gameType 1 and
// FINAL (real response, trimmed).
func TestGetGameByDate_PreseasonNeverCounts(t *testing.T) {
	src := fixtureServer(t, nil)
	got, err := src.GetGameByDate(context.Background(), "LAK", day(t, "2026-09-22"))
	require.NoError(t, err)
	assert.Nil(t, got)

	sched, err := src.GamesOn(context.Background(), day(t, "2026-09-22"))
	require.NoError(t, err)
	assert.Empty(t, sched, "a preseason game is not a game today")
}

// All-Star (gameType 4) and international events (the 2025 4 Nations
// Face-Off was gameType 19) are not NHL games for deals.
func TestGetGameByDate_ExhibitionsNeverCount(t *testing.T) {
	for _, gt := range []string{"4", "19"} {
		t.Run(gt, func(t *testing.T) {
			src := staticServer(t, `{"games":[{"id":1,"gameType":`+gt+`,"gameDate":"2026-02-01",
				"gameState":"OFF","gameScheduleState":"OK",
				"awayTeam":{"abbrev":"LAK","score":5},"homeTeam":{"abbrev":"ANA","score":1},
				"goals":[{"period":1,"periodDescriptor":{"periodType":"REG"},"teamAbbrev":"LAK","timeInPeriod":"01:00"}]}]}`)
			got, err := src.GetGameByDate(context.Background(), "LAK", day(t, "2026-02-01"))
			require.NoError(t, err)
			assert.Nil(t, got)
		})
	}
}

// SJS won 6-5 at UTA on 2026-11-14 in a shootout: 5-5 plus the shootout
// winner the scoreboard credits as a goal. Goal metrics count real goals only.
func TestGetGameByDate_ShootoutGoalIsNotAGoal(t *testing.T) {
	src := fixtureServer(t, nil)

	sharks, err := src.GetGameByDate(context.Background(), "SJS", day(t, "2026-11-14"))
	require.NoError(t, err)
	require.NotNil(t, sharks)
	assert.True(t, sharks.Won)
	assert.Equal(t, map[string]int{
		"win": 1, "home_win": 0, "goals": 5, "home_goals_scored": 0,
		"shutout": 0, "overtime": 0, "shootout": 1, "p2_last_minute_goals": 0,
	}, sharks.Metrics)

	mammoth, err := src.GetGameByDate(context.Background(), "UTA", day(t, "2026-11-14"))
	require.NoError(t, err)
	require.NotNil(t, mammoth)
	assert.Equal(t, 5, mammoth.Metrics["goals"])
	assert.Equal(t, 5, mammoth.Metrics["home_goals_scored"])
	assert.Equal(t, 0, mammoth.Metrics["win"])
}

// A 200 with an error body or no games key is an error, not "no games", and
// it is not cached.
func TestGetGameByDate_BadBodyIsAnError(t *testing.T) {
	for _, bad := range []string{`{}`, `{"message":"Internal error"}`, `{"games":null}`} {
		t.Run(bad, func(t *testing.T) {
			var hits atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if hits.Add(1) == 1 {
					_, _ = w.Write([]byte(bad))
					return
				}
				_, _ = w.Write([]byte(`{"games":[]}`))
			}))
			t.Cleanup(srv.Close)
			src := NewSourceWithClient(srv.URL, srv.Client())
			_, err := src.GetGameByDate(context.Background(), "ANA", day(t, "2026-11-08"))
			require.Error(t, err)

			got, err := src.GetGameByDate(context.Background(), "ANA", day(t, "2026-11-08"))
			require.NoError(t, err, "the bad body was not cached")
			assert.Nil(t, got)
		})
	}
}

// The NHL files a game under its local date. A 1pm matinee in Tokyo on
// Sunday Nov 22 is 04:00Z, which is Saturday Nov 21 in Los Angeles: the deal
// date is the Pacific one.
func TestGetGameByDate_OverseasGameUsesPacificDate(t *testing.T) {
	src := byDay(t, map[string]string{
		"2026-11-22": `{"games":[{"id":7,"gameType":2,"gameDate":"2026-11-22",
			"startTimeUTC":"2026-11-22T04:00:00Z","gameState":"OFF","gameScheduleState":"OK",
			"awayTeam":{"abbrev":"LAK","score":3,"name":{"default":"Kings"}},
			"homeTeam":{"abbrev":"NYI","score":1,"name":{"default":"Islanders"}},
			"goals":[
			 {"period":1,"periodDescriptor":{"periodType":"REG"},"teamAbbrev":"LAK","timeInPeriod":"01:00"},
			 {"period":2,"periodDescriptor":{"periodType":"REG"},"teamAbbrev":"LAK","timeInPeriod":"02:00"},
			 {"period":3,"periodDescriptor":{"periodType":"REG"},"teamAbbrev":"LAK","timeInPeriod":"03:00"},
			 {"period":3,"periodDescriptor":{"periodType":"REG"},"teamAbbrev":"NYI","timeInPeriod":"04:00"}]}]}`,
	})

	got, err := src.GetGameByDate(context.Background(), "LAK", day(t, "2026-11-21"))
	require.NoError(t, err)
	require.NotNil(t, got, "the game started on Pacific Nov 21")
	assert.Equal(t, "nhl-7", got.GameID)
	assert.Equal(t, 3, got.Metrics["goals"])

	got, err = src.GetGameByDate(context.Background(), "LAK", day(t, "2026-11-22"))
	require.NoError(t, err)
	assert.Nil(t, got, "and not on Pacific Nov 22, the NHL's local date")
}

func TestGetGameByDate_404IsAnError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	_, err := NewSourceWithClient(srv.URL, srv.Client()).GetGameByDate(context.Background(), "ANA", day(t, "2020-01-01"))
	require.Error(t, err, "upstream 404 is an error, not 'no game'")
}

func TestGetGamesByDate(t *testing.T) {
	src := fixtureServer(t, nil)
	games, err := src.GetGamesByDate(context.Background(), "ANA", day(t, "2026-11-08"))
	require.NoError(t, err)
	require.Len(t, games, 1)
	assert.Equal(t, "Penguins", games[0].Opponent)
}

// staticServer answers every date with body.
func staticServer(t *testing.T, body string) *Source {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewSourceWithClient(srv.URL, srv.Client())
}

// byDay serves bodies[date] for /score/{date} and an empty day otherwise.
func byDay(t *testing.T, bodies map[string]string) *Source {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, ok := bodies[strings.TrimPrefix(r.URL.Path, "/score/")]
		if !ok {
			b = `{"games":[]}`
		}
		_, _ = w.Write([]byte(b))
	}))
	t.Cleanup(srv.Close)
	return NewSourceWithClient(srv.URL, srv.Client())
}

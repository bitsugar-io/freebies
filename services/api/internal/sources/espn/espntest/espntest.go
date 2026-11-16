// Package espntest serves ESPN API fixtures (test/fixtures/espn_*.json,
// trimmed from real responses) and runs the table of expected games every
// league source test shares.
package espntest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

// FixtureDir is the directory holding the recorded responses.
func FixtureDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "test", "fixtures")
}

// Server serves
//
//	/{sport}/{league}/scoreboard?dates=YYYYMMDD  from espn_{sport}_{league}_{date}.json
//	/{sport}/{league}/summary?event=ID           from espn_{sport}_{league}_summary_{id}.json
//
// A scoreboard with no fixture is an empty day (no games), a summary with no
// fixture is a 404. hits counts every request when non-nil.
func Server(t *testing.T, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			hits.Add(1)
		}
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		p := strings.Trim(r.URL.Path, "/")
		var name string
		switch {
		case strings.HasSuffix(p, "/scoreboard"):
			name = "espn_" + strings.ReplaceAll(strings.TrimSuffix(p, "/scoreboard"), "/", "_") +
				"_" + r.URL.Query().Get("dates") + ".json"
		case strings.HasSuffix(p, "/summary"):
			name = "espn_" + strings.ReplaceAll(strings.TrimSuffix(p, "/summary"), "/", "_") +
				"_summary_" + r.URL.Query().Get("event") + ".json"
		default:
			http.NotFound(w, r)
			return
		}
		b, err := os.ReadFile(filepath.Join(FixtureDir(), name))
		if err != nil {
			if strings.Contains(name, "_summary_") {
				http.NotFound(w, r)
				return
			}
			b = []byte(`{"events":[]}`)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// Day parses YYYY-MM-DD.
func Day(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	require.NoError(t, err)
	return d
}

// Case is one expected result.
type Case struct {
	Name     string
	Team     string
	Date     string // Pacific YYYY-MM-DD
	Nil      bool   // no final game expected
	Opponent string
	Home     bool
	Won      bool
	Metrics  map[string]int
}

// Run checks every case against src and that the union of metrics produced
// equals what src.Metrics() declares, so the declared list cannot drift from
// what Build emits.
func Run(t *testing.T, src sources.Source, league string, cases []Case) {
	t.Helper()
	produced := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := src.GetGameByDate(context.Background(), tc.Team, Day(t, tc.Date))
			require.NoError(t, err)
			if tc.Nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tc.Team, got.TeamID)
			assert.Equal(t, tc.Opponent, got.Opponent)
			assert.Equal(t, tc.Home, got.HomeGame, "HomeGame")
			assert.Equal(t, tc.Won, got.Won, "Won")
			assert.True(t, strings.HasPrefix(got.GameID, "espn-"+league+"-"), got.GameID)
			assert.Equal(t, tc.Metrics, got.Metrics)
			for k := range got.Metrics {
				produced[k] = true
			}
		})
	}

	var got, want []string
	for k := range produced {
		got = append(got, k)
	}
	want = append(want, src.Metrics()...)
	sort.Strings(got)
	sort.Strings(want)
	assert.Equal(t, want, got, "Metrics() must list exactly what the source emits")
	assert.Equal(t, league, src.League())
}

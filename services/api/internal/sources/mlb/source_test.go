package mlb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mlbsdk "github.com/retr0h/mlb-sdk/pkg/mlb"
)

// fakeMLB wires a single httptest server to serve both /schedule and
// /game/{pk}/boxscore responses keyed on path. Either body may be "" to
// produce a 200 with empty body, or status may be non-200 to simulate
// an upstream error.
type fakeMLB struct {
	scheduleStatus int
	scheduleBody   string
	boxscoreStatus int
	boxscoreBody   string
}

func (f fakeMLB) server() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/schedule"):
				w.WriteHeader(f.scheduleStatus)
				_, _ = w.Write([]byte(f.scheduleBody))
			case strings.Contains(r.URL.Path, "/boxscore"):
				w.WriteHeader(f.boxscoreStatus)
				_, _ = w.Write([]byte(f.boxscoreBody))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
}

func TestSource_GetGameByDate(t *testing.T) {
	const happySchedule = `{
		"dates": [{"games": [{
			"gamePk": 823957,
			"gameDate": "2026-05-08T19:10:00Z",
			"status": {"abstractGameState": "Final"},
			"teams": {
				"home": {"team": {"id": 119, "name": "Los Angeles Dodgers"}, "score": 3},
				"away": {"team": {"id": 144, "name": "Atlanta Braves"},      "score": 1}
			}
		}]}]
	}`
	const happyBoxscore = `{
		"teams": {
			"home": {
				"team": {"id": 119, "name": "Los Angeles Dodgers"},
				"teamStats": {
					"pitching": {"strikeOuts": 11, "hits": 5, "runs": 1, "homeRuns": 0, "baseOnBalls": 3},
					"batting":  {"runs": 3, "hits": 9, "homeRuns": 2, "rbi": 3, "stolenBases": 1}
				},
				"info": [{"title": "FIELDING", "fieldList": [{"label": "DP", "value": "2 (a; b)."}]}]
			},
			"away": {
				"team": {"id": 144, "name": "Atlanta Braves"},
				"teamStats": {
					"pitching": {"strikeOuts": 7, "hits": 9, "runs": 3, "homeRuns": 2, "baseOnBalls": 4},
					"batting":  {"runs": 1, "hits": 5, "homeRuns": 0, "rbi": 1, "stolenBases": 0}
				}
			}
		}
	}`

	date := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		teamID      string
		fake        fakeMLB
		wantNil     bool
		wantErr     string
		wantOpp     string
		wantHome    bool
		wantWon     bool
		wantMetrics map[string]int // subset; only listed keys checked
	}{
		{
			name:   "Dodgers home win — full metrics",
			teamID: "LAD",
			fake: fakeMLB{
				scheduleStatus: 200, scheduleBody: happySchedule,
				boxscoreStatus: 200, boxscoreBody: happyBoxscore,
			},
			wantOpp:  "Atlanta Braves",
			wantHome: true,
			wantWon:  true,
			wantMetrics: map[string]int{
				"strikeouts":        11,
				"pitching_walks":    3,
				"runs":              3,
				"home_runs":         2,
				"stolen_bases":      1,
				"double_plays":      2,
				"home_double_plays": 2,
				"win":               1,
				"home_win":          1,
			},
		},
		{
			name:   "Braves away loss — same boxscore, opposite side",
			teamID: "ATL",
			fake: fakeMLB{
				scheduleStatus: 200, scheduleBody: happySchedule,
				boxscoreStatus: 200, boxscoreBody: happyBoxscore,
			},
			wantOpp:  "Los Angeles Dodgers",
			wantHome: false,
			wantWon:  false,
			wantMetrics: map[string]int{
				"strikeouts":        7,
				"runs":              1,
				"double_plays":      0,
				"home_runs_scored":  0, // away game ⇒ home_* are zero
				"home_stolen_bases": 0,
				"home_double_plays": 0,
				"win":               0,
				"home_win":          0,
			},
		},
		{
			name:    "unknown team",
			teamID:  "XXX",
			wantErr: "unknown team ID",
		},
		{
			name:   "no Final game on date returns nil, nil",
			teamID: "LAD",
			fake: fakeMLB{
				scheduleStatus: 200,
				scheduleBody: `{"dates":[{"games":[{
					"gamePk": 1, "status": {"abstractGameState": "Live"},
					"teams": {"home":{"team":{"id":119}},"away":{"team":{"id":144}}}
				}]}]}`,
			},
			wantNil: true,
		},
		{
			name:   "schedule HTTP error is wrapped",
			teamID: "LAD",
			fake: fakeMLB{
				scheduleStatus: 500, scheduleBody: `oops`,
			},
			wantErr: "fetching schedule",
		},
		{
			name:   "boxscore HTTP error is wrapped",
			teamID: "LAD",
			fake: fakeMLB{
				scheduleStatus: 200, scheduleBody: happySchedule,
				boxscoreStatus: 500, boxscoreBody: `oops`,
			},
			wantErr: "fetching boxscore",
		},
		{
			name:   "team missing from boxscore",
			teamID: "LAD",
			fake: fakeMLB{
				scheduleStatus: 200, scheduleBody: happySchedule,
				boxscoreStatus: 200,
				boxscoreBody: `{"teams":{
					"home":{"team":{"id": 999}},
					"away":{"team":{"id": 888}}
				}}`,
			},
			wantErr: "not present in boxscore",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var src *Source
			if c.fake != (fakeMLB{}) {
				srv := c.fake.server()
				defer srv.Close()
				src = NewSourceWithClient(mlbsdk.New(mlbsdk.WithBaseURL(srv.URL)))
			} else {
				src = NewSource() // unknown-team path never reaches HTTP
			}

			got, err := src.GetGameByDate(context.Background(), c.teamID, date)

			if c.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", c.wantErr)
				}
				if !strings.Contains(err.Error(), c.wantErr) {
					t.Errorf("err = %v, want substring %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantNil {
				if got != nil {
					t.Errorf("expected nil GameStats, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil GameStats")
			}
			if got.Opponent != c.wantOpp {
				t.Errorf("Opponent = %q, want %q", got.Opponent, c.wantOpp)
			}
			if got.HomeGame != c.wantHome {
				t.Errorf("HomeGame = %v, want %v", got.HomeGame, c.wantHome)
			}
			if got.Won != c.wantWon {
				t.Errorf("Won = %v, want %v", got.Won, c.wantWon)
			}
			for k, want := range c.wantMetrics {
				if got.Metrics[k] != want {
					t.Errorf("Metrics[%q] = %d, want %d", k, got.Metrics[k], want)
				}
			}
		})
	}
}

func TestSource_League(t *testing.T) {
	if got := NewSource().League(); got != "mlb" {
		t.Errorf("League() = %q, want %q", got, "mlb")
	}
}

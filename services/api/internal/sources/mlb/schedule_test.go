package mlb

import (
	"context"
	"net/http"
	"testing"
	"time"

	mlbsdk "github.com/retr0h/mlb-sdk/pkg/mlb"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

func TestSource_GamesOn(t *testing.T) {
	const schedule = `{"dates": [{"games": [
		{"gamePk": 1, "gameDate": "2026-05-09T02:10:00Z",
		 "status": {"abstractGameState": "Preview", "detailedState": "Scheduled"},
		 "teams": {"home": {"team": {"id": 119, "name": "Los Angeles Dodgers"}},
		           "away": {"team": {"id": 144, "name": "Atlanta Braves"}}}},
		{"gamePk": 2, "gameDate": "2026-05-08T20:07:00Z",
		 "status": {"abstractGameState": "Live", "detailedState": "In Progress"},
		 "teams": {"home": {"team": {"id": 141, "name": "Toronto Blue Jays"}},
		           "away": {"team": {"id": 108, "name": "Los Angeles Angels"}}}},
		{"gamePk": 3, "gameDate": "2026-05-08T23:05:00Z",
		 "status": {"abstractGameState": "Final", "detailedState": "Postponed"},
		 "teams": {"home": {"team": {"id": 137, "name": "San Francisco Giants"}},
		           "away": {"team": {"id": 135, "name": "San Diego Padres"}}}}
	]}]}`
	srv := fakeMLB{scheduleStatus: http.StatusOK, scheduleBody: schedule}.server()
	t.Cleanup(srv.Close)
	src := NewSourceWithClient(mlbsdk.New(mlbsdk.WithBaseURL(srv.URL)))

	got, err := src.GamesOn(context.Background(), time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	want := []sources.ScheduledGame{
		{TeamID: "TOR", Opponent: "Los Angeles Angels", StartsAt: time.Date(2026, 5, 8, 20, 7, 0, 0, time.UTC), Home: true, State: "in"},
		{TeamID: "LAA", Opponent: "Toronto Blue Jays", StartsAt: time.Date(2026, 5, 8, 20, 7, 0, 0, time.UTC), Home: false, State: "in"},
		{TeamID: "LAD", Opponent: "Atlanta Braves", StartsAt: time.Date(2026, 5, 9, 2, 10, 0, 0, time.UTC), Home: true, State: "pre"},
		{TeamID: "ATL", Opponent: "Los Angeles Dodgers", StartsAt: time.Date(2026, 5, 9, 2, 10, 0, 0, time.UTC), Home: false, State: "pre"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("game %d: got %+v, want %+v", i, got[i], want[i])
		}
	}

	srv500 := fakeMLB{scheduleStatus: http.StatusInternalServerError}.server()
	t.Cleanup(srv500.Close)
	if _, err := NewSourceWithClient(mlbsdk.New(mlbsdk.WithBaseURL(srv500.URL))).GamesOn(
		context.Background(), time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("expected an error from a failing schedule")
	}
}

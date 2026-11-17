package handlers

import (
	"database/sql"
	"testing"

	"github.com/retr0h/freebie/services/api/internal/db"
)

func ruleEvent(id, metric string, inVenue int64) db.Event {
	return db.Event{
		ID:          id,
		TriggerRule: sql.NullString{String: `{"metric": "` + metric + `"}`, Valid: true},
		InVenue:     inVenue,
	}
}

func ids(evs []db.Event) []string {
	out := make([]string, len(evs))
	for i, e := range evs {
		out[i] = e.ID
	}
	return out
}

func TestDealsForVenueDropsHomeOnlyDealsOnTheRoad(t *testing.T) {
	ducks := []db.Event{
		ruleEvent("mcgriddles", "home_win", 0),
		ruleEvent("cfa-at-game", "home_goals_scored", 1),
		ruleEvent("ucla-football", "football_home_points", 0),
		ruleEvent("kings-mcflurry", "p2_last_minute_goals", 0),
		ruleEvent("venue-any-metric", "win", 1),
	}
	if got := ids(dealsForVenue(ducks, false)); len(got) != 1 || got[0] != "kings-mcflurry" {
		t.Errorf("road game deals = %v, want only kings-mcflurry", got)
	}
	if got := dealsForVenue(ducks, true); len(got) != len(ducks) {
		t.Errorf("home game kept %d deals, want all %d", len(got), len(ducks))
	}
}

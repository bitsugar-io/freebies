package nwsl

import (
	"testing"

	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
)

func TestGetGameByDate(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := NewSourceWithClient(srv.URL, srv.Client())

	espntest.Run(t, src, "nwsl", []espntest.Case{
		{
			Name: "home win, scored first", Team: "ACFC", Date: "2026-11-01",
			Opponent: "Boston Legacy FC", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "draw": 0, "goals": 4, "home_goals_scored": 4,
				"first_half_goals": 1, "home_first_half_goals": 1,
				"scored_first_in_first_half": 1, "home_scored_first_in_first_half": 1,
			},
		},
		{
			Name: "away win, three first-half goals but conceded first", Team: "ACFC", Date: "2026-10-17",
			Opponent: "Denver Summit FC", Home: false, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 0, "draw": 0, "goals": 3, "home_goals_scored": 0,
				"first_half_goals": 3, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "home 0-0 draw", Team: "ACFC", Date: "2026-08-23",
			Opponent: "Gotham FC", Home: true, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 1, "goals": 0, "home_goals_scored": 0,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "away loss", Team: "ACFC", Date: "2026-08-26",
			Opponent: "North Carolina Courage", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 0, "goals": 0, "home_goals_scored": 0,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "home 3-3 draw, 45th-minute goal is first half", Team: "ACFC", Date: "2026-09-11",
			Opponent: "Denver Summit FC", Home: true, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 1, "goals": 3, "home_goals_scored": 3,
				"first_half_goals": 1, "home_first_half_goals": 1,
				"scored_first_in_first_half": 1, "home_scored_first_in_first_half": 1,
			},
		},
		{
			Name: "playoff loss after extra time (STATUS_FINAL_AET)", Team: "ACFC", Date: "2026-11-07",
			Opponent: "Kansas City Current", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 0, "goals": 0, "home_goals_scored": 0,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{Name: "no game that day", Team: "ACFC", Date: "2026-11-05", Nil: true},
	})
}

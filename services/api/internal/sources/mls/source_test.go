package mls

import (
	"testing"

	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
)

func TestGetGameByDate(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := NewSourceWithClient(srv.URL, srv.Client())

	espntest.Run(t, src, "mls", []espntest.Case{
		{
			Name: "home win, scored first in the first half (the LAFC deal)", Team: "LAFC", Date: "2026-09-09",
			Opponent: "Red Bull New York", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "draw": 0, "goals": 2, "home_goals_scored": 2,
				"first_half_goals": 1, "home_first_half_goals": 1,
				"scored_first_in_first_half": 1, "home_scored_first_in_first_half": 1,
			},
		},
		{
			Name: "home draw, goal at 46 is a second-half goal", Team: "LAFC", Date: "2026-10-10",
			Opponent: "Vancouver Whitecaps", Home: true, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 1, "goals": 1, "home_goals_scored": 1,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "home win with no first-half goal", Team: "LAFC", Date: "2026-10-14",
			Opponent: "Austin FC", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "draw": 0, "goals": 3, "home_goals_scored": 3,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "away draw, 45+1 stoppage goal counts as first half but not at home", Team: "LAFC", Date: "2026-10-17",
			Opponent: "New York City FC", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 1, "goals": 1, "home_goals_scored": 0,
				"first_half_goals": 1, "home_first_half_goals": 0,
				"scored_first_in_first_half": 1, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "home draw, opponent scored first", Team: "LAFC", Date: "2026-10-25",
			Opponent: "LA Galaxy", Home: true, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 1, "goals": 2, "home_goals_scored": 2,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "the Galaxy in that game: away, scored first", Team: "LAG", Date: "2026-10-25",
			Opponent: "LAFC", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 1, "goals": 2, "home_goals_scored": 0,
				"first_half_goals": 1, "home_first_half_goals": 0,
				"scored_first_in_first_half": 1, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "home loss", Team: "LAFC", Date: "2026-11-01",
			Opponent: "Minnesota United FC", Home: true, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 0, "goals": 1, "home_goals_scored": 1,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "away loss", Team: "LAFC", Date: "2026-11-07",
			Opponent: "Seattle Sounders FC", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 0, "goals": 1, "home_goals_scored": 0,
				"first_half_goals": 0, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "own goal counts for the team it benefits", Team: "LAG", Date: "2026-11-07",
			Opponent: "Real Salt Lake", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "draw": 0, "goals": 1, "home_goals_scored": 0,
				"first_half_goals": 1, "home_first_half_goals": 0,
				"scored_first_in_first_half": 0, "home_scored_first_in_first_half": 0,
			},
		},
		{
			Name: "home win for the Galaxy, both goals", Team: "LAG", Date: "2026-10-14",
			Opponent: "Portland Timbers", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "draw": 0, "goals": 2, "home_goals_scored": 2,
				"first_half_goals": 1, "home_first_half_goals": 1,
				"scored_first_in_first_half": 1, "home_scored_first_in_first_half": 1,
			},
		},
		{
			Name: "away win, two first-half goals incl. stoppage time, scored first", Team: "LAFC", Date: "2026-07-17",
			Opponent: "LA Galaxy", Home: false, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 0, "draw": 0, "goals": 3, "home_goals_scored": 0,
				"first_half_goals": 2, "home_first_half_goals": 0,
				"scored_first_in_first_half": 1, "home_scored_first_in_first_half": 0,
			},
		},
		{Name: "no game that day", Team: "LAFC", Date: "2026-11-05", Nil: true},
	})
}

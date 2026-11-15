package wnba

import (
	"testing"

	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
)

func TestGetGameByDate(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := NewSourceWithClient(srv.URL, srv.Client())

	espntest.Run(t, src, "wnba", []espntest.Case{
		{
			Name: "home win", Team: "LAS", Date: "2026-09-20",
			Opponent: "Portland Fire", Home: true, Won: true,
			Metrics: map[string]int{"win": 1, "home_win": 1, "points": 105, "home_points": 105},
		},
		{
			Name: "away win is not a home win", Team: "LAS", Date: "2026-08-30",
			Opponent: "Seattle Storm", Home: false, Won: true,
			Metrics: map[string]int{"win": 1, "home_win": 0, "points": 90, "home_points": 0},
		},
		{
			Name: "away loss", Team: "LAS", Date: "2026-09-17",
			Opponent: "Dallas Wings", Home: false, Won: false,
			Metrics: map[string]int{"win": 0, "home_win": 0, "points": 77, "home_points": 0},
		},
		{
			Name: "one-point home loss", Team: "LAS", Date: "2026-09-24",
			Opponent: "Golden State Valkyries", Home: true, Won: false,
			Metrics: map[string]int{"win": 0, "home_win": 0, "points": 75, "home_points": 75},
		},
		{Name: "no game that day", Team: "LAS", Date: "2026-09-18", Nil: true},
	})
}

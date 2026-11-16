package nfl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
)

func TestGetGameByDate(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := NewSourceWithClient(srv.URL, srv.Client())

	espntest.Run(t, src, "nfl", []espntest.Case{
		{
			Name: "home win, one interception", Team: "LAR", Date: "2026-09-21",
			Opponent: "New York Giants", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "points": 28, "home_points": 28, "interceptions": 1,
				"long_field_goal": 0,
			},
		},
		{
			Name: "home win, two interceptions", Team: "LAR", Date: "2026-11-01",
			Opponent: "Los Angeles Chargers", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "points": 24, "home_points": 24, "interceptions": 2,
				"long_field_goal": 30,
			},
		},
		{
			Name: "away win, no interception", Team: "LAR", Date: "2026-11-08",
			Opponent: "Washington Commanders", Home: false, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 0, "points": 20, "home_points": 0, "interceptions": 0,
				"long_field_goal": 33,
			},
		},
		{
			Name: "the other side of that game", Team: "LACH", Date: "2026-11-01",
			Opponent: "Los Angeles Rams", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "points": 9, "home_points": 0, "interceptions": 1,
				"long_field_goal": 46,
			},
		},
		{
			Name: "home loss", Team: "LAR", Date: "2026-10-12",
			Opponent: "Buffalo Bills", Home: true, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "points": 34, "home_points": 34, "interceptions": 1,
				"long_field_goal": 0,
			},
		},
		{
			Name: "neutral site is not a home game", Team: "LAR", Date: "2026-09-10",
			Opponent: "San Francisco 49ers", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "points": 7, "home_points": 0, "interceptions": 1,
				"long_field_goal": 0,
			},
		},
		{Name: "no game that day (bye week)", Team: "LAR", Date: "2026-11-15", Nil: true},
	})
}

func TestSummaryMissingFails(t *testing.T) {
	// The Chargers' game on 2026-09-27 has a scoreboard fixture but no
	// summary fixture: a missing box score must be an error, never a silent
	// interceptions == 0.
	srv := espntest.Server(t, nil)
	src := NewSourceWithClient(srv.URL, srv.Client())
	got, err := src.GetGameByDate(context.Background(), "LACH", espntest.Day(t, "2026-09-27"))
	require.Error(t, err)
	assert.Nil(t, got)
}

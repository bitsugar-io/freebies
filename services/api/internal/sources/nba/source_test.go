package nba

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

	espntest.Run(t, src, "nba", []espntest.Case{
		{
			Name: "home win", Team: "LAL", Date: "2026-11-06",
			Opponent: "Portland Trail Blazers", Home: true, Won: true,
			Metrics: map[string]int{"win": 1, "home_win": 1, "points": 124, "home_points": 124, "home_q4_opponent_ft_miss_streak": 1},
		},
		{
			Name: "away win is not a home win", Team: "LAL", Date: "2026-11-09",
			Opponent: "Atlanta Hawks", Home: false, Won: true,
			Metrics: map[string]int{"win": 1, "home_win": 0, "points": 124, "home_points": 0, "home_q4_opponent_ft_miss_streak": 0},
		},
		{
			Name: "home loss", Team: "LAL", Date: "2026-10-27",
			Opponent: "Portland Trail Blazers", Home: true, Won: false,
			Metrics: map[string]int{"win": 0, "home_win": 0, "points": 129, "home_points": 129, "home_q4_opponent_ft_miss_streak": 3},
		},
		{
			Name: "away loss, same game as the Lakers' home win", Team: "LAC", Date: "2026-10-23",
			Opponent: "Los Angeles Lakers", Home: false, Won: false,
			Metrics: map[string]int{"win": 0, "home_win": 0, "points": 117, "home_points": 0, "home_q4_opponent_ft_miss_streak": 0},
		},
		{
			Name: "the other team in that game", Team: "LAL", Date: "2026-10-23",
			Opponent: "LA Clippers", Home: true, Won: true,
			Metrics: map[string]int{"win": 1, "home_win": 1, "points": 125, "home_points": 125, "home_q4_opponent_ft_miss_streak": 1},
		},
		{
			Name: "Clippers home game, visitors never miss two in a row", Team: "LAC", Date: "2026-11-13",
			Opponent: "Oklahoma City Thunder", Home: true, Won: true,
			Metrics: map[string]int{"win": 1, "home_win": 1, "points": 113, "home_points": 113, "home_q4_opponent_ft_miss_streak": 1},
		},
		{Name: "game not final", Team: "LAL", Date: "2026-11-15", Nil: true},
		{Name: "no game that day", Team: "LAL", Date: "2026-11-13", Nil: true},
	})
}

func TestUnknownTeam(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := NewSourceWithClient(srv.URL, srv.Client())
	got, err := src.GetGameByDate(context.Background(), "XXX", espntest.Day(t, "2026-11-06"))
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "unknown team ID")
}

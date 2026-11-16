package ncaa

import (
	"testing"

	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
)

func TestGetGameByDate(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := NewSourceWithClient(srv.URL, srv.Client())

	espntest.Run(t, src, "ncaa", []espntest.Case{
		{
			Name: "football home win", Team: "USC", Date: "2026-10-31",
			Opponent: "Ohio State Buckeyes", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "points": 30, "home_points": 30,
				"football_win": 1, "football_home_win": 1, "football_points": 30, "football_home_points": 30,
			},
		},
		{
			Name: "football away loss", Team: "USC", Date: "2026-11-14",
			Opponent: "Indiana Hoosiers", Home: false, Won: false,
			Metrics: map[string]int{
				"win": 0, "home_win": 0, "points": 13, "home_points": 0,
				"football_win": 0, "football_home_win": 0, "football_points": 13, "football_home_points": 0,
			},
		},
		{
			Name: "football away win", Team: "UCLA", Date: "2026-11-07",
			Opponent: "Minnesota Golden Gophers", Home: false, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 0, "points": 34, "home_points": 0,
				"football_win": 1, "football_home_win": 0, "football_points": 34, "football_home_points": 0,
			},
		},
		{
			Name: "football Friday-night home win", Team: "UCLA", Date: "2026-11-13",
			Opponent: "Illinois Fighting Illini", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "points": 22, "home_points": 22,
				"football_win": 1, "football_home_win": 1, "football_points": 22, "football_home_points": 22,
			},
		},
		{
			Name: "basketball home win", Team: "USC", Date: "2026-11-02",
			Opponent: "Milwaukee Panthers", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "points": 95, "home_points": 95,
				"basketball_win": 1, "basketball_home_win": 1, "basketball_points": 95, "basketball_home_points": 95,
				"basketball_home_h2_opponent_ft_miss_streak": 1,
			},
		},
		{
			Name: "basketball neutral-site win is not a home win", Team: "UCLA", Date: "2026-11-02",
			Opponent: "Arizona Wildcats", Home: false, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 0, "points": 84, "home_points": 0,
				"basketball_win": 1, "basketball_home_win": 0, "basketball_points": 84, "basketball_home_points": 0,
				"basketball_home_h2_opponent_ft_miss_streak": 0,
			},
		},
		{
			Name: "basketball home win, blowout", Team: "UCLA", Date: "2026-11-11",
			Opponent: "Arkansas-Pine Bluff Golden Lions", Home: true, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 1, "points": 106, "home_points": 106,
				"basketball_win": 1, "basketball_home_win": 1, "basketball_points": 106, "basketball_home_points": 106,
				"basketball_home_h2_opponent_ft_miss_streak": 2,
			},
		},
		{
			Name: "basketball away win", Team: "USC", Date: "2026-11-11",
			Opponent: "California Golden Bears", Home: false, Won: true,
			Metrics: map[string]int{
				"win": 1, "home_win": 0, "points": 77, "home_points": 0,
				"basketball_win": 1, "basketball_home_win": 0, "basketball_points": 77, "basketball_home_points": 0,
				"basketball_home_h2_opponent_ft_miss_streak": 0,
			},
		},
		{Name: "no game that day", Team: "USC", Date: "2026-11-15", Nil: true},
	})
}

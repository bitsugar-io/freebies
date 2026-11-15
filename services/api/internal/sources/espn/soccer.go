package espn

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

// SoccerMetrics builds the soccer metrics from the scoreboard's goal events:
//
//	win, home_win                      1 if the team scored more (at home)
//	draw                               1 if the match ended level (a penalty
//	                                   shootout after a level game is a draw)
//	goals, home_goals_scored           goals scored (0 for home_ when away)
//	first_half_goals                   goals up to 45' plus first-half stoppage
//	home_first_half_goals              same, but 0 for away games
//	scored_first_in_first_half         1 if the team scored the match's first
//	                                   goal and it came in the first half
//	home_scored_first_in_first_half    same, but 0 for away games
//
// An own goal counts for the team it benefits, as on the scoreboard. If the
// goal events do not add up to the final score the game is rejected rather
// than guessed at, so a bad feed can never fire a first-half deal.
func SoccerMetrics(_ context.Context, _ *Client, g *Game) (map[string]int, error) {
	type goal struct {
		ours   bool
		second bool
		value  float64
		stop   int
	}
	var goals []goal
	ours, theirs := 0, 0
	for _, d := range g.Comp.Details {
		if !d.ScoringPlay || d.Shootout {
			continue
		}
		mins, stop, err := parseMinute(d.Clock.Display)
		if err != nil {
			return nil, err
		}
		gl := goal{ours: d.Team.ID == g.Us.Team.ID, second: mins > 45, value: d.Clock.Value, stop: stop}
		if gl.ours {
			ours++
		} else {
			theirs++
		}
		goals = append(goals, gl)
	}
	if ours != g.OurScore || theirs != g.TheirScore {
		return nil, fmt.Errorf("goal events (%d-%d) do not match the final score (%d-%d)",
			ours, theirs, g.OurScore, g.TheirScore)
	}
	sort.SliceStable(goals, func(i, j int) bool {
		a, b := goals[i], goals[j]
		if a.second != b.second {
			return !a.second
		}
		if a.value != b.value {
			return a.value < b.value
		}
		return a.stop < b.stop
	})

	firstHalf := 0
	for _, gl := range goals {
		if gl.ours && !gl.second {
			firstHalf++
		}
	}
	first := len(goals) > 0 && goals[0].ours && !goals[0].second

	m := map[string]int{
		"draw":                            boolToInt(g.OurScore == g.TheirScore),
		"goals":                           g.OurScore,
		"home_goals_scored":               conditionalInt(g.Home, g.OurScore),
		"first_half_goals":                firstHalf,
		"home_first_half_goals":           conditionalInt(g.Home, firstHalf),
		"scored_first_in_first_half":      boolToInt(first),
		"home_scored_first_in_first_half": boolToInt(first && g.Home),
	}
	g.winMetrics(m)
	return m, nil
}

var minuteRE = regexp.MustCompile(`^(\d+)'(?:\+(\d+)')?`)

// parseMinute reads ESPN's goal clock ("52'", "45'+2'", "90'+6'"). ESPN shows
// the minute in progress, so 45' and 45'+N' are first half and 46' is the
// second half's first minute.
func parseMinute(s string) (minute, stoppage int, err error) {
	m := minuteRE.FindStringSubmatch(s)
	if m == nil {
		return 0, 0, fmt.Errorf("unreadable goal minute %q", s)
	}
	minute, _ = strconv.Atoi(m[1])
	if m[2] != "" {
		stoppage, _ = strconv.Atoi(m[2])
	}
	return minute, stoppage, nil
}

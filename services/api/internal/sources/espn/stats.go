package espn

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// LongFieldGoal returns teamID's longest made field goal in yards from the box
// score's kicking category, or 0 when the team made none. A summary with no
// player box score at all is an error, never a silent 0.
func (c *Client) LongFieldGoal(ctx context.Context, path, eventID, teamID string) (int, error) {
	sum, err := c.summary(ctx, path, eventID)
	if err != nil {
		return 0, err
	}
	if len(sum.Boxscore.Players) == 0 {
		return 0, fmt.Errorf("ESPN summary %s: no player box score", eventID)
	}
	long := 0
	for _, t := range sum.Boxscore.Players {
		if t.Team.ID != teamID {
			continue
		}
		for _, cat := range t.Statistics {
			if cat.Name != "kicking" {
				continue
			}
			i := indexOf(cat.Keys, "longFieldGoalMade")
			if i < 0 {
				return 0, fmt.Errorf("ESPN summary %s: kicking has no longFieldGoalMade", eventID)
			}
			for _, a := range cat.Athletes {
				if i >= len(a.Stats) {
					continue
				}
				v := strings.TrimSpace(a.Stats[i])
				if v == "" || v == "--" {
					continue
				}
				n, err := strconv.Atoi(v)
				if err != nil {
					return 0, fmt.Errorf("ESPN summary %s: longFieldGoalMade=%q is not a number", eventID, v)
				}
				long = max(long, n)
			}
		}
	}
	return long, nil
}

// FreeThrowMissStreak returns the most free throws teamID missed in a row
// during period (a quarter in the NBA, a half in college) (a make by teamID resets the count; the other team's shots
// don't). A summary with no plays is an error, never a silent 0.
func (c *Client) FreeThrowMissStreak(ctx context.Context, path, eventID, teamID string, period int) (int, error) {
	sum, err := c.summary(ctx, path, eventID)
	if err != nil {
		return 0, err
	}
	if len(sum.Plays) == 0 {
		return 0, fmt.Errorf("ESPN summary %s: no play-by-play", eventID)
	}
	best, cur := 0, 0
	for _, p := range sum.Plays {
		if p.Period.Number != period || p.Team.ID != teamID || !isFreeThrow(p.Type.Text) {
			continue
		}
		if p.ScoringPlay {
			cur = 0
			continue
		}
		cur++
		best = max(best, cur)
	}
	return best, nil
}

// isFreeThrow matches the NBA's play types ("Free Throw - 1 of 2") and
// college's ("MadeFreeThrow", used for misses too; scoringPlay tells them apart).
func isFreeThrow(t string) bool {
	return strings.HasPrefix(t, "Free Throw") || t == "MadeFreeThrow"
}

func indexOf(xs []string, x string) int {
	for i, v := range xs {
		if v == x {
			return i
		}
	}
	return -1
}

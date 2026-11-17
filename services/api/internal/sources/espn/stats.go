package espn

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// LongFieldGoal returns teamID's longest made field goal in yards from the box
// score's kicking category. A team that made none legitimately reads 0: its
// kicker's longFieldGoalMade is 0 or "--", or it has no kicking row at all
// because it never kicked (a shutout). Missing structure is an error, never a
// silent 0: no player box score, no entry for the team, or no kicking category
// for a team that scored points (a scoring team kicked at least an extra
// point or had a kicking row ESPN failed to send).
func (c *Client) LongFieldGoal(ctx context.Context, path, eventID, teamID string, points int) (int, error) {
	sum, err := c.summary(ctx, path, eventID)
	if err != nil {
		return 0, err
	}
	if len(sum.Boxscore.Players) == 0 {
		return 0, fmt.Errorf("ESPN summary %s: no player box score", eventID)
	}
	long := 0
	sawTeam, sawKicking := false, false
	for _, t := range sum.Boxscore.Players {
		if t.Team.ID != teamID {
			continue
		}
		sawTeam = true
		for _, cat := range t.Statistics {
			if cat.Name != "kicking" {
				continue
			}
			sawKicking = true
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
	if !sawTeam {
		return 0, fmt.Errorf("ESPN summary %s: no player box score for team %s", eventID, teamID)
	}
	if !sawKicking && points > 0 {
		return 0, fmt.Errorf("ESPN summary %s: no kicking stats for team %s, which scored %d",
			eventID, teamID, points)
	}
	return long, nil
}

// FreeThrowMissStreak returns the most free throws teamID missed in a row
// during period (a quarter in the NBA, a half in college); a make by teamID
// resets the count, the other team's shots don't. Every free throw counts,
// technical, flagrant and clear-path ones included: the sponsors' terms
// ("miss two free throws in a row") make no exception. A summary with no plays
// is an error, never a silent 0.
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

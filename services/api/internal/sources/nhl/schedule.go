package nhl

import (
	"context"
	"sort"
	"time"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

var _ sources.Scheduler = (*Source)(nil)

// GamesOn returns the mapped teams' games on date (the NHL files a game under
// its local date, which for an LA team is the Pacific date), in any state.
// Postponed, canceled and suspended games are left out. It shares the
// per-date cache with GetGameByDate.
func (s *Source) GamesOn(ctx context.Context, date time.Time) ([]sources.ScheduledGame, error) {
	day := date.Format("2006-01-02")
	games, err := s.games(ctx, day)
	if err != nil {
		return nil, err
	}
	byAbbrev := make(map[string]string, len(TeamIDs))
	for code, abbrev := range TeamIDs {
		byAbbrev[abbrev] = code
	}

	var out []sources.ScheduledGame
	for i := range games {
		g := &games[i]
		if g.GameDate != "" && g.GameDate != day {
			continue
		}
		if g.GameScheduleState != "" && g.GameScheduleState != "OK" {
			continue
		}
		state, ok := scheduleState(g.GameState)
		if !ok {
			continue
		}
		starts := startTime(g, date)
		for _, side := range []struct {
			us, them team
			home     bool
		}{{g.HomeTeam, g.AwayTeam, true}, {g.AwayTeam, g.HomeTeam, false}} {
			code, ok := byAbbrev[side.us.Abbrev]
			if !ok {
				continue
			}
			out = append(out, sources.ScheduledGame{
				TeamID:   code,
				Opponent: side.them.Name.Default,
				StartsAt: starts,
				Home:     side.home,
				State:    state,
			})
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].StartsAt.Before(out[b].StartsAt) })
	return out, nil
}

// scheduleState maps the NHL gameState to "pre", "in" or "post".
func scheduleState(s string) (string, bool) {
	switch s {
	case "FUT", "PRE":
		return sources.StatePre, true
	case "LIVE", "CRIT":
		return sources.StateIn, true
	case "FINAL", "OFF":
		return sources.StatePost, true
	}
	return "", false
}

// startTime reads startTimeUTC; when it is missing or unreadable the game is
// placed at midnight Pacific on its date so it still sorts within the day.
func startTime(g *game, date time.Time) time.Time {
	if t, err := time.Parse(time.RFC3339, g.StartTimeUTC); err == nil {
		return t.UTC()
	}
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		loc = time.FixedZone("PT", -8*3600)
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).UTC()
}

package espn

import (
	"context"
	"sort"
	"time"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

var _ sources.Scheduler = (*Source)(nil)

// GamesOn returns the mapped teams' games that start on the given Pacific
// date, from every feed (NCAA reads football and basketball), in any state.
// Postponed and canceled games are left out. Results are sorted by start.
func (s *Source) GamesOn(ctx context.Context, date time.Time) ([]sources.ScheduledGame, error) {
	byESPN := make(map[string]string, len(s.cfg.Teams))
	for code, id := range s.cfg.Teams {
		byESPN[id] = code
	}

	var out []sources.ScheduledGame
	for _, fc := range s.cfg.Feeds {
		events, err := s.client.Games(ctx, fc.Feed, date)
		if err != nil {
			return nil, err
		}
		for i := range events {
			out = append(out, scheduledGames(&events[i], byESPN)...)
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].StartsAt.Before(out[b].StartsAt) })
	return out, nil
}

// scheduledGames returns e from the side of each mapped team playing in it.
func scheduledGames(e *Event, byESPN map[string]string) []sources.ScheduledGame {
	if len(e.Competitions) == 0 {
		return nil
	}
	comp := &e.Competitions[0]
	if len(comp.Competitors) != 2 || skippedStatus(comp.Status.Type.Name) {
		return nil
	}
	state := comp.Status.Type.State
	switch state {
	case sources.StatePre, sources.StateIn, sources.StatePost:
	default:
		return nil
	}
	var out []sources.ScheduledGame
	for i := range comp.Competitors {
		us, them := &comp.Competitors[i], &comp.Competitors[1-i]
		code, ok := byESPN[us.Team.ID]
		if !ok {
			continue
		}
		out = append(out, sources.ScheduledGame{
			TeamID:   code,
			Opponent: shortName(them),
			StartsAt: e.Date.UTC(),
			// Same rule as finalGame: ESPN still labels a side "home" at a
			// neutral site.
			Home:  us.HomeAway == "home" && !comp.NeutralSite,
			State: state,
		})
	}
	return out
}

// skippedStatus reports a game that will not be played as scheduled.
func skippedStatus(name string) bool {
	switch name {
	case "STATUS_POSTPONED", "STATUS_CANCELED", "STATUS_CANCELLED", "STATUS_FORFEIT":
		return true
	}
	return false
}

// shortName is the opponent's nickname ("Hornets", "Arizona"), matching the
// NHL feed's names so Today's matchups read alike; the full name if missing.
func shortName(c *Competitor) string {
	if c.Team.ShortName != "" {
		return c.Team.ShortName
	}
	return c.Team.DisplayName
}

package mlb

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	mlbsdk "github.com/retr0h/mlb-sdk/pkg/mlb"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

var _ sources.Scheduler = (*Source)(nil)

// GamesOn returns the mapped teams' games on date (MLB's official date, the
// local date of the game), in any state, from one league-wide schedule call.
// Postponed, canceled and suspended games, spring training and exhibitions
// are left out.
func (s *Source) GamesOn(ctx context.Context, date time.Time) ([]sources.ScheduledGame, error) {
	games, err := s.schedule(ctx, 0, date)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule: %w", err)
	}
	byID := make(map[mlbsdk.TeamID]string, len(TeamIDs))
	for code, id := range TeamIDs {
		byID[id] = code
	}

	var out []sources.ScheduledGame
	for _, g := range games {
		if skippedStatus(g.Status.DetailedState) || !g.counts() {
			continue
		}
		state, ok := scheduleState(mlbsdk.GameStatus(g.Status.AbstractGameState))
		if !ok {
			continue
		}
		for _, side := range []struct {
			us, them scheduleSide
			home     bool
		}{{g.Teams.Home, g.Teams.Away, true}, {g.Teams.Away, g.Teams.Home, false}} {
			code, ok := byID[mlbsdk.TeamID(side.us.Team.ID)]
			if !ok {
				continue
			}
			out = append(out, sources.ScheduledGame{
				TeamID:   code,
				Opponent: side.them.Team.Name,
				StartsAt: g.GameDate.UTC(),
				Home:     side.home,
				State:    state,
			})
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].StartsAt.Before(out[b].StartsAt) })
	return out, nil
}

// skippedStatus reports a game that was not played (to the end) that day:
// postponed, canceled or suspended ("Suspended: Rain"). MLB marks such a
// game's coarse state Final, so the detailed state decides.
func skippedStatus(detailed string) bool {
	d := strings.ToLower(detailed)
	return strings.HasPrefix(d, "postponed") || strings.HasPrefix(d, "cancelled") ||
		strings.HasPrefix(d, "canceled") || strings.HasPrefix(d, "suspended")
}

func scheduleState(s mlbsdk.GameStatus) (string, bool) {
	switch s {
	case mlbsdk.StatusPreview:
		return sources.StatePre, true
	case mlbsdk.StatusLive:
		return sources.StateIn, true
	case mlbsdk.StatusFinal:
		return sources.StatePost, true
	}
	return "", false
}

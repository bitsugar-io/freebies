package espn_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn/espntest"
	"github.com/retr0h/freebie/services/api/internal/sources/nba"
	"github.com/retr0h/freebie/services/api/internal/sources/ncaa"
)

func utc(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return ts
}

func TestGamesOnLiveGames(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := nba.NewSourceWithClient(srv.URL, srv.Client())

	got, err := src.GamesOn(context.Background(), day(t, "2026-11-15"))
	require.NoError(t, err)
	// Both LA games are in progress; the PHI-IND final is not an LA team.
	assert.Equal(t, []sources.ScheduledGame{
		{TeamID: "LAL", Opponent: "Suns", StartsAt: utc(t, "2026-11-15T21:00:00Z"), Home: false, State: "in"},
		{TeamID: "LAC", Opponent: "Hornets", StartsAt: utc(t, "2026-11-15T21:30:00Z"), Home: true, State: "in"},
	}, got)
}

func TestGamesOnFinalAndEmpty(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := nba.NewSourceWithClient(srv.URL, srv.Client())

	// 7:30pm PT Nov 13 is filed under ESPN's Nov 13 at 03:30Z Nov 14.
	got, err := src.GamesOn(context.Background(), day(t, "2026-11-13"))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "LAC", got[0].TeamID)
	assert.Equal(t, "post", got[0].State)
	assert.True(t, got[0].Home)

	got, err = src.GamesOn(context.Background(), day(t, "2026-12-25"))
	require.NoError(t, err)
	assert.Empty(t, got, "no scoreboard is an empty day")
}

func TestGamesOnSkipsPostponedAndCanceled(t *testing.T) {
	for _, status := range []string{"STATUS_POSTPONED", "STATUS_CANCELED"} {
		t.Run(status, func(t *testing.T) {
			c := byDate(t, nil, map[string]string{
				"20261107": board("1", "2026-11-07T20:00Z", status, false, "LAL", "X", ""),
			})
			got, err := nbaSource(c).GamesOn(context.Background(), day(t, "2026-11-07"))
			require.NoError(t, err)
			assert.Empty(t, got)
		})
	}
	c := byDate(t, nil, map[string]string{
		"20261107": board("1", "2026-11-07T20:00Z", "STATUS_SCHEDULED", false, "LAL", "X", ""),
	})
	got, err := nbaSource(c).GamesOn(context.Background(), day(t, "2026-11-07"))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, sources.ScheduledGame{
		TeamID: "LAL", Opponent: "Away", StartsAt: utc(t, "2026-11-07T20:00:00Z"), Home: true, State: "pre",
	}, got[0])
}

func TestGamesOnNCAAReadsBothFeeds(t *testing.T) {
	srv := espntest.Server(t, nil)
	src := ncaa.NewSourceWithClient(srv.URL, srv.Client())

	// Basketball: UCLA's game is at a neutral site, so it is not a home game.
	got, err := src.GamesOn(context.Background(), day(t, "2026-11-02"))
	require.NoError(t, err)
	assert.Equal(t, []sources.ScheduledGame{
		{TeamID: "UCLA", Opponent: "Arizona", StartsAt: utc(t, "2026-11-03T03:00:00Z"), Home: false, State: "post", Sport: "basketball"},
		{TeamID: "USC", Opponent: "Milwaukee", StartsAt: utc(t, "2026-11-03T03:30:00Z"), Home: true, State: "post", Sport: "basketball"},
	}, got)

	// Football.
	got, err = src.GamesOn(context.Background(), day(t, "2026-11-13"))
	require.NoError(t, err)
	assert.Equal(t, []sources.ScheduledGame{
		{TeamID: "UCLA", Opponent: "Illinois", StartsAt: utc(t, "2026-11-14T02:00:00Z"), Home: true, State: "post", Sport: "football"},
	}, got)
}

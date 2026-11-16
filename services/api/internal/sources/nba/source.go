// Package nba implements sources.Source for the NBA on ESPN's public site API
// (see package espn for the data-terms caveat). The NBA's own feed
// (cdn.nba.com liveData) answers 403 to server traffic, so it is not usable.
//
// Metrics for the deal's team in a finished game:
//
//	win, home_win        1 if the team won (at home), else 0
//	points               points scored, overtime included
//	home_points          points if the team was home, else 0
//	home_q4_opponent_ft_miss_streak
//	                     at home, the most free throws the visitors missed in a
//	                     row in the 4th quarter (overtime excluded), else 0;
//	                     read from the play-by-play
package nba

import (
	"context"
	"net/http"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// TeamIDs maps Freebies' team codes (events.team_id) to ESPN team ids,
// derived from ESPN's teams endpoint for basketball/nba.
var TeamIDs = map[string]string{
	"LAL": "13",
	"LAC": "12",
}

func config() espn.Config {
	return espn.Config{
		League:  "nba",
		Teams:   TeamIDs,
		Metrics: []string{"win", "home_win", "points", "home_points", "home_q4_opponent_ft_miss_streak"},
		Feeds: []espn.FeedConfig{{
			Feed:  espn.Feed{Name: "basketball", Path: "basketball/nba"},
			Build: buildMetrics,
		}},
	}
}

// NewSource creates the nba source on ESPN's public site API, with retry.
func NewSource() *espn.Source { return espn.NewSource(config(), espn.NewClient()) }

// NewSourceWithClient builds the source against baseURL using hc. Tests point
// it at an httptest server.
func NewSourceWithClient(baseURL string, hc *http.Client) *espn.Source {
	return espn.NewSource(config(), espn.NewClientWith(baseURL, hc))
}

func init() { sources.Register(NewSource()) }

// buildMetrics adds the visitors' 4th-quarter free-throw miss streak at home
// (Clippers / Chick-fil-A "Fowl Shot": 2 in a row). Away games skip the
// play-by-play fetch.
func buildMetrics(ctx context.Context, c *espn.Client, g *espn.Game) (map[string]int, error) {
	m, err := espn.PointsMetrics(ctx, c, g)
	if err != nil {
		return nil, err
	}
	m["home_q4_opponent_ft_miss_streak"] = 0
	if g.Home {
		n, err := c.FreeThrowMissStreak(ctx, g.Feed.Path, g.Event.ID, g.Them.Team.ID, 4)
		if err != nil {
			return nil, err
		}
		m["home_q4_opponent_ft_miss_streak"] = n
	}
	return m, nil
}

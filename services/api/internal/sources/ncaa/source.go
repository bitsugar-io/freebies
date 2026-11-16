// Package ncaa implements sources.Source for USC and UCLA on ESPN's public
// site API (see package espn for the data-terms caveat). The NCAA's own data
// endpoints are not an official public API; ncaa-api.henrygd.me is a
// third-party proxy and is not used.
//
// The league covers two sports, so a team can play football in the fall and
// men's basketball in the winter (the teams table lists them as "football").
// Each game produces the generic metrics and the same metrics prefixed with its
// sport, so a deal can pick one sport and never fire on the other:
//
//	win, home_win, points, home_points
//	football_win, football_home_win, football_points, football_home_points
//	basketball_win, basketball_home_win, basketball_points, basketball_home_points
//
// Neutral-site games (bowls, tournaments) are never home games. Women's
// basketball and other sports are not read.
package ncaa

import (
	"context"
	"net/http"
	"net/url"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// TeamIDs maps Freebies' team codes (events.team_id) to ESPN team ids,
// derived from ESPN's teams endpoint for football/college-football.
var TeamIDs = map[string]string{
	"USC":  "30",
	"UCLA": "26",
}

func config() espn.Config {
	return espn.Config{
		League: "ncaa",
		Teams:  TeamIDs,
		Metrics: []string{"win", "home_win", "points", "home_points",
			"football_win", "football_home_win", "football_points", "football_home_points",
			"basketball_win", "basketball_home_win", "basketball_points", "basketball_home_points",
			h2FTStreak},
		Feeds: []espn.FeedConfig{
			{
				// groups=80 is FBS; limit keeps a full Saturday on one page.
				Feed: espn.Feed{
					Name: "football", Path: "football/college-football",
					Query: url.Values{"groups": {"80"}, "limit": {"400"}},
				},
				Build: sportMetrics("football"),
			},
			{
				// The default basketball scoreboard lists only featured games;
				// groups=50 is Division I.
				Feed: espn.Feed{
					Name: "basketball", Path: "basketball/mens-college-basketball",
					Query: url.Values{"groups": {"50"}, "limit": {"400"}},
				},
				Build: sportMetrics("basketball"),
			},
		},
	}
}

// NewSource creates the ncaa source on ESPN's public site API, with retry.
// h2FTStreak is, at a home basketball game, the most free throws the visitors
// missed in a row in the 2nd half (play-by-play), else 0.
const h2FTStreak = "basketball_home_h2_opponent_ft_miss_streak"

func NewSource() *espn.Source { return espn.NewSource(config(), espn.NewClient()) }

// NewSourceWithClient builds the source against baseURL using hc. Tests point
// it at an httptest server.
func NewSourceWithClient(baseURL string, hc *http.Client) *espn.Source {
	return espn.NewSource(config(), espn.NewClientWith(baseURL, hc))
}

func init() { sources.Register(NewSource()) }

// sportMetrics returns the generic points metrics plus sport-prefixed copies.
func sportMetrics(sport string) func(context.Context, *espn.Client, *espn.Game) (map[string]int, error) {
	return func(ctx context.Context, c *espn.Client, g *espn.Game) (map[string]int, error) {
		m, err := espn.PointsMetrics(ctx, c, g)
		if err != nil {
			return nil, err
		}
		out := make(map[string]int, 2*len(m)+1)
		for k, v := range m {
			out[k] = v
			out[sport+"_"+k] = v
		}
		if sport == "basketball" {
			// UCLA / Fat Sal's "Missed Fry Throws": visitors miss 2 free
			// throws in a row in the 2nd half of a home game.
			out[h2FTStreak] = 0
			if g.Home {
				n, err := c.FreeThrowMissStreak(ctx, g.Feed.Path, g.Event.ID, g.Them.Team.ID, 2)
				if err != nil {
					return nil, err
				}
				out[h2FTStreak] = n
			}
		}
		return out, nil
	}
}

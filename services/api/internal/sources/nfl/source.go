// Package nfl implements sources.Source for the NFL on ESPN's public site API
// (see package espn for the data-terms caveat). The NFL's own feeds are not
// publicly reachable.
//
// Metrics for the deal's team in a finished game:
//
//	win, home_win        1 if the team won (at home), else 0; a tie is neither
//	points               points scored, overtime included
//	home_points          points if the team was home, else 0
//	interceptions        passes the team's defense intercepted (any venue),
//	                     read from the box score; pick-sixes count
//	long_field_goal      the team's longest made field goal in yards, 0 if none
//	                     (box score kicking; any kicker on the team)
//
// A neutral-site game (international games, the Super Bowl) is never a home
// game for the home_* metrics, whichever side ESPN labels "home".
package nfl

import (
	"context"
	"net/http"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// TeamIDs maps Freebies' team codes (events.team_id) to ESPN team ids,
// derived from ESPN's teams endpoint for football/nfl.
var TeamIDs = map[string]string{
	"LAR":  "14",
	"LACH": "24",
}

func config() espn.Config {
	return espn.Config{
		League:  "nfl",
		Teams:   TeamIDs,
		Metrics: []string{"win", "home_win", "points", "home_points", "interceptions", "long_field_goal"},
		Feeds: []espn.FeedConfig{{
			Feed:  espn.Feed{Name: "football", Path: "football/nfl"},
			Build: buildMetrics,
		}},
	}
}

// NewSource creates the nfl source on ESPN's public site API, with retry.
func NewSource() *espn.Source { return espn.NewSource(config(), espn.NewClient()) }

// NewSourceWithClient builds the source against baseURL using hc. Tests point
// it at an httptest server.
func NewSourceWithClient(baseURL string, hc *http.Client) *espn.Source {
	return espn.NewSource(config(), espn.NewClientWith(baseURL, hc))
}

func init() { sources.Register(NewSource()) }

// buildMetrics adds interceptions, which the scoreboard lacks. ESPN's box
// score lists "interceptions" per team as passes that team threw, so the
// team's defensive interceptions are the opponent's stat.
func buildMetrics(ctx context.Context, c *espn.Client, g *espn.Game) (map[string]int, error) {
	m, err := espn.PointsMetrics(ctx, c, g)
	if err != nil {
		return nil, err
	}
	n, err := c.TeamStat(ctx, g.Feed.Path, g.Event.ID, g.Them.Team.ID, "interceptions")
	if err != nil {
		return nil, err
	}
	m["interceptions"] = n
	// Longest field goal our kicker made, in yards (Chargers / Firehouse Subs: 40+).
	long, err := c.LongFieldGoal(ctx, g.Feed.Path, g.Event.ID, g.Us.Team.ID)
	if err != nil {
		return nil, err
	}
	m["long_field_goal"] = long
	return m, nil
}

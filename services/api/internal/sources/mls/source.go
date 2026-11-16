// Package mls implements sources.Source for Major League Soccer on ESPN's
// public site API (see package espn for the data-terms caveat). MLS's own
// stats API is not publicly reachable.
//
// Metrics for the deal's team in a finished match (see espn.SoccerMetrics):
//
//	win, home_win, draw
//	goals, home_goals_scored
//	first_half_goals, home_first_half_goals    goals up to 45' plus stoppage
//	scored_first_in_first_half                 first goal of the match, in the first half
//	home_scored_first_in_first_half            same, home games only
package mls

import (
	"net/http"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// TeamIDs maps Freebies' team codes (events.team_id) to ESPN team ids,
// derived from ESPN's teams endpoint for soccer/usa.1.
var TeamIDs = map[string]string{
	"LAFC": "18966",
	"LAG":  "187",
}

func config() espn.Config {
	return espn.Config{
		League: "mls",
		Teams:  TeamIDs,
		Metrics: []string{"win", "home_win", "draw", "goals", "home_goals_scored",
			"first_half_goals", "home_first_half_goals",
			"scored_first_in_first_half", "home_scored_first_in_first_half"},
		Feeds: []espn.FeedConfig{{
			Feed:  espn.Feed{Name: "soccer", Path: "soccer/usa.1"},
			Build: espn.SoccerMetrics,
		}},
	}
}

// NewSource creates the mls source on ESPN's public site API, with retry.
func NewSource() *espn.Source { return espn.NewSource(config(), espn.NewClient()) }

// NewSourceWithClient builds the source against baseURL using hc. Tests point
// it at an httptest server.
func NewSourceWithClient(baseURL string, hc *http.Client) *espn.Source {
	return espn.NewSource(config(), espn.NewClientWith(baseURL, hc))
}

func init() { sources.Register(NewSource()) }

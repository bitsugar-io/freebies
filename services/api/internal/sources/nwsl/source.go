// Package nwsl implements sources.Source for the NWSL on ESPN's public site
// API (see package espn for the data-terms caveat). The NWSL's own API is not
// publicly reachable.
//
// Metrics are the soccer set, see espn.SoccerMetrics and package mls.
package nwsl

import (
	"net/http"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// TeamIDs maps Freebies' team codes (events.team_id) to ESPN team ids,
// derived from ESPN's teams endpoint for soccer/usa.nwsl.
var TeamIDs = map[string]string{
	"ACFC": "21422",
}

func config() espn.Config {
	return espn.Config{
		League: "nwsl",
		Teams:  TeamIDs,
		Metrics: []string{"win", "home_win", "draw", "goals", "home_goals_scored",
			"first_half_goals", "home_first_half_goals",
			"scored_first_in_first_half", "home_scored_first_in_first_half"},
		Feeds: []espn.FeedConfig{{
			Feed:  espn.Feed{Name: "soccer", Path: "soccer/usa.nwsl"},
			Build: espn.SoccerMetrics,
		}},
	}
}

// NewSource creates the nwsl source on ESPN's public site API, with retry.
func NewSource() *espn.Source { return espn.NewSource(config(), espn.NewClient()) }

// NewSourceWithClient builds the source against baseURL using hc. Tests point
// it at an httptest server.
func NewSourceWithClient(baseURL string, hc *http.Client) *espn.Source {
	return espn.NewSource(config(), espn.NewClientWith(baseURL, hc))
}

func init() { sources.Register(NewSource()) }

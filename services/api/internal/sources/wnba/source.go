// Package wnba implements sources.Source for the WNBA on ESPN's public site
// API (see package espn for the data-terms caveat). The WNBA's own feed
// (cdn.wnba.com liveData) answers 403 to server traffic, so it is not usable.
//
// Metrics are the basketball set, see package nba.
package wnba

import (
	"net/http"

	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// TeamIDs maps Freebies' team codes (events.team_id) to ESPN team ids,
// derived from ESPN's teams endpoint for basketball/wnba.
var TeamIDs = map[string]string{
	"LAS": "6",
}

func config() espn.Config {
	return espn.Config{
		League:  "wnba",
		Teams:   TeamIDs,
		Metrics: []string{"win", "home_win", "points", "home_points"},
		Feeds: []espn.FeedConfig{{
			Feed:  espn.Feed{Name: "basketball", Path: "basketball/wnba"},
			Build: espn.PointsMetrics,
		}},
	}
}

// NewSource creates the wnba source on ESPN's public site API, with retry.
func NewSource() *espn.Source { return espn.NewSource(config(), espn.NewClient()) }

// NewSourceWithClient builds the source against baseURL using hc. Tests point
// it at an httptest server.
func NewSourceWithClient(baseURL string, hc *http.Client) *espn.Source {
	return espn.NewSource(config(), espn.NewClientWith(baseURL, hc))
}

func init() { sources.Register(NewSource()) }

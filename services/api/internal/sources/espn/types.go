package espn

import (
	"fmt"
	"time"
)

// scoreboardResponse is the subset of GET /{sport}/{league}/scoreboard we read.
type scoreboardResponse struct {
	Events []Event `json:"events"`
}

// Event is one game on the scoreboard.
type Event struct {
	ID   string   `json:"id"`
	Date espnTime `json:"date"`
	Name string   `json:"name"`
	// Season is the part of the season the game belongs to; see seasonPhase
	// for how Type and Slug are read.
	Season struct {
		Type int    `json:"type"`
		Slug string `json:"slug"`
	} `json:"season"`
	Competitions []Competition `json:"competitions"`
}

// Competition is the game itself (ESPN lists exactly one per event for the
// leagues we use).
type Competition struct {
	NeutralSite bool `json:"neutralSite"`
	// Type.Abbreviation is "STD" for a normal game, "ALLSTAR" for an All-Star
	// game and "EXH" for an exhibition.
	Type struct {
		Abbreviation string `json:"abbreviation"`
	} `json:"type"`
	Status      Status       `json:"status"`
	Competitors []Competitor `json:"competitors"`
	Details     []Detail     `json:"details"`
}

// Status is the game state.
type Status struct {
	Type struct {
		Name      string `json:"name"`
		State     string `json:"state"`
		Completed bool   `json:"completed"`
	} `json:"type"`
}

// Competitor is one side of a game.
type Competitor struct {
	HomeAway string `json:"homeAway"`
	Score    string `json:"score"`
	Winner   bool   `json:"winner"`
	Team     struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		ShortName   string `json:"shortDisplayName"`
		Abbrev      string `json:"abbreviation"`
	} `json:"team"`
}

// Detail is a match event (soccer lists goals and cards here).
type Detail struct {
	Clock struct {
		Value   float64 `json:"value"`
		Display string  `json:"displayValue"`
	} `json:"clock"`
	Team struct {
		ID string `json:"id"`
	} `json:"team"`
	ScoringPlay bool `json:"scoringPlay"`
	OwnGoal     bool `json:"ownGoal"`
	Shootout    bool `json:"shootout"`
}

// espnTime parses ESPN's minute-precision UTC timestamps ("2026-11-15T00:00Z").
type espnTime struct{ time.Time }

func (t *espnTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		return nil
	}
	if len(s) < 2 || s[0] != '"' {
		return fmt.Errorf("espn: bad time %s", s)
	}
	s = s[1 : len(s)-1]
	for _, layout := range []string{"2006-01-02T15:04Z07:00", "2006-01-02T15:04:05Z07:00"} {
		if p, err := time.Parse(layout, s); err == nil {
			t.Time = p
			return nil
		}
	}
	return fmt.Errorf("espn: bad time %q", s)
}

// summaryResponse is the subset of GET /{sport}/{league}/summary we read.
type summaryResponse struct {
	Boxscore struct {
		Teams []struct {
			Team struct {
				ID string `json:"id"`
			} `json:"team"`
			Statistics []struct {
				Name         string `json:"name"`
				DisplayValue string `json:"displayValue"`
			} `json:"statistics"`
		} `json:"teams"`
		// Players holds per-player categories ("kicking", ...) per team.
		Players []struct {
			Team struct {
				ID string `json:"id"`
			} `json:"team"`
			Statistics []struct {
				Name     string   `json:"name"`
				Keys     []string `json:"keys"`
				Athletes []struct {
					Stats []string `json:"stats"`
				} `json:"athletes"`
			} `json:"statistics"`
		} `json:"players"`
	} `json:"boxscore"`
	// Plays is the play-by-play (basketball; football uses drives).
	Plays []Play `json:"plays"`
}

// Play is one play-by-play entry.
type Play struct {
	Period struct {
		Number int `json:"number"`
	} `json:"period"`
	Type struct {
		Text string `json:"text"`
	} `json:"type"`
	ScoringPlay bool `json:"scoringPlay"`
	Team        struct {
		ID string `json:"id"`
	} `json:"team"`
}

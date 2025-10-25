package mlb

// Team ID mappings (MLB API team IDs)
var TeamIDs = map[string]int{
	"LAD": 119, // Los Angeles Dodgers
	"SD":  135, // San Diego Padres
	"SF":  137, // San Francisco Giants
	"ARI": 109, // Arizona Diamondbacks
	"COL": 115, // Colorado Rockies
	"LAA": 108, // Los Angeles Angels
	"OAK": 133, // Oakland Athletics
	"SEA": 136, // Seattle Mariners
	"TEX": 140, // Texas Rangers
	"HOU": 117, // Houston Astros
	"NYY": 147, // New York Yankees
	"NYM": 121, // New York Mets
	"BOS": 111, // Boston Red Sox
	"CHC": 112, // Chicago Cubs
	"CHW": 145, // Chicago White Sox
	"ATL": 144, // Atlanta Braves
	"MIA": 146, // Miami Marlins
	"PHI": 143, // Philadelphia Phillies
	"WSH": 120, // Washington Nationals
	"BAL": 110, // Baltimore Orioles
	"TB":  139, // Tampa Bay Rays
	"TOR": 141, // Toronto Blue Jays
	"CLE": 114, // Cleveland Guardians
	"DET": 116, // Detroit Tigers
	"KC":  118, // Kansas City Royals
	"MIN": 142, // Minnesota Twins
	"CIN": 113, // Cincinnati Reds
	"MIL": 158, // Milwaukee Brewers
	"PIT": 134, // Pittsburgh Pirates
	"STL": 138, // St. Louis Cardinals
}

// ScheduleResponse represents the MLB schedule API response
type ScheduleResponse struct {
	Dates []struct {
		Games []Game `json:"games"`
	} `json:"dates"`
}

// Game represents a game from the schedule
type Game struct {
	GamePk   int    `json:"gamePk"`
	GameDate string `json:"gameDate"`
	Status   struct {
		AbstractGameState string `json:"abstractGameState"` // "Final", "Live", "Preview"
	} `json:"status"`
	Teams struct {
		Away struct {
			Team struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"team"`
			Score int `json:"score"`
		} `json:"away"`
		Home struct {
			Team struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"team"`
			Score int `json:"score"`
		} `json:"home"`
	} `json:"teams"`
}

// BoxScoreResponse represents the boxscore API response
type BoxScoreResponse struct {
	Teams struct {
		Away TeamBoxScore `json:"away"`
		Home TeamBoxScore `json:"home"`
	} `json:"teams"`
}

// TeamBoxScore represents a team's boxscore
type TeamBoxScore struct {
	Team struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	TeamStats struct {
		Pitching struct {
			StrikeOuts int `json:"strikeOuts"`
			Hits       int `json:"hits"`
			Runs       int `json:"runs"`
			HomeRuns   int `json:"homeRuns"`
			Walks      int `json:"baseOnBalls"`
		} `json:"pitching"`
		Batting struct {
			Runs     int `json:"runs"`
			Hits     int `json:"hits"`
			HomeRuns int `json:"homeRuns"`
			RBI      int `json:"rbi"`
		} `json:"batting"`
	} `json:"teamStats"`
}

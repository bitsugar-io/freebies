package nhl

// TeamIDs maps Freebies' team codes (as stored in events.team_id) to the NHL
// web API's three-letter team abbreviations. The mapping is explicit so a
// team that is not listed here fails loudly instead of silently never
// matching.
var TeamIDs = map[string]string{
	"ANA": "ANA",
	"BOS": "BOS",
	"BUF": "BUF",
	"CAR": "CAR",
	"CBJ": "CBJ",
	"CGY": "CGY",
	"CHI": "CHI",
	"COL": "COL",
	"DAL": "DAL",
	"DET": "DET",
	"EDM": "EDM",
	"FLA": "FLA",
	"LAK": "LAK",
	"MIN": "MIN",
	"MTL": "MTL",
	"NJD": "NJD",
	"NSH": "NSH",
	"NYI": "NYI",
	"NYR": "NYR",
	"OTT": "OTT",
	"PHI": "PHI",
	"PIT": "PIT",
	"SEA": "SEA",
	"SJS": "SJS",
	"STL": "STL",
	"TBL": "TBL",
	"TOR": "TOR",
	"UTA": "UTA",
	"VAN": "VAN",
	"VGK": "VGK",
	"WPG": "WPG",
	"WSH": "WSH",
}

// scoreResponse is the subset of GET /v1/score/{date} we read.
type scoreResponse struct {
	// Games is nil when the key is missing or null, which is not a score page.
	Games *[]game `json:"games"`
}

type game struct {
	ID           int64  `json:"id"`
	GameDate     string `json:"gameDate"`
	StartTimeUTC string `json:"startTimeUTC"`
	GameState    string `json:"gameState"`
	// GameType: 1 preseason, 2 regular season, 3 playoffs, 4 All-Star,
	// 19 international events (see counts).
	GameType          int    `json:"gameType"`
	GameScheduleState string `json:"gameScheduleState"`
	AwayTeam          team   `json:"awayTeam"`
	HomeTeam          team   `json:"homeTeam"`
	GameOutcome       struct {
		LastPeriodType string `json:"lastPeriodType"`
	} `json:"gameOutcome"`
	PeriodDescriptor struct {
		PeriodType string `json:"periodType"`
	} `json:"periodDescriptor"`
	// Goals lists every goal with its period and clock (shootout goals too).
	Goals []goal `json:"goals"`
}

type team struct {
	Abbrev string `json:"abbrev"`
	Score  *int   `json:"score"`
	Name   struct {
		Default string `json:"default"`
	} `json:"name"`
}

type goal struct {
	Period           int `json:"period"`
	PeriodDescriptor struct {
		PeriodType string `json:"periodType"`
	} `json:"periodDescriptor"`
	// TimeInPeriod is elapsed time, "MM:SS".
	TimeInPeriod string `json:"timeInPeriod"`
	TeamAbbrev   string `json:"teamAbbrev"`
}

package mlb

import "github.com/retr0h/mlb-sdk/pkg/mlb"

// TeamIDs maps Freebies' wire-format team codes (the same strings carried in
// the database and on Source method calls) to the MLB Stats API's numeric
// team IDs as typed by mlb-sdk. Source code-callers pass strings; we resolve
// once here.
var TeamIDs = map[string]mlb.TeamID{
	"LAD": mlb.LAD,
	"SD":  mlb.SD,
	"SF":  mlb.SF,
	"ARI": mlb.ARI,
	"COL": mlb.COL,
	"LAA": mlb.LAA,
	"OAK": mlb.OAK,
	"SEA": mlb.SEA,
	"TEX": mlb.TEX,
	"HOU": mlb.HOU,
	"NYY": mlb.NYY,
	"NYM": mlb.NYM,
	"BOS": mlb.BOS,
	"CHC": mlb.CHC,
	"CHW": mlb.CHW,
	"ATL": mlb.ATL,
	"MIA": mlb.MIA,
	"PHI": mlb.PHI,
	"WSH": mlb.WSH,
	"BAL": mlb.BAL,
	"TB":  mlb.TB,
	"TOR": mlb.TOR,
	"CLE": mlb.CLE,
	"DET": mlb.DET,
	"KC":  mlb.KC,
	"MIN": mlb.MIN,
	"CIN": mlb.CIN,
	"MIL": mlb.MIL,
	"PIT": mlb.PIT,
	"STL": mlb.STL,
}

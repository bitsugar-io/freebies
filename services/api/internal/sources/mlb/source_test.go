package mlb

import (
	"encoding/json"
	"testing"
)

// Verifies that the BoxScoreResponse struct correctly unmarshals the Info
// block shape returned by the live MLB Stats API, so doublePlaysTurned can
// find DP entries on real responses.
func TestBoxScoreResponseUnmarshal_Info(t *testing.T) {
	raw := []byte(`{
		"teams": {
			"home": {
				"team": {"id": 119, "name": "Los Angeles Dodgers"},
				"teamStats": {"pitching": {}, "batting": {}},
				"info": [
					{"title": "BATTING", "fieldList": [{"label": "TB", "value": "Smith 2."}]},
					{"title": "FIELDING", "fieldList": [
						{"label": "E", "value": "Jarvis (1, throw)."},
						{"label": "DP", "value": "2 (Freeland; Betts-Freeman)."}
					]}
				]
			},
			"away": {"team": {"id": 144, "name": "Atlanta Braves"}, "teamStats": {"pitching": {}, "batting": {}}}
		}
	}`)

	var resp BoxScoreResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := doublePlaysTurned(resp.Teams.Home); got != 2 {
		t.Errorf("home doublePlaysTurned = %d, want 2", got)
	}
	if got := doublePlaysTurned(resp.Teams.Away); got != 0 {
		t.Errorf("away doublePlaysTurned = %d, want 0", got)
	}
}

func TestParseDPCount(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"whitespace only", "   ", 0},
		{"single DP, no leading number", "(Freeman, F-Rojas, M).", 1},
		{"two DPs with leading 2", "2 (Freeland, A-Betts-Freeman, F; Betts-Freeman, F).", 2},
		{"three DPs with leading 3", "3 (2 Rocchio-Arias, G-Hoskins; Hoskins-Arias, G-Hoskins).", 3},
		{"leading whitespace then number", "  2 (X-Y).", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := parseDPCount(c.input); got != c.want {
				t.Errorf("parseDPCount(%q) = %d, want %d", c.input, got, c.want)
			}
		})
	}
}

func TestDoublePlaysTurned(t *testing.T) {
	makeTeam := func(sections ...BoxScoreInfoSection) TeamBoxScore {
		return TeamBoxScore{Info: sections}
	}

	cases := []struct {
		name string
		team TeamBoxScore
		want int
	}{
		{
			name: "no info block",
			team: TeamBoxScore{},
			want: 0,
		},
		{
			name: "fielding section without DP entry",
			team: makeTeam(BoxScoreInfoSection{
				Title:     "FIELDING",
				FieldList: []BoxScoreInfoItem{{Label: "E", Value: "Jarvis (1, throw)."}},
			}),
			want: 0,
		},
		{
			name: "DP in non-fielding section is ignored",
			team: makeTeam(BoxScoreInfoSection{
				Title:     "BATTING",
				FieldList: []BoxScoreInfoItem{{Label: "DP", Value: "2 (X-Y)."}},
			}),
			want: 0,
		},
		{
			name: "single DP entry",
			team: makeTeam(BoxScoreInfoSection{
				Title:     "FIELDING",
				FieldList: []BoxScoreInfoItem{{Label: "DP", Value: "(Freeman, F-Rojas, M)."}},
			}),
			want: 1,
		},
		{
			name: "two DPs",
			team: makeTeam(BoxScoreInfoSection{
				Title:     "FIELDING",
				FieldList: []BoxScoreInfoItem{{Label: "DP", Value: "2 (Freeland; Betts-Freeman)."}},
			}),
			want: 2,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := doublePlaysTurned(c.team); got != c.want {
				t.Errorf("doublePlaysTurned() = %d, want %d", got, c.want)
			}
		})
	}
}

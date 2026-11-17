package mlb

import "testing"

// Playoff rounds are postseason; the regular season, a missing type and
// spring training are not (spring training never counts at all).
func TestPostseasonGameTypes(t *testing.T) {
	for gt, want := range map[string]bool{"R": false, "": false, "S": false, "F": true, "D": true, "L": true, "W": true} {
		if got := (&scheduleGame{GameType: gt}).postseason(); got != want {
			t.Errorf("gameType %q: postseason = %v, want %v", gt, got, want)
		}
	}
	if !(&Source{}).KnowsPostseason() {
		t.Error("the MLB source must report that it knows playoff games")
	}
}

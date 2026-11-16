package triggers

import (
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

// "Valid through the following business day" (UCLA / Fat Sal's): usable as soon
// as it triggers, ending with the first weekday after the game. Unlike
// next_business_day it is never locked over a weekend.
func TestWindowThroughNextBusinessDay(t *testing.T) {
	loc := pt(t)
	rule := Rule{RedemptionWindow: ThroughNextBusinessDay}
	tests := []struct {
		name       string
		game       time.Time
		wantExpire time.Time
	}{
		{"Saturday game: usable Sunday, through Monday", time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"Friday game: through Monday", time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"Wednesday game: through Thursday", time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 11, 12, 23, 59, 59, 0, loc)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a, e := windowFor(tc.game, rule, loc)
			if !a.IsZero() {
				t.Errorf("availableAt = %s, want usable immediately", a.In(loc))
			}
			if !e.Equal(tc.wantExpire) {
				t.Errorf("expiresAt = %s, want %s", e.In(loc), tc.wantExpire)
			}
		})
	}
}

// "Valid 10am-11:59pm" (Angels / Del Taco), "not during breakfast" (Dodgers /
// McDonald's): available_from_local moves the start to that clock time on the
// first usable day; with a delayed unlock it applies on the unlock day.
func TestWindowAvailableFromLocal(t *testing.T) {
	loc := pt(t)
	game := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC) // Saturday

	a, e := windowFor(game, Rule{RedemptionWindow: SameDay, AvailableFromLocal: "10:00"}, loc)
	if want := time.Date(2026, 11, 15, 10, 0, 0, 0, loc); !a.Equal(want) {
		t.Errorf("same_day from 10:00: availableAt = %s, want %s", a.In(loc), want)
	}
	if want := time.Date(2026, 11, 15, 23, 59, 59, 0, loc); !e.Equal(want) {
		t.Errorf("same_day from 10:00: expiresAt = %s, want %s", e.In(loc), want)
	}

	a, _ = windowFor(game, Rule{UnlockAfterDays: 2, ValidDays: 7, AvailableFromLocal: "10:30"}, loc)
	if want := time.Date(2026, 11, 16, 10, 30, 0, 0, loc); !a.Equal(want) {
		t.Errorf("unlock day from 10:30: availableAt = %s, want %s", a.In(loc), want)
	}

	// At-the-game deals are claimed that night; a start time never applies.
	a, _ = windowFor(game, Rule{RedemptionWindow: GameNight, AvailableFromLocal: "10:00"}, loc)
	if !a.IsZero() {
		t.Errorf("game_night ignores available_from_local, got %s", a.In(loc))
	}
}

// "During the 2026-2027 regular season" (Chargers / Firehouse): a playoff game
// never fires the deal, however good the stat.
func TestRegularSeasonOnly(t *testing.T) {
	rule := Rule{Metric: "long_field_goal", Operator: ">=", Value: 40, Season: "regular"}
	stats := &sources.GameStats{Metrics: map[string]int{"long_field_goal": 52}}
	if !evaluateRule(rule, stats) {
		t.Fatal("the stat itself meets the rule")
	}
	c := NewChecker(nil)
	stats.Postseason = true
	if got := c.seasonAllows(rule, stats); got {
		t.Error("regular-season-only deal allowed a playoff game")
	}
	stats.Postseason = false
	if got := c.seasonAllows(rule, stats); !got {
		t.Error("regular-season-only deal blocked a regular-season game")
	}
	if !c.seasonAllows(Rule{}, &sources.GameStats{Postseason: true}) {
		t.Error("a deal with no season limit must allow playoff games")
	}
}

// Business-day windows skip US federal holidays: a Wednesday Nov 25 2026 game
// is not valid on Thanksgiving (Thu Nov 26); it runs through Friday Nov 27.
func TestBusinessDayWindowsSkipHolidays(t *testing.T) {
	loc := pt(t)
	wed := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	want := time.Date(2026, 11, 27, 23, 59, 59, 0, loc)
	if _, e := windowFor(wed, Rule{RedemptionWindow: ThroughNextBusinessDay}, loc); !e.Equal(want) {
		t.Errorf("through_next_business_day: expires %s, want Fri %s", e.In(loc), want)
	}
	a, e := windowFor(wed, Rule{RedemptionWindow: NextBusinessDay}, loc)
	if wantA := time.Date(2026, 11, 27, 0, 0, 0, 0, loc); !a.Equal(wantA) || !e.Equal(want) {
		t.Errorf("next_business_day: %s to %s, want Fri Nov 27 only", a.In(loc), e.In(loc))
	}
}

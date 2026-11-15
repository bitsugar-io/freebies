package triggers

import (
	"testing"
	"time"
)

func pt(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

// TestWindowFor: every window is computed from the game date (Sun Nov 15
// here) and ends at 23:59:59 Pacific on its last valid day.
func TestWindowFor(t *testing.T) {
	loc := pt(t)
	game := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		rule          Rule
		wantAvailable time.Time // zero = immediately
		wantExpires   time.Time
	}{
		{"same day is the day after the game", Rule{RedemptionWindow: "same_day"},
			time.Time{}, time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"same day breakfast cutoff", Rule{RedemptionWindow: "same_day", ExpiresAtLocal: "10:30"},
			time.Time{}, time.Date(2026, 11, 16, 10, 30, 0, 0, loc)},
		{"bad clock falls back to end of day", Rule{RedemptionWindow: "same_day", ExpiresAtLocal: "nope"},
			time.Time{}, time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"next day", Rule{RedemptionWindow: "next_day"},
			time.Time{}, time.Date(2026, 11, 17, 23, 59, 59, 0, loc)},
		{"unknown defaults to next day", Rule{RedemptionWindow: "bogus"},
			time.Time{}, time.Date(2026, 11, 17, 23, 59, 59, 0, loc)},
		{"week lasts seven days from the day after the game", Rule{RedemptionWindow: "week"},
			time.Time{}, time.Date(2026, 11, 22, 23, 59, 59, 0, loc)},
		{"unlock 2 days after the game, good for a week ends on its last day",
			Rule{RedemptionWindow: "week", UnlockAfterDays: 2, ValidDays: 7},
			time.Date(2026, 11, 17, 0, 0, 0, 0, loc), time.Date(2026, 11, 23, 23, 59, 59, 0, loc)},
		{"unlock without valid_days lasts the unlock day", Rule{UnlockAfterDays: 1},
			time.Date(2026, 11, 16, 0, 0, 0, 0, loc), time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"expires_at_local is ignored when the deal unlocks later",
			Rule{UnlockAfterDays: 1, ExpiresAtLocal: "10:30"},
			time.Date(2026, 11, 16, 0, 0, 0, 0, loc), time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"valid_days counts calendar days from the day after the game", Rule{ValidDays: 3},
			time.Time{}, time.Date(2026, 11, 18, 23, 59, 59, 0, loc)},
		{"valid_days still honors expires_at_local", Rule{ValidDays: 2, ExpiresAtLocal: "10:30"},
			time.Time{}, time.Date(2026, 11, 17, 10, 30, 0, 0, loc)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotA, gotE := windowFor(game, tc.rule, loc)
			if tc.wantAvailable.IsZero() != gotA.IsZero() ||
				(!gotA.IsZero() && !gotA.Equal(tc.wantAvailable)) {
				t.Errorf("availableAt: got %s want %s", gotA, tc.wantAvailable)
			}
			if !gotE.Equal(tc.wantExpires) {
				t.Errorf("expiresAt: got %s want %s", gotE, tc.wantExpires)
			}
		})
	}
}

// TestWindowForEndsAtLastMoment: no window may end at midnight starting a day;
// the deadline always falls on its last valid day.
func TestWindowForEndsAtLastMoment(t *testing.T) {
	loc := pt(t)
	game := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	for _, r := range []Rule{
		{RedemptionWindow: "same_day"}, {RedemptionWindow: "next_day"}, {RedemptionWindow: "week"},
		{RedemptionWindow: "next_business_day"}, {ValidDays: 4},
		{UnlockAfterDays: 2, ValidDays: 7}, {UnlockAfterDays: 3},
	} {
		_, e := windowFor(game, r, loc)
		l := e.In(loc)
		if l.Hour() != 23 || l.Minute() != 59 || l.Second() != 59 {
			t.Errorf("%+v ends %s, want 23:59:59", r, l)
		}
	}
}

// TestWindowForIgnoresCheckTime: the same game gives the same window however
// late the run happens; only the game date matters.
func TestWindowForGameDateOnly(t *testing.T) {
	loc := pt(t)
	a := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 11, 15, 19, 30, 0, 0, loc) // same Pacific date, any clock time
	for _, r := range []Rule{{RedemptionWindow: "same_day"}, {ValidDays: 3}, {RedemptionWindow: "next_business_day"}} {
		_, ea := windowFor(a, r, loc)
		_, eb := windowFor(b, r, loc)
		if !ea.Equal(eb) {
			t.Errorf("%+v: %s != %s", r, ea, eb)
		}
	}
}

func TestWindowForNextBusinessDay(t *testing.T) {
	loc := pt(t)
	rule := Rule{RedemptionWindow: "next_business_day"}
	tests := []struct {
		name       string
		game       time.Time
		wantAvail  time.Time
		wantExpire time.Time
	}{
		{"Friday game: valid Monday", time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 11, 16, 0, 0, 0, 0, loc), time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"Saturday game: valid Monday", time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 11, 16, 0, 0, 0, 0, loc), time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"Sunday game: valid Monday, open the morning after", time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC),
			time.Time{}, time.Date(2026, 11, 16, 23, 59, 59, 0, loc)},
		{"Wednesday game: valid Thursday, open the morning after", time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC),
			time.Time{}, time.Date(2026, 11, 12, 23, 59, 59, 0, loc)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a, e := windowFor(tc.game, rule, loc)
			if tc.wantAvail.IsZero() != a.IsZero() || (!a.IsZero() && !a.Equal(tc.wantAvail)) {
				t.Errorf("availableAt got %s want %s", a, tc.wantAvail)
			}
			if !e.Equal(tc.wantExpire) {
				t.Errorf("expiresAt got %s want %s", e, tc.wantExpire)
			}
		})
	}
}

// Window is what `deals trigger --unlock-after-days/--valid-days` uses; it
// must agree with the checker's own windowFor.
func TestWindowMatchesWindowFor(t *testing.T) {
	game := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	r := Rule{UnlockAfterDays: 2, ValidDays: 7}
	a, e := Window(game, r)
	wa, we := windowFor(game, r, pt(t))
	if !a.Equal(wa) || !e.Equal(we) {
		t.Fatalf("Window = %s..%s, windowFor = %s..%s", a, e, wa, we)
	}
}

package triggers

import (
	"testing"
	"time"
)

// Observed US federal holidays, checked against the OPM calendar. A Saturday
// holiday is observed the Friday before, a Sunday one the Monday after; New
// Year's Day 2028 (a Saturday) is observed Fri Dec 31 2027.
var observedHolidays = map[string]string{
	"2026-01-01": "New Year's Day",
	"2026-01-19": "Martin Luther King Jr. Day",
	"2026-02-16": "Presidents Day",
	"2026-05-25": "Memorial Day",
	"2026-06-19": "Juneteenth",
	"2026-07-03": "Independence Day (observed; Jul 4 is a Saturday)",
	"2026-09-07": "Labor Day",
	"2026-10-12": "Columbus Day",
	"2026-11-11": "Veterans Day",
	"2026-11-26": "Thanksgiving Day",
	"2026-12-25": "Christmas Day",
	"2027-01-01": "New Year's Day",
	"2027-01-18": "Martin Luther King Jr. Day",
	"2027-02-15": "Presidents Day",
	"2027-05-31": "Memorial Day",
	"2027-06-18": "Juneteenth (observed; Jun 19 is a Saturday)",
	"2027-07-05": "Independence Day (observed; Jul 4 is a Sunday)",
	"2027-09-06": "Labor Day",
	"2027-10-11": "Columbus Day",
	"2027-11-11": "Veterans Day",
	"2027-11-25": "Thanksgiving Day",
	"2027-12-24": "Christmas Day (observed; Dec 25 is a Saturday)",
	"2027-12-31": "New Year's Day 2028 (observed; Jan 1 is a Saturday)",
}

// Every day of 2026 and 2027: a business day is a weekday that is not an
// observed federal holiday.
func TestIsBusinessDayEveryDay2026And2027(t *testing.T) {
	pt, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	holidays := 0
	for d := time.Date(2026, 1, 1, 0, 0, 0, 0, pt); d.Year() < 2028; d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		name, holiday := observedHolidays[key]
		weekend := d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
		want := !weekend && !holiday
		// Any time of the day in the deal's zone gives the same answer.
		for _, h := range []int{0, 12, 23} {
			at := time.Date(d.Year(), d.Month(), d.Day(), h, 30, 0, 0, pt)
			if got := IsBusinessDay(at); got != want {
				t.Errorf("%s %s %02d:30 (%s): IsBusinessDay = %v, want %v", key, d.Weekday(), h, name, got, want)
			}
		}
		if holiday {
			holidays++
			if weekend {
				t.Errorf("%s: observed holiday on a weekend in the table", key)
			}
		}
	}
	if holidays != len(observedHolidays) {
		t.Errorf("checked %d holidays, table has %d", holidays, len(observedHolidays))
	}
}

func TestIsBusinessDayCases(t *testing.T) {
	d := func(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 9, 0, 0, 0, time.UTC) }
	for _, tc := range []struct {
		name string
		at   time.Time
		want bool
	}{
		{"Wed before Thanksgiving 2026", d(2026, 11, 25), true},
		{"Thanksgiving 2026", d(2026, 11, 26), false},
		{"Black Friday 2026 is a business day", d(2026, 11, 27), true},
		{"Sat Jul 4 2026 is a weekend", d(2026, 7, 4), false},
		{"Fri Jul 3 2026 observed", d(2026, 7, 3), false},
		{"Sun Jul 4 2027 is a weekend", d(2027, 7, 4), false},
		{"Mon Jul 5 2027 observed", d(2027, 7, 5), false},
		{"Tue Jul 6 2027", d(2027, 7, 6), true},
		{"Fri Dec 31 2027 observes New Year's 2028", d(2027, 12, 31), false},
		{"Sat Jan 1 2028", d(2028, 1, 1), false},
		{"Mon Jan 3 2028", d(2028, 1, 3), true},
		{"Christmas Eve 2026 (Thu) is a business day", d(2026, 12, 24), true},
		{"Fri Dec 24 2027 observes Christmas", d(2027, 12, 24), false},
		{"Day after Christmas 2026 (Sat)", d(2026, 12, 26), false},
		{"Mon Dec 28 2026", d(2026, 12, 28), true},
	} {
		if got := IsBusinessDay(tc.at); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

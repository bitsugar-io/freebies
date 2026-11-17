package triggers

import "time"

// IsBusinessDay reports whether t's calendar date (in t's own location) is a
// US business day: a weekday that is not a US federal holiday, as observed.
// A holiday on a Saturday is observed the Friday before, one on a Sunday the
// Monday after, so the observed day is not a business day.
//
// Deals that say "next business day" mean the next day the partner's offices
// and banks are open; the sponsors follow the federal calendar.
func IsBusinessDay(t time.Time) bool {
	if wd := t.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return false
	}
	return !isObservedFederalHoliday(t.Year(), t.Month(), t.Day())
}

// isObservedFederalHoliday reports whether y-m-d is the observed date of a US
// federal holiday (5 U.S.C. 6103).
func isObservedFederalHoliday(y int, m time.Month, d int) bool {
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	// New Year's Day of the following year can be observed on Dec 31.
	for _, year := range []int{y, y + 1} {
		for _, h := range federalHolidays(year) {
			if observed(h).Equal(date) {
				return true
			}
		}
	}
	return false
}

// federalHolidays returns the actual dates (before weekend observance) of the
// eleven US federal holidays in year y, at midnight UTC.
func federalHolidays(y int) []time.Time {
	fixed := func(m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }
	return []time.Time{
		fixed(time.January, 1),                         // New Year's Day
		nthWeekday(y, time.January, time.Monday, 3),    // Martin Luther King Jr. Day
		nthWeekday(y, time.February, time.Monday, 3),   // Presidents Day
		lastWeekday(y, time.May, time.Monday),          // Memorial Day
		fixed(time.June, 19),                           // Juneteenth
		fixed(time.July, 4),                            // Independence Day
		nthWeekday(y, time.September, time.Monday, 1),  // Labor Day
		nthWeekday(y, time.October, time.Monday, 2),    // Columbus Day
		fixed(time.November, 11),                       // Veterans Day
		nthWeekday(y, time.November, time.Thursday, 4), // Thanksgiving Day
		fixed(time.December, 25),                       // Christmas Day
	}
}

// observed moves a Saturday holiday to Friday and a Sunday one to Monday.
func observed(h time.Time) time.Time {
	switch h.Weekday() {
	case time.Saturday:
		return h.AddDate(0, 0, -1)
	case time.Sunday:
		return h.AddDate(0, 0, 1)
	}
	return h
}

// nthWeekday is the nth wd of month m in year y (n starts at 1).
func nthWeekday(y int, m time.Month, wd time.Weekday, n int) time.Time {
	first := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	offset := (int(wd) - int(first.Weekday()) + 7) % 7
	return first.AddDate(0, 0, offset+7*(n-1))
}

// lastWeekday is the last wd of month m in year y.
func lastWeekday(y int, m time.Month, wd time.Weekday) time.Time {
	last := time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC)
	offset := (int(last.Weekday()) - int(wd) + 7) % 7
	return last.AddDate(0, 0, -offset)
}

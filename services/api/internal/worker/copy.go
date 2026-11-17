package worker

import (
	"fmt"
	"strings"
	"time"
)

// dealDeepLink is the URL scheme the mobile app opens for a triggered deal.
func dealDeepLink(triggeredEventID string) string {
	return "freebie://deal/" + triggeredEventID
}

// dealCopy is the data needed to word a push about a deal.
type dealCopy struct {
	Partner   string
	Offer     string
	Condition string // why it triggered, e.g. "Ducks win at home"
	ExpiresAt time.Time
	HasExpiry bool
	// AvailableAt is when a delayed deal's code unlocks; zero when usable now.
	AvailableAt time.Time
}

var pacific = func() *time.Location {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// formatClock renders a time as "10:30am" or "3pm".
func formatClock(t time.Time) string {
	t = t.In(pacific)
	if t.Minute() == 0 {
		return t.Format("3pm")
	}
	return t.Format("3:04pm")
}

// formatDeadline renders the deadline relative to now in Pacific time:
// "10:30am" today, "tomorrow 11:59pm", otherwise the dated last day, e.g.
// "Mon Nov 23 11:59pm". Deadlines are always on their last valid day; a
// deadline at exactly midnight is read as the end of the day before.
func formatDeadline(expires, now time.Time) string {
	e, n := expires.In(pacific), now.In(pacific)
	if e.Hour() == 0 && e.Minute() == 0 && e.Second() == 0 {
		e = e.Add(-time.Second)
	}
	switch {
	case e.Format("2006-01-02") == n.Format("2006-01-02"):
		return formatClock(e)
	case e.Format("2006-01-02") == n.AddDate(0, 0, 1).Format("2006-01-02"):
		return "tomorrow " + formatClock(e)
	default:
		return e.Format("Mon Jan 2") + " " + formatClock(e)
	}
}

// formatUnlock renders when a delayed code unlocks, as a phrase after
// "Code": "starts at 10am" later the same Pacific day, "unlocks tomorrow",
// or "unlocks Tue" on a later day.
func formatUnlock(at, now time.Time) string {
	a, n := at.In(pacific), now.In(pacific)
	switch a.Format("2006-01-02") {
	case n.Format("2006-01-02"):
		return "starts at " + formatClock(a)
	case n.AddDate(0, 0, 1).Format("2006-01-02"):
		return "unlocks tomorrow"
	default:
		return "unlocks " + a.Format("Mon")
	}
}

// formatRemaining renders time left as "3h", "45m" or "2d".
func formatRemaining(d time.Duration) string {
	switch {
	case d >= 48*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		m := int(d.Minutes())
		if m < 1 {
			m = 1
		}
		return fmt.Sprintf("%dm", m)
	}
}

// unlockedMessage words the push sent when a deal triggers: the reason and
// the offer in the title, where and by when to claim in the body.
func unlockedMessage(c dealCopy, now time.Time) (title, body string) {
	title = c.Offer
	if c.Condition != "" {
		title = fmt.Sprintf("%s! %s", c.Condition, c.Offer)
	}
	body = fmt.Sprintf("Claim at %s", c.Partner)
	if !c.AvailableAt.IsZero() && c.AvailableAt.After(now) {
		body = fmt.Sprintf("Code %s. Claim at %s", formatUnlock(c.AvailableAt, now), c.Partner)
	}
	if c.HasExpiry {
		body += " by " + formatDeadline(c.ExpiresAt, now)
	}
	return title, body + "."
}

// reminderMessage words the push sent as a deal nears its deadline.
func reminderMessage(c dealCopy, now time.Time) (title, body string) {
	// The unlock day, before the start time (a deal that starts after the last
	// reminder run is reminded early): say when it starts.
	if !c.AvailableAt.IsZero() && pacificDay(c.AvailableAt) == pacificDay(now) && c.AvailableAt.After(now) {
		title = fmt.Sprintf("Your %s code starts at %s", c.Offer, formatClock(c.AvailableAt))
		body = fmt.Sprintf("Claim at %s", c.Partner)
		if c.HasExpiry {
			body += " by " + formatDeadline(c.ExpiresAt, now)
		}
		return title, body + "."
	}
	// The day a delayed code unlocks, that is the news: "it works now".
	if !c.AvailableAt.IsZero() && pacificDay(c.AvailableAt) == pacificDay(now) && !c.AvailableAt.After(now) {
		title = fmt.Sprintf("Your %s code works now", c.Offer)
		body = fmt.Sprintf("Claim at %s", c.Partner)
		if c.HasExpiry {
			body += " by " + formatDeadline(c.ExpiresAt, now)
		}
		return title, body + "."
	}
	title = fmt.Sprintf("Your %s", c.Offer)
	if c.HasExpiry {
		title += " expires in " + formatRemaining(c.ExpiresAt.Sub(now))
	} else {
		title += " is still waiting"
	}
	body = fmt.Sprintf("Claim at %s", c.Partner)
	if c.HasExpiry {
		body += " by " + formatDeadline(c.ExpiresAt, now)
	}
	if c.Condition != "" {
		body += " (" + strings.TrimSpace(c.Condition) + ")"
	}
	return title, body + "."
}

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

// formatUnlock renders when a delayed code unlocks: "tomorrow" or "Tue".
func formatUnlock(at, now time.Time) string {
	a, n := at.In(pacific), now.In(pacific)
	if a.Format("2006-01-02") == n.AddDate(0, 0, 1).Format("2006-01-02") {
		return "tomorrow"
	}
	return a.Format("Mon")
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
		body = fmt.Sprintf("Code unlocks %s. Claim at %s", formatUnlock(c.AvailableAt, now), c.Partner)
	}
	if c.HasExpiry {
		body += " by " + formatDeadline(c.ExpiresAt, now)
	}
	return title, body + "."
}

// reminderMessage words the push sent as a deal nears its deadline.
func reminderMessage(c dealCopy, now time.Time) (title, body string) {
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

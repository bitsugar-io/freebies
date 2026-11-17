package triggers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/sources"
)

// Rule represents a structured trigger rule
type Rule struct {
	Metric           string `json:"metric"`
	Scope            string `json:"scope"`
	Operator         string `json:"operator"`
	Value            int    `json:"value"`
	RedemptionWindow string `json:"redemption_window"`
	// ExpiresAtLocal optionally narrows the deadline to a Pacific-time clock
	// time ("HH:MM", 24h) on the day the window ends, e.g. "10:30" for a
	// breakfast-only deal. Empty means end of day (23:59:59).
	ExpiresAtLocal string `json:"expires_at_local,omitempty"`
	// UnlockAfterDays delays the deal: its code works from midnight Pacific
	// that many days after the game date (2 = "unlocks 2 days after the game").
	// Default 0: usable as soon as it triggers.
	UnlockAfterDays int `json:"unlock_after_days,omitempty"`
	// ValidDays is how many calendar days the deal lasts. With UnlockAfterDays
	// the first day is the unlock day (7 = "unlock day plus the next 6",
	// default 1); without, the first day is the day after the game, which is
	// when the check runs (3 = that day and the next two), and it replaces
	// RedemptionWindow. Every window ends at 23:59:59 Pacific on its last day.
	// expires_at_local still applies on the last day when UnlockAfterDays is 0;
	// it is ignored when the deal unlocks later.
	ValidDays int `json:"valid_days,omitempty"`
	// AvailableFromLocal delays the start to a Pacific clock time ("HH:MM") on
	// the deal's first usable day, for terms like "valid 10am-11:59pm" or "not
	// during breakfast". Empty: usable as soon as it triggers (or unlocks).
	AvailableFromLocal string `json:"available_from_local,omitempty"`
	// Season limits the deal to part of the season. "regular" ignores playoff
	// games ("during the 2026-2027 regular season"). Empty: any game.
	Season string `json:"season,omitempty"`
}

// Redemption windows a rule may name (see windowFor).
const (
	SameDay                = "same_day"
	NextDay                = "next_day"
	Week                   = "week"
	NextBusinessDay        = "next_business_day"
	ThroughNextBusinessDay = "through_next_business_day"
)

// KnownWindow reports whether w is a redemption window windowFor understands.
func KnownWindow(w string) bool {
	switch w {
	case SameDay, NextDay, Week, NextBusinessDay, ThroughNextBusinessDay, GameNight:
		return true
	}
	return false
}

// GameNight is the redemption window of an at-the-game deal claimed in the
// venue the night of the game (Chick-fil-A's arena deals). It ends at 23:59:59
// Pacific on the game date, so the next-morning check never alerts on it.
const GameNight = "game_night"

// Checker orchestrates trigger checking for events
type Checker struct {
	queries *db.Queries
	loc     *time.Location
	logger  *slog.Logger
	now     func() time.Time
}

// Option configures a Checker.
type Option func(*Checker)

// WithLogger sets the structured logger (default slog.Default()).
func WithLogger(l *slog.Logger) Option { return func(c *Checker) { c.logger = l } }

// WithClock sets the clock used to decide whether a deal's window has already
// closed; tests use it.
func WithClock(now func() time.Time) Option { return func(c *Checker) { c.now = now } }

// NewChecker creates a new trigger checker
func NewChecker(queries *db.Queries, opts ...Option) *Checker {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		loc = time.FixedZone("PT", -8*3600)
	}
	c := &Checker{queries: queries, loc: loc, logger: slog.Default(), now: time.Now}
	for _, o := range opts {
		o(c)
	}
	return c
}

// SkippedEvent is a deal the run did not record: it can never fire, or its
// redemption window had already closed when the run reached it.
type SkippedEvent struct {
	EventID string `json:"eventId"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

// StatusExpired marks a triggered deal whose window had already closed.
const StatusExpired = "EXPIRED"

// CheckResult contains the result of checking a single event
type CheckResult struct {
	EventID          string
	TeamID           string
	Triggered        bool
	TriggeredEventID string    // ID of created triggered_event (empty if already existed)
	ExpiresAt        time.Time // redemption deadline of the created triggered_event
	Expired          bool      // the rule fired but its window had already closed; nothing recorded
	AvailableAt      time.Time // when the deal unlocks; zero when usable immediately
	Rule             *Rule
	// Stats is the game the deal fired on, or the day's first game when it
	// did not fire.
	Stats *sources.GameStats
	// Games is every finished game the team played on the date (two in a
	// doubleheader), in order.
	Games []*sources.GameStats
	Event *db.Event
	Error error
}

// CheckAll checks all active events for trigger conditions (yesterday's games)
func (c *Checker) CheckAll(ctx context.Context) ([]CheckResult, []SkippedEvent, error) {
	yesterday := c.now().In(c.loc).AddDate(0, 0, -1)
	return c.CheckAllForDate(ctx, yesterday)
}

// CheckAllForDate checks all active events for the games on date (only its
// Y/M/D, read as a Pacific date, are used). It also returns the deals it
// skipped: ones that can never fire, and ones whose window had already closed.
func (c *Checker) CheckAllForDate(ctx context.Context, date time.Time) ([]CheckResult, []SkippedEvent, error) {
	// Get all active events with trigger rules
	events, err := c.queries.ListActiveEvents(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("listing active events: %w", err)
	}

	var results []CheckResult
	var skipped []SkippedEvent

	// Capability check: a deal whose league has no source, or whose metric the
	// source never emits, can never fire. Say so loudly and skip it instead of
	// failing the whole run or failing quietly every morning.
	broken := map[string]bool{}
	for _, row := range Failures(Audit(events)) {
		c.logger.Warn("deal can never fire; skipping it",
			"event_id", row.EventID, "status", string(row.Status), "reason", row.Detail)
		skipped = append(skipped, SkippedEvent{
			EventID: row.EventID, Status: string(row.Status), Reason: row.Detail,
		})
		broken[row.EventID] = true
	}

	for _, event := range events {
		if broken[event.ID] {
			continue
		}
		result := c.checkEventForDate(ctx, event, date)
		result.Event = &event

		if result.Error != nil {
			c.logger.Error("checking event failed", "event_id", event.ID, "error", result.Error)
			results = append(results, result)
			continue
		}

		if result.Triggered {
			if err := c.createTriggeredEvent(ctx, event, &result, date); err != nil {
				c.logger.Error("creating triggered event failed", "event_id", event.ID, "error", err)
				result.Error = err
			}
			if result.Expired {
				skipped = append(skipped, SkippedEvent{
					EventID: event.ID,
					Status:  StatusExpired,
					Reason:  "redemption window already closed at " + result.ExpiresAt.In(c.loc).Format(time.RFC3339),
				})
			}
		}
		results = append(results, result)
	}

	return results, skipped, nil
}

func (c *Checker) checkEventForDate(ctx context.Context, event db.Event, date time.Time) CheckResult {
	result := CheckResult{
		EventID: event.ID,
		TeamID:  event.TeamID,
	}

	// Parse the trigger rule
	if !event.TriggerRule.Valid || event.TriggerRule.String == "" {
		result.Error = fmt.Errorf("no trigger rule defined")
		return result
	}

	var rule Rule
	if err := json.Unmarshal([]byte(event.TriggerRule.String), &rule); err != nil {
		result.Error = fmt.Errorf("parsing trigger rule: %w", err)
		return result
	}
	result.Rule = &rule

	// Get the source for this league
	source, err := sources.Get(event.League)
	if err != nil {
		result.Error = fmt.Errorf("getting source: %w", err)
		return result
	}

	// Every finished game the team played that day: a doubleheader, or NCAA
	// football and basketball, has more than one.
	games, err := gamesByDate(ctx, source, event.TeamID, date)
	if err != nil {
		result.Error = fmt.Errorf("fetching game stats: %w", err)
		return result
	}
	if len(games) == 0 {
		// No game on this date
		return result
	}
	result.Games = games
	result.Stats = games[0]

	// The deal fires on the first game that meets the rule. A
	// regular-season-only deal never fires on a playoff game.
	for _, stats := range games {
		if !c.seasonAllows(rule, stats) {
			c.logger.Info("playoff game; deal is regular season only", "event_id", event.ID, "game", stats.GameID)
			continue
		}
		if evaluateRule(rule, stats) {
			result.Stats = stats
			result.Triggered = true
			break
		}
	}

	return result
}

// gamesByDate returns every finished game teamID played on date: all of them
// from a sources.MultiGameSource, else the one GetGameByDate finds.
func gamesByDate(ctx context.Context, src sources.Source, teamID string, date time.Time) ([]*sources.GameStats, error) {
	if m, ok := src.(sources.MultiGameSource); ok {
		return m.GetGamesByDate(ctx, teamID, date)
	}
	stats, err := src.GetGameByDate(ctx, teamID, date)
	if err != nil || stats == nil {
		return nil, err
	}
	return []*sources.GameStats{stats}, nil
}

// createTriggeredEvent records the deal for result and fills in its id and
// window. A deal whose window has already closed (a late or backfill run, say
// a breakfast-only deal checked at noon) is not recorded: result.Expired is set
// and nothing is created, so no push goes out for it.
func (c *Checker) createTriggeredEvent(
	ctx context.Context,
	event db.Event,
	result *CheckResult,
	gameDate time.Time,
) error {
	// A deal fires at most once per date (sponsors give one a day), so if any
	// of the day's games already triggered it, a rerun adds nothing, even when
	// a different game is now the first to meet the rule.
	dayGames := result.Games
	if len(dayGames) == 0 {
		dayGames = []*sources.GameStats{result.Stats}
	}
	for _, g := range dayGames {
		existing, err := c.queries.GetTriggeredEventByGameID(ctx, db.GetTriggeredEventByGameIDParams{
			EventID: event.ID,
			GameID:  sql.NullString{String: g.GameID, Valid: true},
		})
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("checking existing triggered event: %w", err)
		}
		if existing.ID != "" {
			c.logger.Info("triggered event already exists for this date",
				"event_id", event.ID, "game", g.GameID)
			return nil // Already exists, don't notify again
		}
	}

	availableAt, expiresAt := windowFor(gameDate, *result.Rule, c.loc)
	result.AvailableAt, result.ExpiresAt = availableAt, expiresAt
	if result.Rule.RedemptionWindow == GameNight {
		// Claimed at the venue that night; the morning check is always too
		// late, so it is never recorded or pushed. The app lists it on game days.
		result.Expired = true
		c.logger.Info("at-the-game deal triggered; claimed at the venue, no alert",
			"event_id", event.ID, "game", result.Stats.GameID)
		return nil
	}
	if !expiresAt.After(c.now()) {
		result.Expired = true
		c.logger.Warn("deal window already closed; not recording it",
			"event_id", event.ID,
			"status", StatusExpired,
			"reason", "redemption window closed before this run",
			"game", result.Stats.GameID,
			"expires_at", expiresAt.In(c.loc).Format(time.RFC3339),
		)
		return nil
	}
	available := sql.NullTime{}
	if !availableAt.IsZero() {
		available = sql.NullTime{Time: availableAt, Valid: true}
	}

	// Create the triggered event
	triggeredID := fmt.Sprintf("triggered-%s-%s", event.ID, result.Stats.GameID)
	_, err := c.queries.CreateTriggeredEvent(ctx, db.CreateTriggeredEventParams{
		ID:          triggeredID,
		EventID:     event.ID,
		GameID:      sql.NullString{String: result.Stats.GameID, Valid: true},
		ExpiresAt:   sql.NullTime{Time: expiresAt, Valid: true},
		AvailableAt: available,
	})
	if err != nil {
		return fmt.Errorf("creating triggered event: %w", err)
	}

	result.TriggeredEventID = triggeredID
	c.logger.Info("created triggered event",
		"triggered_event_id", triggeredID,
		"available_at", availableAt.Format(time.RFC3339),
		"expires_at", expiresAt.Format(time.RFC3339))
	return nil
}

// Window returns the unlock time (zero if usable at once) and deadline, in UTC,
// the checker would give a deal with rule r for a game on gameDate (Pacific
// Y/M/D). `deals trigger` uses it so a hand-made deal gets the real window.
func Window(gameDate time.Time, r Rule) (availableAt, expiresAt time.Time) {
	return windowFor(gameDate, r, NewChecker(nil).loc)
}

// windowFor returns when a triggered deal unlocks (zero if immediately) and
// when it expires, both UTC. Every window is computed from the game's Pacific
// date (the Y/M/D of gameDate), never from when the check happens to run, and
// ends at 23:59:59 Pacific on its last valid day (or at expires_at_local on
// that day for windows that honor it).
//
// Day 1 is the day after the game, when the morning check normally runs:
// same_day lasts through day 1, next_day through day 2, week through day 7,
// valid_days N through day N. game_night ends on the game date itself.
// next_business_day is that weekday only (locked until then);
// through_next_business_day is usable now and ends with it.
// available_from_local then moves the start to a clock time on the first
// usable day.
func windowFor(gameDate time.Time, r Rule, loc *time.Location) (availableAt, expiresAt time.Time) {
	availableAt, expiresAt = baseWindow(gameDate, r, loc)
	if r.RedemptionWindow == GameNight || r.AvailableFromLocal == "" {
		return availableAt, expiresAt
	}
	// "Valid from 10am": the deal starts at that clock time on its first usable day.
	t, err := time.Parse("15:04", r.AvailableFromLocal)
	if err != nil {
		return availableAt, expiresAt
	}
	first := availableAt
	if first.IsZero() {
		first = time.Date(gameDate.Year(), gameDate.Month(), gameDate.Day()+1, 0, 0, 0, 0, loc)
	}
	first = first.In(loc)
	from := time.Date(first.Year(), first.Month(), first.Day(), t.Hour(), t.Minute(), 0, 0, loc).UTC()
	if from.After(availableAt) && from.Before(expiresAt) {
		availableAt = from
	}
	return availableAt, expiresAt
}

// baseWindow is windowFor before any available_from_local start time.
func baseWindow(gameDate time.Time, r Rule, loc *time.Location) (availableAt, expiresAt time.Time) {
	// day(n) is midnight Pacific n days after the game date.
	day := func(n int) time.Time {
		return time.Date(gameDate.Year(), gameDate.Month(), gameDate.Day()+n, 0, 0, 0, 0, loc)
	}
	// lastMoment is the deadline on d: 23:59:59, or atLocal ("HH:MM") if valid.
	lastMoment := func(d time.Time, atLocal string) time.Time {
		hour, min, sec := 23, 59, 59
		if t, err := time.Parse("15:04", atLocal); err == nil {
			hour, min, sec = t.Hour(), t.Minute(), 0
		}
		return time.Date(d.Year(), d.Month(), d.Day(), hour, min, sec, 0, loc).UTC()
	}

	switch {
	case r.UnlockAfterDays > 0:
		days := r.ValidDays
		if days < 1 {
			days = 1
		}
		unlock := day(r.UnlockAfterDays)
		return unlock.UTC(), lastMoment(day(r.UnlockAfterDays+days-1), "")
	case r.ValidDays > 0:
		return time.Time{}, lastMoment(day(r.ValidDays), r.ExpiresAtLocal)
	case r.RedemptionWindow == "next_business_day":
		// "Valid the next business day after the game": the first business day
		// (not a weekend or US federal holiday) after the game date, whole day (or until expires_at_local). It is a
		// delayed unlock only when that weekday is later than day 1.
		n := 1
		for !IsBusinessDay(day(n)) { // skips weekends and US federal holidays
			n++
		}
		if n > 1 {
			availableAt = day(n).UTC()
		}
		return availableAt, lastMoment(day(n), r.ExpiresAtLocal)
	case r.RedemptionWindow == ThroughNextBusinessDay:
		// "Valid through the following business day": usable now, until the end
		// of the first weekday after the game date (Sat game: through Monday).
		n := 1
		for !IsBusinessDay(day(n)) { // skips weekends and US federal holidays
			n++
		}
		return time.Time{}, lastMoment(day(n), r.ExpiresAtLocal)
	case r.RedemptionWindow == GameNight:
		return time.Time{}, lastMoment(day(0), r.ExpiresAtLocal)
	case r.RedemptionWindow == "same_day":
		return time.Time{}, lastMoment(day(1), r.ExpiresAtLocal)
	case r.RedemptionWindow == "week":
		return time.Time{}, lastMoment(day(7), "")
	default: // "next_day" and anything unknown
		return time.Time{}, lastMoment(day(2), r.ExpiresAtLocal)
	}
}

// seasonAllows reports whether the rule's season limit admits this game:
// "regular" excludes playoff games; no limit admits every game.
func (c *Checker) seasonAllows(r Rule, stats *sources.GameStats) bool {
	return !(r.Season == "regular" && stats.Postseason)
}

// evaluateRule checks if the game stats satisfy the rule
func evaluateRule(rule Rule, stats *sources.GameStats) bool {
	// Get the metric value from stats
	metricValue, ok := stats.Metrics[rule.Metric]
	if !ok {
		return false
	}

	// Compare based on operator
	switch rule.Operator {
	case ">=":
		return metricValue >= rule.Value
	case ">":
		return metricValue > rule.Value
	case "<=":
		return metricValue <= rule.Value
	case "<":
		return metricValue < rule.Value
	case "==", "=":
		return metricValue == rule.Value
	case "!=":
		return metricValue != rule.Value
	default:
		return false
	}
}

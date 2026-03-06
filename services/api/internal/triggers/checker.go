package triggers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
}

// Checker orchestrates trigger checking for events
type Checker struct {
	queries *db.Queries
	loc     *time.Location
}

// NewChecker creates a new trigger checker
func NewChecker(queries *db.Queries) *Checker {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	return &Checker{
		queries: queries,
		loc:     loc,
	}
}

// CheckResult contains the result of checking a single event
type CheckResult struct {
	EventID          string
	TeamID           string
	Triggered        bool
	TriggeredEventID string // ID of created triggered_event (empty if already existed)
	Rule             *Rule
	Stats            *sources.GameStats
	Event            *db.Event
	Error            error
}

// CheckAll checks all active events for trigger conditions (yesterday's games)
func (c *Checker) CheckAll(ctx context.Context) ([]CheckResult, error) {
	yesterday := time.Now().In(c.loc).AddDate(0, 0, -1)
	return c.CheckAllForDate(ctx, yesterday)
}

// CheckAllForDate checks all active events for a specific date
func (c *Checker) CheckAllForDate(ctx context.Context, date time.Time) ([]CheckResult, error) {
	// Get all active events with trigger rules
	events, err := c.queries.ListActiveEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing active events: %w", err)
	}

	var results []CheckResult

	for _, event := range events {
		result := c.checkEventForDate(ctx, event, date)
		result.Event = &event

		if result.Error != nil {
			log.Printf("Error checking event %s: %v", event.ID, result.Error)
			results = append(results, result)
			continue
		}

		if result.Triggered {
			// Create triggered event
			triggeredID, err := c.createTriggeredEvent(ctx, event, result)
			if err != nil {
				log.Printf("Error creating triggered event for %s: %v", event.ID, err)
				result.Error = err
			}
			result.TriggeredEventID = triggeredID
		}
		results = append(results, result)
	}

	return results, nil
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

	// Fetch game stats for the specified date
	stats, err := source.GetGameByDate(ctx, event.TeamID, date)
	if err != nil {
		result.Error = fmt.Errorf("fetching game stats: %w", err)
		return result
	}

	if stats == nil {
		// No game on this date
		return result
	}
	result.Stats = stats

	// Evaluate the rule
	result.Triggered = evaluateRule(rule, stats)

	return result
}

func (c *Checker) createTriggeredEvent(ctx context.Context, event db.Event, result CheckResult) (string, error) {
	// Check if we already created a triggered event for this game
	existing, err := c.queries.GetTriggeredEventByGameID(ctx, db.GetTriggeredEventByGameIDParams{
		EventID: event.ID,
		GameID:  sql.NullString{String: result.Stats.GameID, Valid: true},
	})
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("checking existing triggered event: %w", err)
	}
	if existing.ID != "" {
		log.Printf("Triggered event already exists for game %s", result.Stats.GameID)
		return "", nil // Already exists, don't notify again
	}

	// Calculate expiration based on redemption window
	expiresAt := c.calculateExpiration(result.Rule.RedemptionWindow)

	// Create the triggered event
	triggeredID := fmt.Sprintf("triggered-%s-%s", event.ID, result.Stats.GameID)
	_, err = c.queries.CreateTriggeredEvent(ctx, db.CreateTriggeredEventParams{
		ID:        triggeredID,
		EventID:   event.ID,
		GameID:    sql.NullString{String: result.Stats.GameID, Valid: true},
		ExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("creating triggered event: %w", err)
	}

	log.Printf("Created triggered event %s (expires %s)", triggeredID, expiresAt.Format(time.RFC3339))
	return triggeredID, nil
}

func (c *Checker) calculateExpiration(window string) time.Time {
	now := time.Now().In(c.loc)

	switch window {
	case "next_day":
		// Expires at end of next day (11:59:59 PM PT)
		nextDay := now.AddDate(0, 0, 1)
		return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 23, 59, 59, 0, c.loc)
	case "same_day":
		// Expires at end of today
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, c.loc)
	case "week":
		// Expires in 7 days
		return now.AddDate(0, 0, 7)
	default:
		// Default to next day
		nextDay := now.AddDate(0, 0, 1)
		return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 23, 59, 59, 0, c.loc)
	}
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

package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

// Notification kinds recorded in notifications.kind.
const (
	KindTrigger  = "trigger"  // the deal unlocked
	KindReminder = "reminder" // a nudge while it is still claimable
)

// CheckTriggersResult contains the summary of a check-triggers run.
type CheckTriggersResult struct {
	Triggered   int `json:"triggered"`
	Notified    int `json:"notified"`
	TotalEvents int `json:"totalEvents"`
	// SkippedEvents counts deals the run did not record: ones that can never
	// fire (no source, unknown metric) and ones whose window had closed.
	SkippedEvents int                     `json:"skippedEvents"`
	Skipped       []triggers.SkippedEvent `json:"skipped"`
}

// SendRemindersResult contains the summary of a send-reminders run.
type SendRemindersResult struct {
	Sent          int64 `json:"sent"`
	Failed        int64 `json:"failed"`
	ExpiringDeals int   `json:"expiringDeals"`
}

// Service encapsulates the worker logic for checking triggers and sending reminders.
type Service struct {
	queries  *db.Queries
	checker  *triggers.Checker
	notifier *notify.ExpoNotifier
	logger   *slog.Logger
	now      func() time.Time // clock; nil means time.Now
}

func (s *Service) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

// NewService creates a new worker service.
func NewService(queries *db.Queries, checker *triggers.Checker, notifier *notify.ExpoNotifier, logger *slog.Logger) *Service {
	return &Service{
		queries:  queries,
		checker:  checker,
		notifier: notifier,
		logger:   logger,
	}
}

// CheckTriggers checks game results for the given date and creates triggered events.
// If date is zero, it checks yesterday's games.
func (s *Service) CheckTriggers(ctx context.Context, date time.Time) (*CheckTriggersResult, error) {
	var results []triggers.CheckResult
	var skipped []triggers.SkippedEvent
	var err error

	if date.IsZero() {
		s.logger.Info("checking triggers for yesterday")
		results, skipped, err = s.checker.CheckAll(ctx)
	} else {
		s.logger.Info("checking triggers for date", "date", date.Format("2006-01-02"))
		results, skipped, err = s.checker.CheckAllForDate(ctx, date)
	}
	if err != nil {
		return nil, fmt.Errorf("checking triggers: %w", err)
	}

	triggered := 0
	notified := 0
	for _, result := range results {
		if result.Error != nil {
			s.logger.Warn("check failed",
				"event_id", result.EventID,
				"team", result.TeamID,
				"error", result.Error,
			)
			continue
		}

		if result.Expired {
			// Already reported in skipped; nothing was recorded or pushed.
			continue
		}

		if result.Stats == nil {
			s.logger.Info("no game found",
				"event_id", result.EventID,
				"team", result.TeamID,
				"rule", fmt.Sprintf("%s %s %d", result.Rule.Metric, result.Rule.Operator, result.Rule.Value),
			)
			continue
		}

		metricValue := result.Stats.Metrics[result.Rule.Metric]
		if result.Triggered {
			triggered++
			s.logger.Info("TRIGGERED",
				"event_id", result.EventID,
				"team", result.TeamID,
				"opponent", result.Stats.Opponent,
				"metric", result.Rule.Metric,
				"value", metricValue,
				"required", fmt.Sprintf("%s %d", result.Rule.Operator, result.Rule.Value),
				"game", result.Stats.GameID,
			)

			if result.TriggeredEventID != "" {
				sent := s.notifySubscribers(ctx, result)
				notified += sent
			}
		} else {
			s.logger.Info("not triggered",
				"event_id", result.EventID,
				"team", result.TeamID,
				"opponent", result.Stats.Opponent,
				"metric", result.Rule.Metric,
				"value", metricValue,
				"required", fmt.Sprintf("%s %d", result.Rule.Operator, result.Rule.Value),
			)
		}
	}

	res := &CheckTriggersResult{
		Triggered:   triggered,
		Notified:    notified,
		TotalEvents: len(results),

		SkippedEvents: len(skipped),
		Skipped:       skipped,
	}
	if res.Skipped == nil {
		res.Skipped = []triggers.SkippedEvent{}
	}
	for _, sk := range skipped {
		s.logger.Warn("skipped deal", "event_id", sk.EventID, "status", sk.Status, "reason", sk.Reason)
	}
	s.logger.Info("check-triggers complete",
		"triggered", triggered, "notified", notified,
		"total_events", len(results), "skipped_events", len(skipped))
	return res, nil
}

// alertsPaused reports whether a user's alerts are paused at now.
func alertsPaused(until sql.NullTime, now time.Time) bool {
	return until.Valid && until.Time.After(now)
}

// pacificDay is the Pacific calendar day of t as "2006-01-02".
func pacificDay(t time.Time) string { return t.In(pacific).Format("2006-01-02") }

// reminderDue reports whether a deal is due a reminder today. A reminder goes
// out only on the day the deal first becomes claimable (it unlocks, or
// triggered if it was usable at once) and on its last day. A one-day deal
// therefore gets one reminder; a week-long deal gets two, not one per run.
func reminderDue(triggeredAt time.Time, availableAt, expiresAt sql.NullTime, now time.Time) bool {
	first := triggeredAt
	if availableAt.Valid && availableAt.Time.After(first) {
		first = availableAt.Time
	}
	today := pacificDay(now)
	if today == pacificDay(first) {
		return true
	}
	// An expiry at exactly midnight belongs to the day before it.
	return expiresAt.Valid && today == pacificDay(expiresAt.Time.Add(-time.Nanosecond))
}

// SendReminders sends reminder notifications for active deals. Each user gets
// at most one reminder per deal per Pacific day, and only on the deal's first
// claimable day and its last day (see reminderDue). What was already sent is
// read from the notifications table (kind "reminder", status "sent").
func (s *Service) SendReminders(ctx context.Context) (*SendRemindersResult, error) {
	activeDeals, err := s.queries.ListActiveTriggeredEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing active deals: %w", err)
	}

	s.logger.Info("checking for active deals", "count", len(activeDeals))

	var totalSent, totalFailed int64
	now := s.clock() // one instant for the whole pass
	dayStart := time.Date(now.In(pacific).Year(), now.In(pacific).Month(), now.In(pacific).Day(),
		0, 0, 0, 0, pacific).UTC().Format("2006-01-02 15:04:05")

	for _, deal := range activeDeals {
		// A delayed deal's code is not usable yet; nothing to remind about.
		if deal.AvailableAt.Valid && deal.AvailableAt.Time.After(now) {
			continue
		}
		if !reminderDue(deal.TriggeredAt, deal.AvailableAt, deal.ExpiresAt, now) {
			continue
		}
		users, err := s.queries.ListUsersForReminder(ctx, deal.ID)
		if err != nil {
			s.logger.Error("failed to list users for reminder", "deal_id", deal.ID, "error", err)
			continue
		}
		already, err := s.queries.ListUsersRemindedSince(ctx, db.ListUsersRemindedSinceParams{
			TriggeredEventID: deal.ID,
			Since:            dayStart,
		})
		if err != nil {
			s.logger.Error("failed to list reminded users", "deal_id", deal.ID, "error", err)
			continue
		}
		reminded := make(map[string]bool, len(already))
		for _, id := range already {
			reminded[id] = true
		}

		title, body := reminderMessage(dealCopy{
			Partner:   deal.PartnerName,
			Offer:     deal.OfferName,
			Condition: deal.TriggerCondition,
			ExpiresAt: deal.ExpiresAt.Time,
			HasExpiry: deal.ExpiresAt.Valid,
		}, now)
		data := map[string]interface{}{
			"triggeredEventId": deal.ID,
			"eventId":          deal.EventID,
			"url":              dealDeepLink(deal.ID),
		}

		var messages []notify.ExpoPushMessage
		for _, user := range users {
			if !user.PushToken.Valid || user.PushToken.String == "" {
				continue
			}
			if alertsPaused(user.AlertsPausedUntil, now) {
				continue
			}
			if reminded[user.ID] {
				continue
			}
			messages = append(messages, notify.ExpoPushMessage{
				To:       user.PushToken.String,
				Title:    title,
				Body:     body,
				Data:     data,
				Sound:    "default",
				Priority: "high",
			})
		}

		if len(messages) == 0 {
			continue
		}

		messages = notify.DeduplicateMessages(messages)

		userByToken := make(map[string]string, len(users))
		for _, u := range users {
			if u.PushToken.Valid {
				userByToken[u.PushToken.String] = u.ID
			}
		}

		batchResult := s.notifier.SendBatchConcurrent(ctx, s.logger, messages, notify.DefaultWorkers)
		s.recordNotifications(ctx, messages, KindReminder, deal.ID, userByToken, batchResult)
		totalSent += batchResult.Sent
		totalFailed += batchResult.Failed
	}

	res := &SendRemindersResult{
		Sent:          totalSent,
		Failed:        totalFailed,
		ExpiringDeals: len(activeDeals),
	}
	s.logger.Info("send-reminders complete", "sent", totalSent, "failed", totalFailed, "active_deals", len(activeDeals))
	return res, nil
}

func (s *Service) notifySubscribers(ctx context.Context, result triggers.CheckResult) int {
	now := s.clock()
	subscribers, err := s.queries.ListEventSubscribers(ctx, result.EventID)
	if err != nil {
		s.logger.Error("failed to list subscribers", "event_id", result.EventID, "error", err)
		return 0
	}

	title, body := unlockedMessage(dealCopy{
		Partner:   result.Event.PartnerName,
		Offer:     result.Event.OfferName,
		Condition: result.Event.TriggerCondition,
		ExpiresAt: result.ExpiresAt,
		HasExpiry: !result.ExpiresAt.IsZero(),

		AvailableAt: result.AvailableAt,
	}, now)
	data := map[string]interface{}{
		"triggeredEventId": result.TriggeredEventID,
		"eventId":          result.EventID,
		"url":              dealDeepLink(result.TriggeredEventID),
	}
	if !result.AvailableAt.IsZero() {
		data["availableAt"] = result.AvailableAt.UTC().Format(time.RFC3339)
	}

	var messages []notify.ExpoPushMessage
	for _, sub := range subscribers {
		if !sub.PushToken.Valid || sub.PushToken.String == "" {
			continue
		}
		// The triggered event is already recorded; paused users just get no push.
		if alertsPaused(sub.AlertsPausedUntil, now) {
			continue
		}
		messages = append(messages, notify.ExpoPushMessage{
			To:       sub.PushToken.String,
			Title:    title,
			Body:     body,
			Data:     data,
			Sound:    "default",
			Priority: "high",
		})
	}

	if len(messages) == 0 {
		return 0
	}

	messages = notify.DeduplicateMessages(messages)

	userByToken := make(map[string]string, len(subscribers))
	for _, sub := range subscribers {
		if sub.PushToken.Valid {
			userByToken[sub.PushToken.String] = sub.UserID
		}
	}

	batchResult := s.notifier.SendBatchConcurrent(ctx, s.logger, messages, notify.DefaultWorkers)
	s.recordNotifications(ctx, messages, KindTrigger, result.TriggeredEventID, userByToken, batchResult)

	if batchResult.Sent > 0 {
		s.logger.Info("notified subscribers",
			"event_id", result.EventID,
			"sent", batchResult.Sent,
			"failed", batchResult.Failed,
		)
	}

	return int(batchResult.Sent)
}

// recordNotifications writes one row to `notifications` per attempted push,
// with status reflecting whether the send succeeded. Status assignment is
// positional: the first batchResult.Sent messages are marked "sent" and the
// remainder "failed". SendBatchConcurrent only surfaces aggregate counts, so
// per-row attribution may be wrong on a partial-batch failure — the aggregate
// counts of "sent" vs "failed" rows are accurate, the per-user mapping is
// approximate. Good enough for telemetry; precise per-ticket attribution
// would require a notifier API change.
func (s *Service) recordNotifications(
	ctx context.Context,
	messages []notify.ExpoPushMessage,
	kind string,
	triggeredEventID string,
	userByToken map[string]string,
	batchResult *notify.BatchResult,
) {
	if triggeredEventID == "" {
		return
	}
	sentRemaining := batchResult.Sent
	for _, msg := range messages {
		userID, ok := userByToken[msg.To]
		if !ok {
			continue
		}
		status := "failed"
		if sentRemaining > 0 {
			status = "sent"
			sentRemaining--
		}
		if _, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
			ID:               uuid.New().String(),
			UserID:           userID,
			TriggeredEventID: triggeredEventID,
			Status:           status,
			Kind:             kind,
			SentAt:           s.clock().UTC(),
		}); err != nil {
			s.logger.Warn("failed to record notification",
				"user_id", userID,
				"triggered_event_id", triggeredEventID,
				"error", err,
			)
		}
	}
}

package worker

import (
	"context"
	"database/sql"
	"encoding/json"
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
	Sent   int64 `json:"sent"`
	Failed int64 `json:"failed"`
	// Unknown counts pushes Expo may or may not have delivered; they are not
	// sent again (see deliver).
	Unknown       int64 `json:"unknown"`
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

	// Send every "you won" push still owed: this run's new deals, and earlier
	// ones whose push failed (Expo down at 7am) or never went out (the process
	// died between recording the deal and sending). Only live deals qualify.
	notified += s.sendOwedTriggerPushes(ctx)

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

// pushable reports whether a deal's followers may be pushed about it, by the
// same rule as ListTriggerPushDeals: recorded by the checker (no payload) or by
// `deals trigger --notify` (payload "notify": true). Test deals from
// `deals create` and `deals trigger --notify=false` are never pushed.
func pushable(payload sql.NullString) bool {
	if !payload.Valid {
		return true
	}
	var p struct {
		Notify bool `json:"notify"`
	}
	return json.Unmarshal([]byte(payload.String), &p) == nil && p.Notify
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

	var totalSent, totalFailed, totalUnknown int64
	now := s.clock() // one instant for the whole pass
	dayStart := time.Date(now.In(pacific).Year(), now.In(pacific).Month(), now.In(pacific).Day(),
		0, 0, 0, 0, pacific).UTC().Format("2006-01-02 15:04:05")

	for _, deal := range activeDeals {
		// Test deals are never pushed, reminders included.
		if !pushable(deal.Payload) {
			continue
		}
		// A delayed deal's code is not usable yet; nothing to remind about,
		// except on the day it unlocks: a deal that starts later today (even
		// after the last reminder run) gets its reminder worded "starts at 6pm".
		if deal.AvailableAt.Valid && deal.AvailableAt.Time.After(now) &&
			pacificDay(deal.AvailableAt.Time) != pacificDay(now) {
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

			AvailableAt: deal.AvailableAt.Time,
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

		userByToken := make(map[string]string, len(users))
		for _, u := range users {
			if u.PushToken.Valid {
				userByToken[u.PushToken.String] = u.ID
			}
		}

		batchResult := s.deliver(ctx, messages, KindReminder, deal.ID, userByToken)
		totalSent += batchResult.Sent
		totalFailed += batchResult.Failed
		totalUnknown += batchResult.Unknown
	}

	res := &SendRemindersResult{
		Sent:          totalSent,
		Failed:        totalFailed,
		Unknown:       totalUnknown,
		ExpiringDeals: len(activeDeals),
	}
	s.logger.Info("send-reminders complete", "sent", totalSent, "failed", totalFailed,
		"unknown", totalUnknown, "active_deals", len(activeDeals))
	return res, nil
}

// triggerPushDeal is a deal as the "you won" push needs it.
type triggerPushDeal = db.ListTriggerPushDealsRow

// sendOwedTriggerPushes sends the "you won" push for every live deal to the
// followers who have not had it yet (ListUsersOwedTriggerPush) and returns
// how many went out. Each check-triggers run calls it, so a push that failed
// at 7am goes out at 9am, and so on until the deal expires.
func (s *Service) sendOwedTriggerPushes(ctx context.Context) int {
	deals, err := s.queries.ListTriggerPushDeals(ctx, sqlTimeUTC(s.clock()))
	if err != nil {
		s.logger.Error("failed to list deals owed a trigger push", "error", err)
		return 0
	}
	sent := 0
	for _, deal := range deals {
		sent += int(s.pushTrigger(ctx, deal).Sent)
	}
	return sent
}

// NotifyTriggered sends the "you won" push for one recorded deal, exactly as
// check-triggers would: same copy and deep link, paused users skipped, one
// notifications row per push, nobody pushed twice. `deals trigger --notify`
// uses it.
func (s *Service) NotifyTriggered(ctx context.Context, triggeredEventID string) (*notify.BatchResult, error) {
	deal, err := s.queries.GetTriggerPushDeal(ctx, triggeredEventID)
	if err != nil {
		return nil, fmt.Errorf("loading deal %s: %w", triggeredEventID, err)
	}
	return s.pushTrigger(ctx, triggerPushDeal(deal)), nil
}

// pushTrigger sends one deal's "you won" push to the followers still owed it.
func (s *Service) pushTrigger(ctx context.Context, deal triggerPushDeal) *notify.BatchResult {
	now := s.clock()
	users, err := s.queries.ListUsersOwedTriggerPush(ctx, deal.ID)
	if err != nil {
		s.logger.Error("failed to list users owed a trigger push", "triggered_event_id", deal.ID, "error", err)
		return &notify.BatchResult{}
	}

	title, body := unlockedMessage(dealCopy{
		Partner:   deal.PartnerName,
		Offer:     deal.OfferName,
		Condition: deal.TriggerCondition,
		ExpiresAt: deal.ExpiresAt.Time,
		HasExpiry: deal.ExpiresAt.Valid,

		AvailableAt: deal.AvailableAt.Time,
	}, now)
	data := map[string]interface{}{
		"triggeredEventId": deal.ID,
		"eventId":          deal.EventID,
		"url":              dealDeepLink(deal.ID),
	}
	if deal.AvailableAt.Valid {
		data["availableAt"] = deal.AvailableAt.Time.UTC().Format(time.RFC3339)
	}

	var messages []notify.ExpoPushMessage
	userByToken := make(map[string]string, len(users))
	for _, u := range users {
		// Paused users get no push now; a later run sends it if they resume
		// while the deal is still live.
		if alertsPaused(u.AlertsPausedUntil, now) {
			continue
		}
		userByToken[u.PushToken.String] = u.ID
		messages = append(messages, notify.ExpoPushMessage{
			To:       u.PushToken.String,
			Title:    title,
			Body:     body,
			Data:     data,
			Sound:    "default",
			Priority: "high",
		})
	}
	res := s.deliver(ctx, messages, KindTrigger, deal.ID, userByToken)
	if res.Sent > 0 || res.Failed > 0 || res.Unknown > 0 {
		s.logger.Info("notified subscribers",
			"event_id", deal.EventID,
			"triggered_event_id", deal.ID,
			"sent", res.Sent,
			"failed", res.Failed,
			"unknown", res.Unknown,
		)
	}
	return res
}

// sqlTimeUTC formats t the way SQLite's datetime() compares it.
func sqlTimeUTC(t time.Time) string { return t.UTC().Format("2006-01-02 15:04:05") }

// deliver sends messages and records one notifications row per user with
// that user's own outcome, taken from their own Expo ticket.
//
// Each row is written as 'pending' before the send and updated after it, so
// the record never under-counts what may have reached a phone:
//
//   - sent: Expo returned an ok ticket.
//   - failed: Expo rejected it or never got it. The next run retries it.
//   - unknown: Expo may have it (the response was lost, or a 5xx). Not retried.
//   - pending: the process died mid-send. Not retried.
//
// The rule is to prefer an occasional missed push over a duplicate one: only
// 'failed' is ever retried. A user whose row cannot be written is not sent
// to. A DeviceNotRegistered ticket clears that push token for good.
func (s *Service) deliver(
	ctx context.Context,
	messages []notify.ExpoPushMessage,
	kind string,
	triggeredEventID string,
	userByToken map[string]string,
) *notify.BatchResult {
	messages = notify.DeduplicateMessages(messages)
	rowIDs := make([]string, 0, len(messages))
	toSend := make([]notify.ExpoPushMessage, 0, len(messages))
	for _, msg := range messages {
		userID, ok := userByToken[msg.To]
		if !ok {
			continue
		}
		id := uuid.New().String()
		if _, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
			ID:               id,
			UserID:           userID,
			TriggeredEventID: triggeredEventID,
			Status:           "pending",
			Kind:             kind,
			SentAt:           s.clock().UTC(),
		}); err != nil {
			s.logger.Warn("failed to record notification; not sending it",
				"user_id", userID, "triggered_event_id", triggeredEventID, "error", err)
			continue
		}
		rowIDs = append(rowIDs, id)
		toSend = append(toSend, msg)
	}
	if len(toSend) == 0 {
		return &notify.BatchResult{}
	}

	res := s.notifier.SendBatchConcurrent(ctx, s.logger, toSend, notify.DefaultWorkers)
	for i, r := range res.Results {
		if err := s.queries.UpdateNotificationStatus(ctx, db.UpdateNotificationStatusParams{
			Status: string(r.Delivery),
			ID:     rowIDs[i],
		}); err != nil {
			s.logger.Warn("failed to record notification status",
				"user_id", userByToken[r.To], "triggered_event_id", triggeredEventID,
				"status", r.Delivery, "error", err)
		}
		if r.ErrorCode == notify.ErrDeviceNotRegistered {
			// The app is gone from that device; stop sending to the token.
			// The empty id excludes no one, so every user holding it loses it.
			if err := s.queries.ClearPushToken(ctx, db.ClearPushTokenParams{
				PushToken: sql.NullString{String: r.To, Valid: true},
				ID:        "",
			}); err != nil {
				s.logger.Warn("failed to clear dead push token", "user_id", userByToken[r.To], "error", err)
			} else {
				s.logger.Info("cleared dead push token", "user_id", userByToken[r.To])
			}
		}
	}
	for _, err := range res.Errors {
		s.logger.Warn("push batch error", "kind", kind, "triggered_event_id", triggeredEventID, "error", err)
	}
	return res
}

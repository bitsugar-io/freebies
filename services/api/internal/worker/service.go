package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

// CheckTriggersResult contains the summary of a check-triggers run.
type CheckTriggersResult struct {
	Triggered   int `json:"triggered"`
	Notified    int `json:"notified"`
	TotalEvents int `json:"totalEvents"`
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
	var err error

	if date.IsZero() {
		s.logger.Info("checking triggers for yesterday")
		results, err = s.checker.CheckAll(ctx)
	} else {
		s.logger.Info("checking triggers for date", "date", date.Format("2006-01-02"))
		results, err = s.checker.CheckAllForDate(ctx, date)
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
	}
	s.logger.Info("check-triggers complete", "triggered", triggered, "notified", notified, "total_events", len(results))
	return res, nil
}

// SendReminders sends reminder notifications for deals expiring soon.
func (s *Service) SendReminders(ctx context.Context) (*SendRemindersResult, error) {
	expiringDeals, err := s.queries.ListExpiringTriggeredEvents(ctx, sql.NullString{String: "6", Valid: true})
	if err != nil {
		return nil, fmt.Errorf("listing expiring deals: %w", err)
	}

	s.logger.Info("checking for expiring deals", "count", len(expiringDeals))

	var totalSent, totalFailed int64

	for _, deal := range expiringDeals {
		users, err := s.queries.ListUsersForReminder(ctx, deal.ID)
		if err != nil {
			s.logger.Error("failed to list users for reminder", "deal_id", deal.ID, "error", err)
			continue
		}

		timeRemaining := time.Until(deal.ExpiresAt.Time)
		hours := int(timeRemaining.Hours())

		title := fmt.Sprintf("⏰ Deal expires in %d hours!", hours)
		body := fmt.Sprintf("Don't forget: %s - %s", deal.OfferName, deal.PartnerName)
		data := map[string]interface{}{
			"triggeredEventId": deal.ID,
			"eventId":          deal.EventID,
		}

		var messages []notify.ExpoPushMessage
		for _, user := range users {
			if !user.PushToken.Valid || user.PushToken.String == "" {
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

		batchResult := s.notifier.SendBatchConcurrent(ctx, s.logger, messages, notify.DefaultWorkers)
		totalSent += batchResult.Sent
		totalFailed += batchResult.Failed
	}

	res := &SendRemindersResult{
		Sent:          totalSent,
		Failed:        totalFailed,
		ExpiringDeals: len(expiringDeals),
	}
	s.logger.Info("send-reminders complete", "sent", totalSent, "failed", totalFailed, "expiring_deals", len(expiringDeals))
	return res, nil
}

func (s *Service) notifySubscribers(ctx context.Context, result triggers.CheckResult) int {
	subscribers, err := s.queries.ListEventSubscribers(ctx, result.EventID)
	if err != nil {
		s.logger.Error("failed to list subscribers", "event_id", result.EventID, "error", err)
		return 0
	}

	icon := ""
	if result.Event.Icon.Valid {
		icon = result.Event.Icon.String + " "
	}
	title := fmt.Sprintf("%sDeal Unlocked!", icon)
	body := fmt.Sprintf("%s: %s", result.Event.PartnerName, result.Event.OfferName)
	data := map[string]interface{}{
		"triggeredEventId": result.TriggeredEventID,
		"eventId":          result.EventID,
	}

	var messages []notify.ExpoPushMessage
	for _, sub := range subscribers {
		if !sub.PushToken.Valid || sub.PushToken.String == "" {
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

	batchResult := s.notifier.SendBatchConcurrent(ctx, s.logger, messages, notify.DefaultWorkers)

	if batchResult.Sent > 0 {
		s.logger.Info("notified subscribers",
			"event_id", result.EventID,
			"sent", batchResult.Sent,
			"failed", batchResult.Failed,
		)
	}

	return int(batchResult.Sent)
}

package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/retr0h/freebie/internal/db"
	"github.com/retr0h/freebie/internal/notify"
	"github.com/retr0h/freebie/internal/triggers"

	// Register sources
	_ "github.com/retr0h/freebie/internal/sources/mlb"
)

// Scheduler runs background jobs on a schedule
type Scheduler struct {
	cron     *cron.Cron
	queries  *db.Queries
	checker  *triggers.Checker
	notifier *notify.ExpoNotifier
	logger   *slog.Logger
}

// New creates a new scheduler
func New(database *sql.DB, logger *slog.Logger) *Scheduler {
	loc, _ := time.LoadLocation("America/Los_Angeles")

	queries := db.New(database)
	return &Scheduler{
		cron:     cron.New(cron.WithLocation(loc)),
		queries:  queries,
		checker:  triggers.NewChecker(queries),
		notifier: notify.NewExpoNotifier(),
		logger:   logger,
	}
}

// Start begins the scheduled jobs and blocks until context is cancelled
func (s *Scheduler) Start(ctx context.Context) {
	// Check triggers at 6am PT daily
	s.cron.AddFunc("0 6 * * *", func() {
		s.logger.Info("scheduler: running check-triggers")
		s.runCheckTriggers()
	})

	// Send reminders at 6pm PT daily
	s.cron.AddFunc("0 18 * * *", func() {
		s.logger.Info("scheduler: running send-reminders")
		s.runSendReminders()
	})

	s.cron.Start()
	s.logger.Info("scheduler started",
		"timezone", "America/Los_Angeles",
		"jobs", []string{"check-triggers@6am", "send-reminders@6pm"},
	)

	// Wait for context cancellation
	<-ctx.Done()
	s.logger.Info("scheduler stopping")
	s.cron.Stop()
}

func (s *Scheduler) runCheckTriggers() {
	ctx := context.Background()

	results, err := s.checker.CheckAll(ctx)
	if err != nil {
		s.logger.Error("scheduler: check-triggers failed", "error", err)
		return
	}

	triggered := 0
	notified := 0
	for _, result := range results {
		if result.Error != nil || !result.Triggered {
			continue
		}
		triggered++

		// Only notify if this is a new triggered event (not already existed)
		if result.TriggeredEventID != "" {
			sent := s.notifySubscribers(ctx, result)
			notified += sent
		}
	}

	s.logger.Info("scheduler: check-triggers complete", "triggered", triggered, "notified", notified)
}

func (s *Scheduler) notifySubscribers(ctx context.Context, result triggers.CheckResult) int {
	subscribers, err := s.queries.ListEventSubscribers(ctx, result.EventID)
	if err != nil {
		s.logger.Error("failed to list subscribers", "event_id", result.EventID, "error", err)
		return 0
	}

	sent := 0
	for _, sub := range subscribers {
		if !sub.PushToken.Valid || sub.PushToken.String == "" {
			continue
		}

		icon := ""
		if result.Event.Icon.Valid {
			icon = result.Event.Icon.String + " "
		}

		title := fmt.Sprintf("%sDeal Unlocked!", icon)
		body := fmt.Sprintf("%s: %s", result.Event.PartnerName, result.Event.OfferName)
		data := map[string]interface{}{
			"triggered_event_id": result.TriggeredEventID,
			"event_id":           result.EventID,
		}

		_, err := s.notifier.Send(sub.PushToken.String, title, body, data)
		if err != nil {
			s.logger.Error("failed to send notification", "user_id", sub.UserID, "error", err)
			continue
		}
		sent++
	}

	if sent > 0 {
		s.logger.Info("notified subscribers", "event_id", result.EventID, "count", sent)
	}

	return sent
}

func (s *Scheduler) runSendReminders() {
	ctx := context.Background()

	expiringDeals, err := s.queries.ListExpiringTriggeredEvents(ctx, sql.NullString{String: "6", Valid: true})
	if err != nil {
		s.logger.Error("scheduler: send-reminders failed", "error", err)
		return
	}

	sent := 0
	for _, deal := range expiringDeals {
		// ListUsersForReminder already excludes users with "stop_reminding" dismissals
		users, err := s.queries.ListUsersForReminder(ctx, deal.ID)
		if err != nil {
			continue
		}

		for _, user := range users {
			if !user.PushToken.Valid || user.PushToken.String == "" {
				continue
			}

			hours := int(time.Until(deal.ExpiresAt.Time).Hours())
			title := "⏰ Deal expires soon!"
			body := deal.OfferName + " - " + deal.PartnerName
			if hours > 0 {
				title = fmt.Sprintf("⏰ Deal expires in %d hours!", hours)
			}

			data := map[string]interface{}{
				"triggered_event_id": deal.ID,
				"event_id":           deal.EventID,
			}

			_, err := s.notifier.Send(user.PushToken.String, title, body, data)
			if err == nil {
				sent++
			}
		}
	}

	s.logger.Info("scheduler: send-reminders complete", "sent", sent, "expiring_deals", len(expiringDeals))
}

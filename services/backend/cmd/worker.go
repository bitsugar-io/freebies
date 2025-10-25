package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/internal/db"
	"github.com/retr0h/freebie/internal/notify"
	"github.com/retr0h/freebie/internal/triggers"

	// Import sources to register them
	_ "github.com/retr0h/freebie/internal/sources/mlb"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker commands for background processing",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run scheduled jobs (designed to be called hourly)",
	Long:  `Checks the current time and runs appropriate jobs. Call this hourly.`,
	RunE:  runScheduledJobs,
}

var checkTriggersCmd = &cobra.Command{
	Use:   "check-triggers",
	Short: "Check game results and create triggered events",
	Long: `Fetches yesterday's game results and checks if any trigger conditions were met.
For each triggered event, creates a record and notifies subscribed users.`,
	RunE: runCheckTriggers,
}

var sendRemindersCmd = &cobra.Command{
	Use:   "send-reminders",
	Short: "Send reminder notifications for expiring deals",
	Long:  `Finds deals expiring within the next 6 hours and sends reminder notifications.`,
	RunE:  runSendReminders,
}

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.AddCommand(runCmd)
	workerCmd.AddCommand(checkTriggersCmd)
	workerCmd.AddCommand(sendRemindersCmd)
}

var checkDate string

func init() {
	checkTriggersCmd.Flags().StringVar(&checkDate, "date", "", "Check triggers for a specific date (YYYY-MM-DD), defaults to yesterday")
}

func runScheduledJobs(cmd *cobra.Command, args []string) error {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	now := time.Now().In(loc)
	hour := now.Hour()

	logger := slog.Default()
	logger.Info("worker run", "hour_pt", hour, "time_pt", now.Format(time.RFC3339))

	// 6am PT - check yesterday's games
	if hour == 6 {
		logger.Info("running check-triggers (6am PT)")
		return runCheckTriggers(cmd, args)
	}

	// 6pm PT - send reminders for expiring deals
	if hour == 18 {
		logger.Info("running send-reminders (6pm PT)")
		return runSendReminders(cmd, args)
	}

	logger.Info("no scheduled jobs for this hour")
	return nil
}

func runCheckTriggers(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := slog.Default()

	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	queries := db.New(database)
	checker := triggers.NewChecker(queries)
	notifier := notify.NewExpoNotifier()

	// Determine which date to check
	var results []triggers.CheckResult
	if checkDate != "" {
		date, err := time.Parse("2006-01-02", checkDate)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
		logger.Info("checking triggers for date", "date", checkDate)
		results, err = checker.CheckAllForDate(ctx, date)
		if err != nil {
			return fmt.Errorf("checking triggers: %w", err)
		}
	} else {
		logger.Info("checking triggers for yesterday")
		results, err = checker.CheckAll(ctx)
		if err != nil {
			return fmt.Errorf("checking triggers: %w", err)
		}
	}

	triggered := 0
	notified := 0
	for _, result := range results {
		if result.Error != nil {
			logger.Warn("check failed",
				"event_id", result.EventID,
				"team", result.TeamID,
				"error", result.Error,
			)
			continue
		}

		if result.Stats == nil {
			logger.Info("no game found",
				"event_id", result.EventID,
				"team", result.TeamID,
				"rule", fmt.Sprintf("%s %s %d", result.Rule.Metric, result.Rule.Operator, result.Rule.Value),
			)
			continue
		}

		metricValue := result.Stats.Metrics[result.Rule.Metric]
		if result.Triggered {
			triggered++
			logger.Info("TRIGGERED",
				"event_id", result.EventID,
				"team", result.TeamID,
				"opponent", result.Stats.Opponent,
				"metric", result.Rule.Metric,
				"value", metricValue,
				"required", fmt.Sprintf("%s %d", result.Rule.Operator, result.Rule.Value),
				"game", result.Stats.GameID,
			)

			// Send notifications if this is a new triggered event
			if result.TriggeredEventID != "" {
				sent := notifySubscribers(ctx, logger, queries, notifier, result)
				notified += sent
			}
		} else {
			logger.Info("not triggered",
				"event_id", result.EventID,
				"team", result.TeamID,
				"opponent", result.Stats.Opponent,
				"metric", result.Rule.Metric,
				"value", metricValue,
				"required", fmt.Sprintf("%s %d", result.Rule.Operator, result.Rule.Value),
			)
		}
	}

	logger.Info("check-triggers complete", "triggered", triggered, "notified", notified, "total_events", len(results))
	return nil
}

func notifySubscribers(ctx context.Context, logger *slog.Logger, queries *db.Queries, notifier *notify.ExpoNotifier, result triggers.CheckResult) int {
	// Get all subscribers for this event
	subscribers, err := queries.ListEventSubscribers(ctx, result.EventID)
	if err != nil {
		logger.Error("failed to list subscribers", "event_id", result.EventID, "error", err)
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

		_, err := notifier.Send(sub.PushToken.String, title, body, data)
		if err != nil {
			logger.Error("failed to send notification", "user_id", sub.UserID, "error", err)
			continue
		}
		sent++
	}

	if sent > 0 {
		logger.Info("notified subscribers", "event_id", result.EventID, "count", sent)
	}

	return sent
}

func runSendReminders(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := slog.Default()

	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	queries := db.New(database)

	// Find deals expiring in the next 6 hours
	expiringDeals, err := queries.ListExpiringTriggeredEvents(ctx, sql.NullString{String: "6", Valid: true})
	if err != nil {
		return fmt.Errorf("failed to list expiring deals: %w", err)
	}

	logger.Info("checking for expiring deals", "count", len(expiringDeals))

	notifier := notify.NewExpoNotifier()
	sent := 0

	for _, deal := range expiringDeals {
		// Get users who should receive reminders for this deal
		users, err := queries.ListUsersForReminder(ctx, deal.ID)
		if err != nil {
			logger.Error("failed to list users for reminder", "deal_id", deal.ID, "error", err)
			continue
		}

		for _, user := range users {
			if !user.PushToken.Valid || user.PushToken.String == "" {
				continue
			}

			// Calculate time remaining
			timeRemaining := time.Until(deal.ExpiresAt.Time)
			hours := int(timeRemaining.Hours())

			title := fmt.Sprintf("⏰ Deal expires in %d hours!", hours)
			body := fmt.Sprintf("Don't forget: %s - %s", deal.OfferName, deal.PartnerName)
			data := map[string]interface{}{
				"triggered_event_id": deal.ID,
				"event_id":           deal.EventID,
			}

			_, err := notifier.Send(user.PushToken.String, title, body, data)
			if err != nil {
				logger.Error("failed to send reminder", "user_id", user.ID, "error", err)
				continue
			}

			sent++
		}
	}

	logger.Info("send-reminders complete", "sent", sent, "expiring_deals", len(expiringDeals))
	return nil
}

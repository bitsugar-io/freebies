package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/internal/db"
	"github.com/retr0h/freebie/internal/notify"
	"github.com/retr0h/freebie/internal/triggers"
	"github.com/retr0h/freebie/internal/worker"

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

func newWorkerService() (*worker.Service, func(), error) {
	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	queries := db.New(database)
	checker := triggers.NewChecker(queries)
	notifier := notify.NewExpoNotifier()
	svc := worker.NewService(queries, checker, notifier, slog.Default())

	cleanup := func() { database.Close() }
	return svc, cleanup, nil
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
	svc, cleanup, err := newWorkerService()
	if err != nil {
		return err
	}
	defer cleanup()

	var date time.Time
	if checkDate != "" {
		date, err = time.Parse("2006-01-02", checkDate)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
	}

	_, err = svc.CheckTriggers(context.Background(), date)
	return err
}

func runSendReminders(cmd *cobra.Command, args []string) error {
	svc, cleanup, err := newWorkerService()
	if err != nil {
		return err
	}
	defer cleanup()

	_, err = svc.SendReminders(context.Background())
	return err
}

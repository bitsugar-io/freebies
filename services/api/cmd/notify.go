package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Send push notifications",
	Long:  `Send test push notifications to users.`,
}

var notifyTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a test notification to all users with push tokens",
	RunE:  runNotifyTest,
}

var notifySendCmd = &cobra.Command{
	Use:   "send [user-id]",
	Short: "Send a notification to a specific user",
	Args:  cobra.ExactArgs(1),
	RunE:  runNotifySend,
}

func init() {
	rootCmd.AddCommand(notifyCmd)
	notifyCmd.AddCommand(notifyTestCmd)
	notifyCmd.AddCommand(notifySendCmd)

	notifyTestCmd.Flags().String("title", "🎉 Test Notification", "Notification title")
	notifyTestCmd.Flags().String("body", "This is a test notification from Freebies!", "Notification body")

	notifySendCmd.Flags().String("title", "🎉 Test Notification", "Notification title")
	notifySendCmd.Flags().String("body", "This is a test notification from Freebies!", "Notification body")
	notifySendCmd.Flags().String("event-id", "", "Event ID for deep linking (opens deal on tap)")
	notifySendCmd.Flags().String("triggered-event-id", "", "Triggered event ID for deep linking")
}

func runNotifyTest(cmd *cobra.Command, args []string) error {
	title, _ := cmd.Flags().GetString("title")
	body, _ := cmd.Flags().GetString("body")

	logger.Info("config", "database", cfg.Database.Path)

	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	defer database.Close()

	queries := db.New(database)
	ctx := context.Background()

	users, err := queries.ListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}

	notifier := notify.NewExpoNotifier()
	sent := 0
	skipped := 0

	for _, user := range users {
		if !user.PushToken.Valid || user.PushToken.String == "" {
			skipped++
			continue
		}

		token := user.PushToken.String
		if !notify.IsValidExpoToken(token) {
			logger.Warn("invalid token", "user", user.ID, "token", token[:20]+"...")
			skipped++
			continue
		}

		ticket, err := notifier.Send(cmd.Context(), token, title, body, map[string]interface{}{
			"type": "test",
		})
		if err != nil {
			logger.Error("failed to send", "user", user.ID, "error", err)
			continue
		}

		if ticket.Status == "ok" {
			logger.Info("sent notification", "user", user.ID, "ticket", ticket.ID)
			sent++
		} else {
			logger.Error("notification failed", "user", user.ID, "error", ticket.Message)
		}
	}

	logger.Info("notifications complete", "sent", sent, "skipped", skipped)
	return nil
}

func runNotifySend(cmd *cobra.Command, args []string) error {
	userID := args[0]
	title, _ := cmd.Flags().GetString("title")
	body, _ := cmd.Flags().GetString("body")
	eventID, _ := cmd.Flags().GetString("event-id")
	triggeredEventID, _ := cmd.Flags().GetString("triggered-event-id")

	logger.Info("config", "database", cfg.Database.Path, "user", userID)

	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	defer database.Close()

	queries := db.New(database)
	ctx := context.Background()

	user, err := queries.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.PushToken.Valid || user.PushToken.String == "" {
		return fmt.Errorf("user has no push token")
	}

	token := user.PushToken.String
	if !notify.IsValidExpoToken(token) {
		return fmt.Errorf("invalid expo token: %s", token[:20]+"...")
	}

	data := map[string]interface{}{
		"type": "test",
	}
	if eventID != "" {
		data["eventId"] = eventID
	}
	if triggeredEventID != "" {
		data["triggeredEventId"] = triggeredEventID
	}

	notifier := notify.NewExpoNotifier()
	ticket, err := notifier.Send(cmd.Context(), token, title, body, data)
	if err != nil {
		return fmt.Errorf("sending notification: %w", err)
	}

	if ticket.Status == "ok" {
		logger.Info("notification sent", "ticket", ticket.ID)
	} else {
		logger.Error("notification failed", "error", ticket.Message)
	}

	return nil
}

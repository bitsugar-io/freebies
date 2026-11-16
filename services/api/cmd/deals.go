package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

var dealsCmd = &cobra.Command{
	Use:   "deals",
	Short: "Manage active deals (triggered events)",
	Long:  `Create and manage test deals for development and testing.`,
}

var dealsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a test deal for a random event",
	RunE:  runDealsCreate,
}

var dealsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active deals",
	RunE:  runDealsList,
}

var dealsTriggerCmd = &cobra.Command{
	Use:   "trigger [event-id]",
	Short: "Trigger a specific event (creates deal and notifies subscribers)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDealsTrigger,
}

func init() {
	rootCmd.AddCommand(dealsCmd)
	dealsCmd.AddCommand(dealsCreateCmd)
	dealsCmd.AddCommand(dealsListCmd)
	dealsCmd.AddCommand(dealsTriggerCmd)

	dealsCreateCmd.Flags().Int("hours", 24, "Hours until deal expires")
	dealsTriggerCmd.Flags().Int("hours", 24, "Hours until deal expires")
	dealsTriggerCmd.Flags().Int("unlock-after-days", 0,
		"Delay the code: unlocks at midnight Pacific this many days after yesterday's game (rule unlock_after_days)")
	dealsTriggerCmd.Flags().Int("valid-days", 0,
		"How many calendar days the deal lasts (rule valid_days); replaces --hours. Use --notify=false to avoid a push")
	dealsTriggerCmd.Flags().Bool("notify", true, "Send push notifications to subscribers")
}

func runDealsCreate(cmd *cobra.Command, args []string) error {
	hours, _ := cmd.Flags().GetInt("hours")
	logger.Info("config", "database", cfg.Database.Path, "hours", hours)

	database, err := db.Open(string(cfg.Database.Path))
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	queries := db.New(database)
	ctx := context.Background()

	// Get all active events
	events, err := queries.ListActiveEvents(ctx)
	if err != nil {
		return fmt.Errorf("listing events: %w", err)
	}

	if len(events) == 0 {
		logger.Info("no active events found - run 'seed' first")
		return nil
	}

	// Pick a random event
	event := events[rand.Intn(len(events))]

	// Auto-subscribe all users to this event (for testing)
	users, err := queries.ListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}
	var subscribedUsers []string
	for _, user := range users {
		// Check if subscription already exists
		_, err := queries.GetSubscription(ctx, db.GetSubscriptionParams{
			UserID:  user.ID,
			EventID: event.ID,
		})
		if err == sql.ErrNoRows {
			// Create subscription
			_, err = queries.CreateSubscription(ctx, db.CreateSubscriptionParams{
				ID:      uuid.New().String(),
				UserID:  user.ID,
				EventID: event.ID,
			})
			if err != nil {
				logger.Warn("failed to create subscription", "user", user.ID, "error", err)
			} else {
				subscribedUsers = append(subscribedUsers, user.ID)
			}
		}
	}
	if len(subscribedUsers) > 0 {
		logger.Info("auto-subscribed users", "users", subscribedUsers, "event", event.ID)
	}

	// Create the triggered event with expiration
	expiresAt := time.Now().UTC().Add(time.Duration(hours) * time.Hour)

	triggered, err := queries.CreateTriggeredEvent(ctx, db.CreateTriggeredEventParams{
		ID:        uuid.New().String(),
		EventID:   event.ID,
		GameID:    sql.NullString{String: fmt.Sprintf("game-%d", time.Now().Unix()), Valid: true},
		ExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
		Payload:   sql.NullString{String: `{"source": "cli-test"}`, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("creating triggered event: %w", err)
	}

	logger.Info("created test deal",
		"id", triggered.ID,
		"event", event.OfferName,
		"team", event.TeamName,
		"expires", expiresAt.Format("2006-01-02 15:04:05"),
	)

	return nil
}

func runDealsList(cmd *cobra.Command, args []string) error {
	logger.Info("config", "database", cfg.Database.Path)

	database, err := db.Open(string(cfg.Database.Path))
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	queries := db.New(database)
	ctx := context.Background()

	deals, err := queries.ListActiveTriggeredEvents(ctx)
	if err != nil {
		return fmt.Errorf("listing active deals: %w", err)
	}

	if len(deals) == 0 {
		logger.Info("no active deals")
		return nil
	}

	fmt.Println("\nActive Deals:")
	fmt.Println("─────────────")
	for _, deal := range deals {
		expiresIn := "no expiration"
		if deal.ExpiresAt.Valid {
			remaining := time.Until(deal.ExpiresAt.Time)
			hours := int(remaining.Hours())
			minutes := int(remaining.Minutes()) % 60
			expiresIn = fmt.Sprintf("%dh %dm remaining", hours, minutes)
		}

		// Get subscribers for this event
		subscribers, _ := queries.ListEventSubscribers(ctx, deal.EventID)
		var userIDs []string
		for _, s := range subscribers {
			userIDs = append(userIDs, s.UserID)
		}

		fmt.Printf("\n  %s\n", deal.OfferName)
		fmt.Printf("  Team: %s | Partner: %s\n", deal.TeamName, deal.PartnerName)
		fmt.Printf("  Trigger: %s\n", deal.TriggerCondition)
		fmt.Printf("  Expires: %s\n", expiresIn)
		if len(userIDs) > 0 {
			fmt.Printf("  Subscribers: %v\n", userIDs)
		} else {
			fmt.Printf("  Subscribers: none\n")
		}
		fmt.Printf("  Event ID: %s\n", deal.EventID)
		fmt.Printf("  Triggered Event ID: %s\n", deal.ID)
	}
	fmt.Println()

	return nil
}

func runDealsTrigger(cmd *cobra.Command, args []string) error {
	eventID := args[0]
	hours, _ := cmd.Flags().GetInt("hours")
	unlockDays, _ := cmd.Flags().GetInt("unlock-after-days")
	validDays, _ := cmd.Flags().GetInt("valid-days")
	shouldNotify, _ := cmd.Flags().GetBool("notify")
	if unlockDays < 0 || validDays < 0 {
		return fmt.Errorf("--unlock-after-days and --valid-days must not be negative")
	}
	delayed := unlockDays > 0 || validDays > 0
	if delayed && cmd.Flags().Changed("hours") {
		return fmt.Errorf("--hours cannot be combined with --unlock-after-days or --valid-days")
	}
	logger.Info("config", "database", cfg.Database.Path, "eventId", eventID, "hours", hours,
		"unlockAfterDays", unlockDays, "validDays", validDays, "notify", shouldNotify)

	database, err := db.Open(string(cfg.Database.Path))
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	queries := db.New(database)
	ctx := context.Background()

	// Get the event
	event, err := queries.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	// Create the triggered event. With --unlock-after-days / --valid-days the
	// window comes from the same code the checker uses, as if the game were
	// yesterday's; otherwise it is --hours from now.
	expiresAt := time.Now().UTC().Add(time.Duration(hours) * time.Hour)
	var availableAt sql.NullTime
	if delayed {
		pt, err := time.LoadLocation("America/Los_Angeles")
		if err != nil {
			return fmt.Errorf("loading timezone: %w", err)
		}
		avail, exp := triggers.Window(time.Now().In(pt).AddDate(0, 0, -1),
			triggers.Rule{UnlockAfterDays: unlockDays, ValidDays: validDays})
		expiresAt = exp
		if !avail.IsZero() {
			availableAt = sql.NullTime{Time: avail, Valid: true}
		}
	}

	triggered, err := queries.CreateTriggeredEvent(ctx, db.CreateTriggeredEventParams{
		ID:          uuid.New().String(),
		EventID:     event.ID,
		GameID:      sql.NullString{String: fmt.Sprintf("game-%d", time.Now().Unix()), Valid: true},
		ExpiresAt:   sql.NullTime{Time: expiresAt, Valid: true},
		AvailableAt: availableAt,
		Payload:     sql.NullString{String: `{"source": "cli-trigger"}`, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("creating triggered event: %w", err)
	}

	attrs := []any{"id", triggered.ID, "event", event.OfferName,
		"expires", expiresAt.Format("2006-01-02 15:04:05")}
	if availableAt.Valid {
		attrs = append(attrs, "available", availableAt.Time.Format("2006-01-02 15:04:05"))
	}
	logger.Info("deal created", attrs...)

	if !shouldNotify {
		return nil
	}

	// Get subscribers and notify them
	subscribers, err := queries.ListEventSubscribers(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("listing subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		logger.Info("no subscribers for this event")
		return nil
	}

	notifier := notify.NewExpoNotifier()
	sent := 0
	failed := 0

	title := fmt.Sprintf("🎉 %s!", event.OfferName)
	expiresText := fmt.Sprintf("Expires in %d hours", hours)
	if delayed {
		pt, _ := time.LoadLocation("America/Los_Angeles")
		expiresText = "Expires " + expiresAt.In(pt).Format("Mon Jan 2 3:04pm")
		if availableAt.Valid {
			expiresText = "Code unlocks " + availableAt.Time.In(pt).Format("Mon Jan 2") + ". " + expiresText
		}
	}
	body := fmt.Sprintf("%s just triggered! %s. %s", event.TriggerCondition, event.OfferDescription, expiresText)

	for _, sub := range subscribers {
		if !sub.PushToken.Valid || sub.PushToken.String == "" {
			continue
		}

		token := sub.PushToken.String
		if !notify.IsValidExpoToken(token) {
			continue
		}

		ticket, err := notifier.Send(cmd.Context(), token, title, body, map[string]interface{}{
			"type":             "freebie",
			"eventId":          event.ID,
			"triggeredEventId": triggered.ID,
			"expiresAt":        expiresAt.Format(time.RFC3339),
		})
		if err != nil {
			logger.Error("failed to send", "user", sub.UserID, "error", err)
			failed++
			continue
		}

		if ticket.Status == "ok" {
			logger.Info("notified subscriber", "user", sub.UserID)
			sent++
		} else {
			failed++
		}
	}

	logger.Info("notifications sent", "sent", sent, "failed", failed)
	return nil
}

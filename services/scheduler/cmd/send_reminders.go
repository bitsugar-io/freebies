package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/scheduler/internal/client/gen"
)

var sendRemindersCmd = &cobra.Command{
	Use:   "send-reminders",
	Short: "Send reminder notifications for expiring deals",
	RunE:  runSendReminders,
}

func runSendReminders(cmd *cobra.Command, _ []string) error {
	client, err := gen.NewClient(cfg.API.URL)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	logger.Info("sending reminders", "url", cfg.API.URL)

	resp, err := client.SendReminders(cmd.Context(), bearerAuth(cfg.Worker.Secret))
	if err != nil {
		return fmt.Errorf("calling send-reminders: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send-reminders failed (status %d): %s", resp.StatusCode, body)
	}

	var result gen.SendRemindersResult
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	logger.Info("send-reminders completed",
		"sent", result.Sent,
		"failed", result.Failed,
		"expiringDeals", result.ExpiringDeals,
	)

	return nil
}

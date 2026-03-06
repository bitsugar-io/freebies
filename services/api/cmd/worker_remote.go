package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"

	clientgen "github.com/retr0h/freebie/services/api/internal/client/gen"
)

var (
	apiURL       string
	workerSecret string
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Call worker endpoints on a running API server",
	Long:  `Calls the internal worker HTTP endpoints on a remote API server using the generated client.`,
}

var remoteCheckTriggersCmd = &cobra.Command{
	Use:   "check-triggers",
	Short: "Call the check-triggers endpoint on the API server",
	RunE:  runRemoteCheckTriggers,
}

var remoteSendRemindersCmd = &cobra.Command{
	Use:   "send-reminders",
	Short: "Call the send-reminders endpoint on the API server",
	RunE:  runRemoteSendReminders,
}

func init() {
	workerCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(remoteCheckTriggersCmd)
	remoteCmd.AddCommand(remoteSendRemindersCmd)

	remoteCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://freebie-api:8080", "API server URL")
	remoteCmd.PersistentFlags().StringVar(&workerSecret, "worker-secret", "", "Worker bearer token (or set FREEBIE_WORKER_SECRET)")
}

func bearerAuth(token string) clientgen.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func getWorkerSecret() string {
	if workerSecret != "" {
		return workerSecret
	}
	return cfg.Worker.Secret
}

func runRemoteCheckTriggers(cmd *cobra.Command, args []string) error {
	logger := slog.Default()

	secret := getWorkerSecret()
	if secret == "" {
		return fmt.Errorf("worker secret is required (--worker-secret or FREEBIE_WORKER_SECRET)")
	}

	client, err := clientgen.NewClient(apiURL)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	logger.Info("calling check-triggers", "url", apiURL)
	resp, err := client.CheckTriggers(context.Background(), nil, bearerAuth(secret))
	if err != nil {
		return fmt.Errorf("calling check-triggers: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("check-triggers failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var result clientgen.CheckTriggersResult
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	logger.Info("check-triggers complete",
		"triggered", result.Triggered,
		"notified", result.Notified,
		"totalEvents", result.TotalEvents,
	)
	return nil
}

func runRemoteSendReminders(cmd *cobra.Command, args []string) error {
	logger := slog.Default()

	secret := getWorkerSecret()
	if secret == "" {
		return fmt.Errorf("worker secret is required (--worker-secret or FREEBIE_WORKER_SECRET)")
	}

	client, err := clientgen.NewClient(apiURL)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	logger.Info("calling send-reminders", "url", apiURL)
	resp, err := client.SendReminders(context.Background(), bearerAuth(secret))
	if err != nil {
		return fmt.Errorf("calling send-reminders: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send-reminders failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var result clientgen.SendRemindersResult
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	logger.Info("send-reminders complete",
		"sent", result.Sent,
		"failed", result.Failed,
		"expiringDeals", result.ExpiringDeals,
	)
	return nil
}

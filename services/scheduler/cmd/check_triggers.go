package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/scheduler/internal/client/gen"
)

var checkTriggersDate string

var checkTriggersCmd = &cobra.Command{
	Use:   "check-triggers",
	Short: "Check game results and create triggered events",
	RunE:  runCheckTriggers,
}

func init() {
	checkTriggersCmd.Flags().StringVar(
		&checkTriggersDate, "date", "",
		"Check triggers for a specific date (YYYY-MM-DD), defaults to yesterday",
	)
}

// bearerAuth returns a RequestEditorFn that adds bearer token authorization.
func bearerAuth(token string) gen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func runCheckTriggers(cmd *cobra.Command, _ []string) error {
	client, err := gen.NewClient(cfg.API.URL)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	params := &gen.CheckTriggersParams{}
	if checkTriggersDate != "" {
		t, err := time.Parse("2006-01-02", checkTriggersDate)
		if err != nil {
			return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
		}
		d := openapi_types.Date{Time: t}
		params.Date = &d
	}

	logger.Info("checking triggers", "url", cfg.API.URL, "date", checkTriggersDate)

	resp, err := client.CheckTriggers(cmd.Context(), params, bearerAuth(cfg.Worker.Secret))
	if err != nil {
		return fmt.Errorf("calling check-triggers: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("check-triggers failed (status %d): %s", resp.StatusCode, body)
	}

	var result gen.CheckTriggersResult
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	logger.Info("check-triggers completed",
		"triggered", result.Triggered,
		"notified", result.Notified,
		"totalEvents", result.TotalEvents,
	)

	return nil
}

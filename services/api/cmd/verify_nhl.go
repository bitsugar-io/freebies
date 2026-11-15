package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/api/internal/sources"
	_ "github.com/retr0h/freebie/services/api/internal/sources/nhl" // registers
)

var (
	verifyNHLTeam string
	verifyNHLDate string
)

// verifyNHLCmd runs Source.GetGameByDate against the LIVE NHL web API for the
// chosen team/date. No DB, no HTTP server, no notifications.
var verifyNHLCmd = &cobra.Command{
	Use:   "verify-nhl",
	Short: "Sanity-check the NHL source against the live API",
	RunE: func(cmd *cobra.Command, _ []string) error {
		date := time.Now().AddDate(0, 0, -1) // yesterday in local TZ
		if verifyNHLDate != "" {
			t, err := time.Parse("2006-01-02", verifyNHLDate)
			if err != nil {
				return fmt.Errorf("invalid --date (want YYYY-MM-DD): %w", err)
			}
			date = t
		}

		src, err := sources.Get("nhl")
		if err != nil {
			return fmt.Errorf("get nhl source: %w", err)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		stats, err := src.GetGameByDate(ctx, verifyNHLTeam, date)
		if err != nil {
			return fmt.Errorf("GetGameByDate: %w", err)
		}
		if stats == nil {
			fmt.Printf("no final game for %s on %s\n", verifyNHLTeam, date.Format("2006-01-02"))
			return nil
		}

		out, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	verifyNHLCmd.Flags().StringVar(&verifyNHLTeam, "team", "ANA", "Team code (ANA, LAK, ...)")
	verifyNHLCmd.Flags().StringVar(&verifyNHLDate, "date", "", "Date YYYY-MM-DD (default: yesterday)")
	rootCmd.AddCommand(verifyNHLCmd)
}

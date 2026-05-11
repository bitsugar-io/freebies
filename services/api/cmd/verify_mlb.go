package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/api/internal/sources"
	_ "github.com/retr0h/freebie/services/api/internal/sources/mlb" // registers
)

var (
	verifyMLBTeam string
	verifyMLBDate string
)

// verifyMLBCmd runs Source.GetGameByDate against the LIVE MLB Stats API for
// the chosen team/date, exercising the exact code path production uses
// (mlb-sdk → Source → metrics map). No DB. No HTTP server. No notifications.
// Lets us confirm the migration to mlb-sdk produces sane data before
// merging.
var verifyMLBCmd = &cobra.Command{
	Use:   "verify-mlb",
	Short: "Sanity-check the MLB source against the live API",
	RunE: func(cmd *cobra.Command, _ []string) error {
		date := time.Now().AddDate(0, 0, -1) // yesterday in local TZ
		if verifyMLBDate != "" {
			t, err := time.Parse("2006-01-02", verifyMLBDate)
			if err != nil {
				return fmt.Errorf("invalid --date (want YYYY-MM-DD): %w", err)
			}
			date = t
		}

		src, err := sources.Get("mlb")
		if err != nil {
			return fmt.Errorf("get mlb source: %w", err)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		stats, err := src.GetGameByDate(ctx, verifyMLBTeam, date)
		if err != nil {
			return fmt.Errorf("GetGameByDate: %w", err)
		}
		if stats == nil {
			fmt.Printf("no Final game for %s on %s\n", verifyMLBTeam, date.Format("2006-01-02"))
			return nil
		}

		out, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	verifyMLBCmd.Flags().StringVar(&verifyMLBTeam, "team", "LAD", "Team code (LAD, SF, NYY, ...)")
	verifyMLBCmd.Flags().StringVar(&verifyMLBDate, "date", "", "Date YYYY-MM-DD (default: yesterday)")
	rootCmd.AddCommand(verifyMLBCmd)
}

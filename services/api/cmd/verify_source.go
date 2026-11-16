package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

var (
	verifySourceLeague string
	verifySourceTeam   string
	verifySourceDate   string
)

// verifySourceCmd runs Source.GetGameByDate for any registered league against
// the LIVE provider. No DB, no HTTP server, no notifications. verify-mlb and
// verify-nhl predate it and do the same for those two leagues.
var verifySourceCmd = &cobra.Command{
	Use:   "verify-source",
	Short: "Sanity-check a league's source against its live data provider",
	Example: `  go run . verify-source --league mls --team LAFC --date 2026-09-09
  go run . verify-source --league nfl --team LAR --date 2026-11-08`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if verifySourceLeague == "" || verifySourceTeam == "" {
			return fmt.Errorf("--league and --team are required")
		}
		date := time.Now().AddDate(0, 0, -1) // yesterday in local TZ
		if verifySourceDate != "" {
			t, err := time.Parse("2006-01-02", verifySourceDate)
			if err != nil {
				return fmt.Errorf("invalid --date (want YYYY-MM-DD): %w", err)
			}
			date = t
		}

		src, err := sources.Get(verifySourceLeague)
		if err != nil {
			return fmt.Errorf("get %s source: %w (registered: %v)",
				verifySourceLeague, err, sources.List())
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		stats, err := src.GetGameByDate(ctx, verifySourceTeam, date)
		if err != nil {
			return fmt.Errorf("GetGameByDate: %w", err)
		}
		if stats == nil {
			fmt.Printf("no final game for %s %s on %s\n",
				verifySourceLeague, verifySourceTeam, date.Format("2006-01-02"))
			return nil
		}
		out, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	verifySourceCmd.Flags().StringVar(&verifySourceLeague, "league", "", "League id (nba, wnba, nfl, mls, nwsl, ncaa, mlb, nhl)")
	verifySourceCmd.Flags().StringVar(&verifySourceTeam, "team", "", "Team code (LAL, LAFC, LAR, ...)")
	verifySourceCmd.Flags().StringVar(&verifySourceDate, "date", "", "Pacific date YYYY-MM-DD (default: yesterday)")
	rootCmd.AddCommand(verifySourceCmd)
}

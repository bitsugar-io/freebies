package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/api/internal/db"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  `List and manage users in the database.`,
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE:  runUsersList,
}

var usersCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove orphaned users with no push token",
	RunE:  runUsersCleanup,
}

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersCleanupCmd)
}

func runUsersList(cmd *cobra.Command, args []string) error {
	logger.Info("config", "database", cfg.Database.Path)

	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	defer database.Close()

	queries := db.New(database)
	ctx := context.Background()

	// Get all users
	rows, err := database.QueryContext(ctx, "SELECT id, device_id, push_token, platform, token, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDEVICE\tPLATFORM\tAUTH TOKEN\tPUSH TOKEN\tCREATED")
	fmt.Fprintln(w, "--\t------\t--------\t----------\t----------\t-------")

	count := 0
	for rows.Next() {
		var id, deviceID, platform, createdAt string
		var pushToken, authToken *string

		if err := rows.Scan(&id, &deviceID, &pushToken, &platform, &authToken, &createdAt); err != nil {
			return err
		}

		authDisplay := "(none)"
		if authToken != nil && *authToken != "" {
			// Show first 8 chars + "..." for auth token
			if len(*authToken) > 8 {
				authDisplay = (*authToken)[:8] + "..."
			} else {
				authDisplay = *authToken
			}
		}

		pushDisplay := "(none)"
		if pushToken != nil && *pushToken != "" {
			if len(*pushToken) > 20 {
				pushDisplay = (*pushToken)[:20] + "..."
			} else {
				pushDisplay = *pushToken
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			id,
			deviceID,
			platform,
			authDisplay,
			pushDisplay,
			createdAt[:19],
		)
		count++
	}

	w.Flush()

	// Count subscriptions per user
	subCounts, err := queries.CountUserSubscriptions(ctx, "")
	_ = subCounts // TODO: show subscription counts

	fmt.Printf("\nTotal: %d users\n", count)
	return nil
}

func runUsersCleanup(cmd *cobra.Command, args []string) error {
	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	defer database.Close()

	ctx := context.Background()

	// Count orphaned users first
	var count int
	err = database.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE push_token IS NULL").Scan(&count)
	if err != nil {
		return fmt.Errorf("counting orphaned users: %w", err)
	}

	if count == 0 {
		fmt.Println("No orphaned users found.")
		return nil
	}

	fmt.Printf("Found %d orphaned users (no push token). Cleaning up...\n", count)

	// Delete related data, then users
	for _, query := range []string{
		"DELETE FROM notifications WHERE user_id IN (SELECT id FROM users WHERE push_token IS NULL)",
		"DELETE FROM dismissals WHERE user_id IN (SELECT id FROM users WHERE push_token IS NULL)",
		"DELETE FROM subscriptions WHERE user_id IN (SELECT id FROM users WHERE push_token IS NULL)",
		"DELETE FROM users WHERE push_token IS NULL",
	} {
		if _, err := database.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}
	}

	fmt.Printf("Removed %d orphaned users and their associated data.\n", count)
	return nil
}

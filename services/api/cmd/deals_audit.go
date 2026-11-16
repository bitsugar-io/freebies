package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

var dealsAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "List every deal and whether it can ever fire",
	Long: `Checks each event in the database against the registered game sources:
the league has a source, the source emits the rule's metric and maps the team.

Statuses: OK, NO_SOURCE, UNKNOWN_METRIC, UNKNOWN_TEAM, BAD_RULE, INACTIVE.
Exits non-zero if any ACTIVE event is not OK. Run it before shipping a deal.`,
	SilenceUsage: true,
	RunE:         runDealsAudit,
}

func init() {
	dealsCmd.AddCommand(dealsAuditCmd)
}

func runDealsAudit(cmd *cobra.Command, _ []string) error {
	database, err := db.Open(string(cfg.Database.Path))
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()
	// Audit what a server would have, not a stale schema.
	if err := db.Migrate(database); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}

	events, err := db.New(database).ListEvents(context.Background())
	if err != nil {
		return fmt.Errorf("listing events: %w", err)
	}

	rows := triggers.Audit(events)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "STATUS\tACTIVE\tEVENT\tLEAGUE\tTEAM\tMETRIC\tDETAIL")
	for _, r := range rows {
		active := "no"
		if r.Active {
			active = "yes"
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Status, active, r.EventID, r.League, r.TeamID, r.Metric, r.Detail)
	}
	_ = tw.Flush()

	bad := triggers.Failures(rows)
	if len(bad) > 0 {
		return fmt.Errorf("%d active deal(s) can never fire", len(bad))
	}
	return nil
}

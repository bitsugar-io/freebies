package triggers_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	_ "github.com/retr0h/freebie/services/api/internal/sources/test"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

// TestCheckerSkipsDealsThatCanNeverFire: an active deal with no source for its
// league is warned about (structured, WARN) and reported as skipped; the run
// itself still succeeds.
func TestCheckerSkipsDealsThatCanNeverFire(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE events SET is_active = 0`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO events (id, offer_id, team_id, team_name, league,
		partner_name, offer_name, offer_description, trigger_condition, trigger_rule, region_code, is_active)
		VALUES ('event-x', 'x', 'LAX', 'X', 'curling', 'P', 'O', 'D', 'C',
		'{"metric":"win","operator":"==","value":1}', 'us-ca-la', 1)`); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	results, skipped, err := triggers.NewChecker(db.New(database), triggers.WithLogger(logger)).
		CheckAllForDate(context.Background(), time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("run must not fail: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results = %d, want the broken deal skipped", len(results))
	}
	if len(skipped) != 1 || skipped[0].EventID != "event-x" || skipped[0].Status != "NO_SOURCE" ||
		skipped[0].Reason == "" {
		t.Errorf("skipped = %+v, want event-x NO_SOURCE with a reason", skipped)
	}
	out := buf.String()
	for _, want := range []string{"level=WARN", "event_id=event-x", "status=NO_SOURCE", "reason="} {
		if !strings.Contains(out, want) {
			t.Errorf("log missing %q:\n%s", want, out)
		}
	}
}

package triggers_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

func smokeChecker(t *testing.T, rule string, clock time.Time) (*triggers.Checker, *bytes.Buffer, func() int) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE events SET is_active = 0 WHERE id != 'event-test-smoke'`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE events SET is_active = 1, trigger_rule = ? WHERE id = 'event-test-smoke'`, rule); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	c := triggers.NewChecker(db.New(database),
		triggers.WithLogger(slog.New(slog.NewTextHandler(&buf, nil))),
		triggers.WithClock(func() time.Time { return clock }))
	count := func() int {
		var n int
		if err := database.QueryRow(`SELECT COUNT(*) FROM triggered_events`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	return c, &buf, count
}

const duckRule = `{"metric":"score","operator":">=","value":1,"redemption_window":"same_day","expires_at_local":"10:30"}`

// A normal 7am run for last night's game records the breakfast deal and
// reports its 10:30am deadline the next morning.
func TestBreakfastDealRecordedOnNormalRun(t *testing.T) {
	pt, _ := time.LoadLocation("America/Los_Angeles")
	c, _, count := smokeChecker(t, duckRule, time.Date(2026, 11, 15, 7, 0, 0, 0, pt))

	results, skipped, err := c.CheckAllForDate(context.Background(), time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 0 || len(results) != 1 || results[0].TriggeredEventID == "" {
		t.Fatalf("results=%+v skipped=%+v, want one recorded deal", results, skipped)
	}
	if want := time.Date(2026, 11, 15, 10, 30, 0, 0, pt); !results[0].ExpiresAt.Equal(want) {
		t.Errorf("expires %s, want %s", results[0].ExpiresAt, want)
	}
	if count() != 1 {
		t.Errorf("triggered_events = %d, want 1", count())
	}
}

// A late or backfill run after the window closed creates nothing (so nothing
// is pushed) and says why.
func TestExpiredDealIsNotRecorded(t *testing.T) {
	pt, _ := time.LoadLocation("America/Los_Angeles")
	c, buf, count := smokeChecker(t, duckRule, time.Date(2026, 11, 15, 12, 0, 0, 0, pt))

	results, skipped, err := c.CheckAllForDate(context.Background(), time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].Expired || results[0].TriggeredEventID != "" {
		t.Fatalf("results=%+v, want one expired result with no triggered event", results)
	}
	if len(skipped) != 1 || skipped[0].Status != triggers.StatusExpired || skipped[0].EventID != "event-test-smoke" {
		t.Errorf("skipped = %+v, want event-test-smoke EXPIRED", skipped)
	}
	if count() != 0 {
		t.Errorf("triggered_events = %d, want 0", count())
	}
	for _, want := range []string{"level=WARN", "event_id=event-test-smoke", "status=EXPIRED", "reason="} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("log missing %q:\n%s", want, buf.String())
		}
	}
}

// Backfilling an old game never yields a deal, whatever the window.
func TestBackfillOfOldGameIsSkipped(t *testing.T) {
	pt, _ := time.LoadLocation("America/Los_Angeles")
	c, _, count := smokeChecker(t,
		`{"metric":"score","operator":">=","value":1,"valid_days":3}`,
		time.Date(2026, 11, 20, 7, 0, 0, 0, pt))
	_, skipped, err := c.CheckAllForDate(context.Background(), time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 || count() != 0 {
		t.Errorf("skipped=%+v count=%d, want skipped and none recorded", skipped, count())
	}
}

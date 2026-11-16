package worker

import (
	"context"
	"database/sql"
	"github.com/retr0h/freebie/services/api/internal/triggers"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
)

// TestRecordNotifications verifies the worker writes one row to the
// notifications table per attempted push, with status matching the Expo
// outcome. Previously the worker called Expo but never wrote to the table,
// leaving the audit log empty.
func TestRecordNotifications(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	// Seed two users + a triggered_event referencing the pre-existing
	// event-test-smoke from migration 011.
	if _, err := database.ExecContext(ctx, `
		INSERT INTO users (id, device_id, platform) VALUES
			('user-1','dev-1','ios'),
			('user-2','dev-2','ios');
		INSERT INTO triggered_events (id, event_id) VALUES
			('te-1','event-test-smoke');
	`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	svc := &Service{
		queries: db.New(database),
		logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	messages := []notify.ExpoPushMessage{
		{To: "tok-1"},
		{To: "tok-2"},
	}
	userByToken := map[string]string{
		"tok-1": "user-1",
		"tok-2": "user-2",
	}

	svc.recordNotifications(ctx, messages, KindTrigger, "te-1", userByToken, &notify.BatchResult{Sent: 1, Failed: 1})

	var sent, failed, total int
	if err := database.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM notifications WHERE status='sent'`).Scan(&sent); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM notifications WHERE status='failed'`).Scan(&failed); err != nil {
		t.Fatal(err)
	}

	if total != 2 {
		t.Errorf("notifications total = %d, want 2", total)
	}
	if sent != 1 {
		t.Errorf("notifications sent = %d, want 1", sent)
	}
	if failed != 1 {
		t.Errorf("notifications failed = %d, want 1", failed)
	}
}

func TestRecordNotificationsAllSent(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	ctx := context.Background()
	if _, err := database.ExecContext(ctx, `
		INSERT INTO users (id, device_id, platform) VALUES
			('user-1','dev-1','ios'),
			('user-2','dev-2','ios'),
			('user-3','dev-3','ios');
		INSERT INTO triggered_events (id, event_id) VALUES
			('te-1','event-test-smoke');
	`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	svc := &Service{
		queries: db.New(database),
		logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	messages := []notify.ExpoPushMessage{{To: "t1"}, {To: "t2"}, {To: "t3"}}
	userByToken := map[string]string{"t1": "user-1", "t2": "user-2", "t3": "user-3"}

	svc.recordNotifications(ctx, messages, KindTrigger, "te-1", userByToken, &notify.BatchResult{Sent: 3, Failed: 0})

	var sent int
	if err := database.QueryRow(`SELECT COUNT(*) FROM notifications WHERE status='sent'`).Scan(&sent); err != nil {
		t.Fatal(err)
	}
	if sent != 3 {
		t.Errorf("all-success: sent count = %d, want 3", sent)
	}
}

func TestRecordNotificationsSkipsUnknownToken(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	ctx := context.Background()
	if _, err := database.ExecContext(ctx, `
		INSERT INTO users (id, device_id, platform) VALUES ('user-1','dev-1','ios');
		INSERT INTO triggered_events (id, event_id) VALUES ('te-1','event-test-smoke');
	`); err != nil {
		t.Fatalf("seed: %v", err)
	}

	svc := &Service{
		queries: db.New(database),
		logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	// One message we know about; one whose token has no user mapping (should be skipped).
	messages := []notify.ExpoPushMessage{{To: "known"}, {To: "stale"}}
	userByToken := map[string]string{"known": "user-1"}

	svc.recordNotifications(ctx, messages, KindTrigger, "te-1", userByToken, &notify.BatchResult{Sent: 2, Failed: 0})

	var total int
	if err := database.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Errorf("stale-token: total = %d, want 1 (skip unknown)", total)
	}
}

func TestAlertsPaused(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name  string
		until sql.NullTime
		want  bool
	}{
		{"not paused", sql.NullTime{}, false},
		{"future", sql.NullTime{Time: now.Add(time.Hour), Valid: true}, true},
		{"expired", sql.NullTime{Time: now.Add(-time.Hour), Valid: true}, false},
		{"until resumed", sql.NullTime{Time: time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC), Valid: true}, true},
	}
	for _, c := range cases {
		if got := alertsPaused(c.until, now); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// Paused users get no push (the notifier is nil, so any send would panic) but
// the triggered event remains recorded.
func TestPausedUsersSkippedForUnlockAndReminder(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := database.ExecContext(ctx, `
		INSERT INTO users (id, device_id, platform, push_token, alerts_paused_until)
			VALUES ('user-p','dev-p','ios','ExponentPushToken[p]','9999-12-31 00:00:00+00:00');
		INSERT INTO subscriptions (id, user_id, event_id) VALUES ('sub-p','user-p','event-test-smoke');
		INSERT INTO triggered_events (id, event_id, expires_at)
			VALUES ('te-p','event-test-smoke', datetime('now','+2 hours'));
	`); err != nil {
		t.Fatal(err)
	}
	svc := &Service{
		queries: db.New(database),
		logger:  slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	res, err := svc.SendReminders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Sent != 0 || res.Failed != 0 {
		t.Errorf("reminders sent=%d failed=%d, want 0", res.Sent, res.Failed)
	}

	ev, err := svc.queries.GetEvent(ctx, "event-test-smoke")
	if err != nil {
		t.Fatal(err)
	}
	n := svc.notifySubscribers(ctx, triggers.CheckResult{
		EventID: "event-test-smoke", Event: &ev, TriggeredEventID: "te-p",
	})
	if n != 0 {
		t.Errorf("unlock notified %d, want 0", n)
	}

	var te, notes int
	_ = database.QueryRow(`SELECT COUNT(*) FROM triggered_events WHERE id='te-p'`).Scan(&te)
	_ = database.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&notes)
	if te != 1 || notes != 0 {
		t.Errorf("triggered=%d notifications=%d, want 1 and 0", te, notes)
	}
}

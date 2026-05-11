package worker

import (
	"context"
	"log/slog"
	"os"
	"testing"

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

	svc.recordNotifications(ctx, messages, "te-1", userByToken, &notify.BatchResult{Sent: 1, Failed: 1})

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

	svc.recordNotifications(ctx, messages, "te-1", userByToken, &notify.BatchResult{Sent: 3, Failed: 0})

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

	svc.recordNotifications(ctx, messages, "te-1", userByToken, &notify.BatchResult{Sent: 2, Failed: 0})

	var total int
	if err := database.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Errorf("stale-token: total = %d, want 1 (skip unknown)", total)
	}
}

package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
)

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

	won, err := svc.NotifyTriggered(ctx, "te-p")
	if err != nil {
		t.Fatal(err)
	}
	if won.Sent != 0 {
		t.Errorf("unlock notified %d, want 0", won.Sent)
	}

	var te, notes int
	_ = database.QueryRow(`SELECT COUNT(*) FROM triggered_events WHERE id='te-p'`).Scan(&te)
	_ = database.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&notes)
	if te != 1 || notes != 0 {
		t.Errorf("triggered=%d notifications=%d, want 1 and 0", te, notes)
	}
}

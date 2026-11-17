package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
)

// Every push a user can tap opens the YOU WON screen for that exact win: the
// app routes data.url ("freebie://deal/<triggered id>") to the Deal screen
// (apps/mobile/src/navigation/launch.test.ts covers the app side).
func TestPushesLinkToTheirDeal(t *testing.T) {
	var mu sync.Mutex
	var sent []notify.ExpoPushMessage
	expo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var batch []notify.ExpoPushMessage
		if err := json.Unmarshal(body, &batch); err != nil {
			t.Errorf("push body is not a message batch: %v", err)
		}
		mu.Lock()
		sent = append(sent, batch...)
		mu.Unlock()
		tickets := make([]map[string]string, len(batch))
		for i := range tickets {
			tickets[i] = map[string]string{"status": "ok", "id": "ticket"}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": tickets})
	}))
	t.Cleanup(expo.Close)

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
		INSERT INTO users (id, device_id, platform, push_token)
			VALUES ('user-1','dev-1','ios','ExponentPushToken[abc]');
		INSERT INTO subscriptions (id, user_id, event_id) VALUES ('sub-1','user-1','event-test-smoke');
		INSERT INTO triggered_events (id, event_id, expires_at)
			VALUES ('te-1','event-test-smoke', datetime('now','+2 hours'));
	`); err != nil {
		t.Fatal(err)
	}
	svc := &Service{
		queries:  db.New(database),
		notifier: notify.NewExpoNotifierWithURL(expo.URL),
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	check := func(kind string) {
		t.Helper()
		mu.Lock()
		defer mu.Unlock()
		if len(sent) == 0 {
			t.Fatalf("%s: no push sent", kind)
		}
		for _, m := range sent {
			if got := m.Data["url"]; got != "freebie://deal/te-1" {
				t.Errorf("%s push data.url = %v, want freebie://deal/te-1", kind, got)
			}
			if got := m.Data["triggeredEventId"]; got != "te-1" {
				t.Errorf("%s push data.triggeredEventId = %v, want te-1", kind, got)
			}
		}
		sent = nil
	}

	won, err := svc.NotifyTriggered(ctx, "te-1")
	if err != nil {
		t.Fatal(err)
	}
	if won.Sent != 1 {
		t.Fatalf("you-won push notified %d users, want 1", won.Sent)
	}
	check("you-won")

	res, err := svc.SendReminders(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Sent != 1 {
		t.Fatalf("reminders sent = %d, want 1 (result %+v)", res.Sent, res)
	}
	check("reminder")
}

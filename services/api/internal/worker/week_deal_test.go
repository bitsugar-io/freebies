package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
)

// A week-long delayed deal (Rams / Carl's Jr.: game Sat, code unlocks Mon,
// good through Sun): reminders go out on the unlock day ("code works now")
// and the last day only, and never to someone who already claimed it.
func TestWeekLongDealReminders(t *testing.T) {
	var mu sync.Mutex
	var sent []notify.ExpoPushMessage
	expo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var batch []notify.ExpoPushMessage
		_ = json.Unmarshal(body, &batch)
		mu.Lock()
		sent = append(sent, batch...)
		mu.Unlock()
		tickets := make([]map[string]string, len(batch))
		for i := range tickets {
			tickets[i] = map[string]string{"status": "ok", "id": "t"}
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
	unlock := time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC)     // Mon 00:00 PT
	expires := time.Date(2026, 11, 23, 7, 59, 59, 0, time.UTC)  // Sun 23:59:59 PT
	triggered := time.Date(2026, 11, 14, 15, 0, 0, 0, time.UTC) // Sat 7am PT
	if _, err := database.Exec(`
		INSERT INTO users (id, device_id, platform, push_token) VALUES
			('fan','d1','ios','ExponentPushToken[fan]'), ('claimer','d2','ios','ExponentPushToken[claimer]');
		INSERT INTO subscriptions (id, user_id, event_id) VALUES
			('s1','fan','event-test-smoke'), ('s2','claimer','event-test-smoke');
		INSERT INTO triggered_events (id, event_id, triggered_at, available_at, expires_at) VALUES (?, 'event-test-smoke', ?, ?, ?);
		INSERT INTO dismissals (id, user_id, triggered_event_id, type) VALUES ('d1','claimer','te-week','got_it');
	`, "te-week", triggered, unlock, expires); err != nil {
		t.Fatal(err)
	}

	pt := pacific
	run := func(at time.Time) []notify.ExpoPushMessage {
		t.Helper()
		mu.Lock()
		sent = nil
		mu.Unlock()
		svc := &Service{
			queries:  db.New(database),
			notifier: notify.NewExpoNotifierWithURL(expo.URL),
			logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
			now:      func() time.Time { return at },
		}
		if _, err := svc.SendReminders(context.Background()); err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		defer mu.Unlock()
		return append([]notify.ExpoPushMessage(nil), sent...)
	}
	at := func(day, hour int) time.Time { return time.Date(2026, 11, day, hour, 0, 0, 0, pt) }

	if got := run(at(15, 11)); len(got) != 0 {
		t.Errorf("Sun before unlock: %d reminders, want 0 (code is locked)", len(got))
	}
	mon := run(at(16, 11))
	if len(mon) != 1 || mon[0].To != "ExponentPushToken[fan]" {
		t.Fatalf("Mon unlock day: got %+v, want one reminder to the fan only", mon)
	}
	if !strings.Contains(mon[0].Title, "code works now") {
		t.Errorf("unlock-day title = %q, want it to say the code works now", mon[0].Title)
	}
	for _, d := range []int{17, 18, 19, 20, 21} {
		if got := run(at(d, 11)); len(got) != 0 {
			t.Errorf("Nov %d mid-week: %d reminders, want 0", d, len(got))
		}
	}
	sun := run(at(22, 11))
	if len(sun) != 1 || sun[0].To != "ExponentPushToken[fan]" {
		t.Fatalf("Sun last day: got %+v, want one reminder to the fan only", sun)
	}
	if strings.Contains(sun[0].Title, "code works now") {
		t.Errorf("last-day title = %q, want the expiry wording", sun[0].Title)
	}
}

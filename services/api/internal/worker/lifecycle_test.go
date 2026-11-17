package worker

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
	_ "github.com/retr0h/freebie/services/api/internal/sources/test"
	"github.com/retr0h/freebie/services/api/internal/triggers"
)

// One win from the morning check to close-out, through the real checker, the
// worker and a fake Expo: who gets the "you won" push, who gets reminded, and
// that claiming, silencing and unfollowing stop the reminders.
func TestWinLifecycle(t *testing.T) {
	expo := newFakeExpo(t)
	expo.ticket = func(m notify.ExpoPushMessage) notify.ExpoPushTicket {
		if m.To == "ExponentPushToken[gone]" {
			return errorTicket(notify.ErrDeviceNotRegistered)
		}
		return notify.ExpoPushTicket{Status: "ok", ID: "t"}
	}
	database := openTestDB(t)
	if _, err := database.Exec(`UPDATE events SET is_active = 0 WHERE id != 'event-test-smoke'`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE events SET is_active = 1, trigger_rule = ? WHERE id = 'event-test-smoke'`,
		`{"metric":"score","operator":">=","value":1,"redemption_window":"same_day"}`); err != nil {
		t.Fatal(err)
	}
	// SQL compares expiry with its own clock, so the run happens now: the
	// same_day deal for yesterday's game is live until 23:59:59 Pacific today.
	clock := time.Now()
	followed := sqlTime(clock.Add(-48 * time.Hour))
	if _, err := database.Exec(`
		INSERT INTO users (id, device_id, platform, push_token, alerts_paused_until) VALUES
			('fan','d1','ios','ExponentPushToken[fan]', NULL),
			('claimer','d2','ios','ExponentPushToken[claimer]', NULL),
			('silencer','d3','ios','ExponentPushToken[silencer]', NULL),
			('leaver','d4','ios','ExponentPushToken[leaver]', NULL),
			('paused','d5','ios','ExponentPushToken[paused]', '9999-12-31 00:00:00+00:00'),
			('notoken','d6','ios', NULL, NULL),
			('uninstalled','d7','ios','ExponentPushToken[gone]', NULL),
			('stranger','d8','ios','ExponentPushToken[stranger]', NULL);
		INSERT INTO subscriptions (id, user_id, event_id, created_at)
			SELECT 's-' || id, id, 'event-test-smoke', ? FROM users WHERE id != 'stranger';`, followed); err != nil {
		t.Fatal(err)
	}
	svc := testService(database, expo.srv.URL, &clock)
	svc.checker = triggers.NewChecker(db.New(database),
		triggers.WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		triggers.WithClock(func() time.Time { return clock }))
	ctx := context.Background()
	pushed := func() []string {
		got := tokensOf(expo.take())
		for i, tk := range got {
			got[i] = strings.TrimSuffix(strings.TrimPrefix(tk, "ExponentPushToken["), "]")
		}
		slices.Sort(got)
		return got
	}

	// The morning check records the win and pushes every follower who can get it.
	res, err := svc.CheckTriggers(ctx, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Triggered != 1 || res.Notified != 4 {
		t.Errorf("check-triggers = %+v, want 1 triggered and 4 notified", res)
	}
	if got, want := pushed(), []string{"claimer", "fan", "leaver", "silencer"}; !slices.Equal(got, want) {
		t.Errorf("you-won pushes reached %v, want %v (not paused, no token or a stranger)", got, want)
	}
	var token sql.NullString
	_ = database.QueryRow(`SELECT push_token FROM users WHERE id = 'uninstalled'`).Scan(&token)
	if token.Valid {
		t.Errorf("DeviceNotRegistered left the token %q", token.String)
	}
	var teID string
	if err := database.QueryRow(`SELECT id FROM triggered_events`).Scan(&teID); err != nil {
		t.Fatal(err)
	}

	// The retry run sends nothing new.
	clock = clock.Add(time.Minute)
	if _, err := svc.CheckTriggers(ctx, time.Time{}); err != nil {
		t.Fatal(err)
	}
	if got := pushed(); len(got) != 0 {
		t.Errorf("second check-triggers pushed %v again", got)
	}

	// Claim, silence and unfollow before the first reminder run.
	if _, err := database.Exec(`
		INSERT INTO dismissals (id, user_id, triggered_event_id, type) VALUES
			('d-c','claimer', ?, 'got_it'), ('d-s','silencer', ?, 'stop_reminding');`, teID, teID); err != nil {
		t.Fatal(err)
	}
	q := db.New(database)
	if err := q.KeepWonDeals(ctx, db.KeepWonDealsParams{UserID: "leaver", EventID: "event-test-smoke"}); err != nil {
		t.Fatal(err)
	}
	if err := q.DeleteSubscription(ctx, db.DeleteSubscriptionParams{UserID: "leaver", EventID: "event-test-smoke"}); err != nil {
		t.Fatal(err)
	}

	clock = clock.Add(time.Minute)
	if _, err := svc.SendReminders(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pushed(); !slices.Equal(got, []string{"fan"}) {
		t.Errorf("reminders reached %v, want [fan]", got)
	}
	clock = clock.Add(time.Minute)
	if _, err := svc.SendReminders(ctx); err != nil {
		t.Fatal(err)
	}
	if got := pushed(); len(got) != 0 {
		t.Errorf("a second reminder run the same day reached %v", got)
	}

	// The unfollower still sees the win; nobody gets a second "you won" push.
	won, err := q.ListActiveTriggeredEventsForUser(ctx, "leaver")
	if err != nil || len(won) != 1 {
		t.Errorf("leaver's live deals = %d (%v), want 1", len(won), err)
	}
	var trig int
	_ = database.QueryRow(`SELECT COUNT(*) FROM notifications WHERE kind = 'trigger' AND status = 'sent'
		GROUP BY user_id ORDER BY COUNT(*) DESC LIMIT 1`).Scan(&trig)
	if trig != 1 {
		t.Errorf("most trigger pushes to one user = %d, want 1", trig)
	}
}

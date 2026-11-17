package worker

import (
	"context"
	"database/sql"
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

// fakeExpo is an Expo push endpoint whose answer per message is set by the
// test: ticket(msg) returns the ticket for that message, and down makes the
// whole request fail with a 503.
type fakeExpo struct {
	mu     sync.Mutex
	sent   []notify.ExpoPushMessage // messages Expo accepted (status ok)
	posts  int                      // HTTP requests received
	down   bool                     // answer 503
	lose   bool                     // read the request, then drop the connection: delivery unknown
	ticket func(m notify.ExpoPushMessage) notify.ExpoPushTicket
	srv    *httptest.Server
}

func newFakeExpo(t *testing.T) *fakeExpo {
	t.Helper()
	f := &fakeExpo{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var batch []notify.ExpoPushMessage
		_ = json.Unmarshal(body, &batch)
		f.mu.Lock()
		defer f.mu.Unlock()
		f.posts++
		if f.lose {
			if conn, _, err := w.(http.Hijacker).Hijack(); err == nil {
				_ = conn.Close()
			}
			return
		}
		if f.down {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		tickets := make([]notify.ExpoPushTicket, len(batch))
		for i, m := range batch {
			tickets[i] = notify.ExpoPushTicket{Status: "ok", ID: "t"}
			if f.ticket != nil {
				tickets[i] = f.ticket(m)
			}
			if tickets[i].Status == "ok" {
				f.sent = append(f.sent, m)
			}
		}
		_ = json.NewEncoder(w).Encode(notify.ExpoPushResponse{Data: tickets})
	}))
	t.Cleanup(f.srv.Close)
	return f
}

// take returns and clears the accepted messages.
func (f *fakeExpo) take() []notify.ExpoPushMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.sent
	f.sent = nil
	return out
}

func (f *fakeExpo) set(down, lose bool) {
	f.mu.Lock()
	f.down, f.lose = down, lose
	f.mu.Unlock()
}

func errorTicket(code string) notify.ExpoPushTicket {
	tk := notify.ExpoPushTicket{Status: "error", Message: code}
	tk.Details.Error = code
	return tk
}

func tokensOf(msgs []notify.ExpoPushMessage) []string {
	out := make([]string, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, m.To)
	}
	return out
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	return database
}

func testService(database *sql.DB, expoURL string, clock *time.Time) *Service {
	n := notify.NewExpoNotifierWithURL(expoURL)
	n.SetRetryDelay(time.Millisecond)
	return &Service{
		queries:  db.New(database),
		notifier: n,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		now:      func() time.Time { return *clock },
	}
}

func statuses(t *testing.T, database *sql.DB, kind string) map[string][]string {
	t.Helper()
	rows, err := database.Query(`SELECT user_id, status FROM notifications WHERE kind = ? ORDER BY rowid`, kind)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string][]string{}
	for rows.Next() {
		var u, s string
		if err := rows.Scan(&u, &s); err != nil {
			t.Fatal(err)
		}
		out[u] = append(out[u], s)
	}
	return out
}

// seedDeal inserts users a and b following event-test-smoke and a live deal
// te-1 triggered an hour before now.
func seedDeal(t *testing.T, database *sql.DB, now time.Time) {
	t.Helper()
	if _, err := database.Exec(`
		INSERT INTO users (id, device_id, platform, push_token) VALUES
			('a','da','ios','ExponentPushToken[a]'), ('b','db','ios','ExponentPushToken[b]');
		INSERT INTO subscriptions (id, user_id, event_id, created_at) VALUES
			('sa','a','event-test-smoke', ?), ('sb','b','event-test-smoke', ?);
		INSERT INTO triggered_events (id, event_id, triggered_at, expires_at) VALUES ('te-1','event-test-smoke', ?, ?);`,
		sqlTime(now.Add(-48*time.Hour)), sqlTime(now.Add(-48*time.Hour)),
		sqlTime(now.Add(-time.Hour)), sqlTime(now.Add(12*time.Hour))); err != nil {
		t.Fatal(err)
	}
}

// Item 1: Expo rejects a's message and accepts b's. Each user's row must carry
// their own outcome, so b is not reminded twice and a is retried.
func TestReminderStatusFollowsEachTicket(t *testing.T) {
	expo := newFakeExpo(t)
	expo.ticket = func(m notify.ExpoPushMessage) notify.ExpoPushTicket {
		if m.To == "ExponentPushToken[a]" {
			return errorTicket("MessageRateExceeded")
		}
		return notify.ExpoPushTicket{Status: "ok", ID: "t"}
	}
	database := openTestDB(t)
	now := time.Now()
	seedDeal(t, database, now)
	svc := testService(database, expo.srv.URL, &now)

	if _, err := svc.SendReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := statuses(t, database, KindReminder)
	if strings.Join(got["a"], ",") != "failed" || strings.Join(got["b"], ",") != "sent" {
		t.Fatalf("after first run rows = %v, want a:[failed] b:[sent]", got)
	}
	expo.take()

	expo.ticket = nil // Expo healthy again
	if _, err := svc.SendReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	if to := tokensOf(expo.take()); len(to) != 1 || to[0] != "ExponentPushToken[a]" {
		t.Fatalf("second run pushed %v, want only a's retry", to)
	}
}

// Item 2: a DeviceNotRegistered ticket clears that token so we stop sending to it.
func TestDeviceNotRegisteredClearsToken(t *testing.T) {
	expo := newFakeExpo(t)
	expo.ticket = func(m notify.ExpoPushMessage) notify.ExpoPushTicket {
		if m.To == "ExponentPushToken[a]" {
			return errorTicket("DeviceNotRegistered")
		}
		return notify.ExpoPushTicket{Status: "ok", ID: "t"}
	}
	database := openTestDB(t)
	now := time.Now()
	seedDeal(t, database, now)
	svc := testService(database, expo.srv.URL, &now)

	if _, err := svc.SendReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	var a, b sql.NullString
	_ = database.QueryRow(`SELECT push_token FROM users WHERE id='a'`).Scan(&a)
	_ = database.QueryRow(`SELECT push_token FROM users WHERE id='b'`).Scan(&b)
	if a.Valid {
		t.Errorf("a's dead token = %q, want NULL", a.String)
	}
	if b.String != "ExponentPushToken[b]" {
		t.Errorf("b's token = %q, want it kept", b.String)
	}
}

// Item 7: a deal that starts at 6pm, after the last reminder run (5pm), still
// gets its unlock-day reminder, worded with the start time.
func TestLateStartDealGetsUnlockDayReminder(t *testing.T) {
	expo := newFakeExpo(t)
	database := openTestDB(t)
	n := time.Now().In(pacific)
	day := func(h int) time.Time { return time.Date(n.Year(), n.Month(), n.Day(), h, 0, 0, 0, pacific) }
	if _, err := database.Exec(`
		INSERT INTO users (id, device_id, platform, push_token) VALUES ('a','da','ios','ExponentPushToken[a]');
		INSERT INTO subscriptions (id, user_id, event_id) VALUES ('sa','a','event-test-smoke');
		INSERT INTO triggered_events (id, event_id, triggered_at, available_at, expires_at)
			VALUES ('te-late','event-test-smoke', ?, ?, ?);`,
		sqlTime(day(7)), sqlTime(day(18)), sqlTime(day(23).Add(59*time.Minute+59*time.Second))); err != nil {
		t.Fatal(err)
	}
	var clock time.Time
	svc := testService(database, expo.srv.URL, &clock)
	var all []notify.ExpoPushMessage
	for _, h := range []int{11, 15, 17} {
		clock = day(h)
		if _, err := svc.SendReminders(context.Background()); err != nil {
			t.Fatal(err)
		}
		all = append(all, expo.take()...)
	}
	if len(all) != 1 {
		t.Fatalf("reminders on the unlock day = %d, want 1", len(all))
	}
	if text := all[0].Title + " " + all[0].Body; !strings.Contains(text, "starts at 6pm") {
		t.Errorf("reminder %q / %q, want it to say it starts at 6pm", all[0].Title, all[0].Body)
	}
}

// Item 3: Expo is down at 7am, so the "you won" push fails after the deal is
// recorded. The 9am run sends it; later runs send nothing more. The 7am state
// (deal recorded, nobody notified) is also what a crash between recording the
// deal and sending leaves behind, so that case is covered by the same path.
func TestTriggerPushRetriedByLaterRun(t *testing.T) {
	expo := newFakeExpo(t)
	database := openTestDB(t)
	now := time.Now()
	seedDeal(t, database, now)
	if _, err := database.Exec(`
		INSERT INTO users (id, device_id, platform, push_token) VALUES ('late','dl','ios','ExponentPushToken[late]');
		INSERT INTO subscriptions (id, user_id, event_id, created_at) VALUES ('sl','late','event-test-smoke', ?);
		INSERT INTO triggered_events (id, event_id, triggered_at, expires_at, payload)
			VALUES ('te-test','event-test-smoke', ?, ?, '{"source": "cli-test"}');`,
		sqlTime(now), sqlTime(now.Add(-time.Hour)), sqlTime(now.Add(12*time.Hour))); err != nil {
		t.Fatal(err)
	}
	clock := now
	svc := testService(database, expo.srv.URL, &clock)
	ctx := context.Background()

	expo.set(true, false) // 7am: Expo down
	if n := svc.sendOwedTriggerPushes(ctx); n != 0 {
		t.Fatalf("7am with Expo down: %d sent, want 0", n)
	}
	if got := statuses(t, database, KindTrigger); strings.Join(got["a"], ",") != "failed" ||
		strings.Join(got["b"], ",") != "failed" {
		t.Fatalf("7am rows = %v, want a and b failed", got)
	}

	expo.set(false, false) // 9am: Expo back
	clock = now.Add(2 * time.Hour)
	if n := svc.sendOwedTriggerPushes(ctx); n != 2 {
		t.Fatalf("9am: %d sent, want 2", n)
	}
	got := expo.take()
	if to := strings.Join(tokensOf(got), ","); to != "ExponentPushToken[a],ExponentPushToken[b]" {
		t.Errorf("9am pushed %s, want a and b only (not the late follower, not the test deal)", to)
	}
	for _, m := range got {
		if m.Data["url"] != "freebie://deal/te-1" {
			t.Errorf("push url = %v, want freebie://deal/te-1", m.Data["url"])
		}
	}

	clock = now.Add(26 * time.Hour) // the deal expired 14h ago
	if n := svc.sendOwedTriggerPushes(ctx); n != 0 {
		t.Errorf("after expiry: %d sent, want 0", n)
	}
	clock = now.Add(4 * time.Hour)
	if n := svc.sendOwedTriggerPushes(ctx); n != 0 || len(expo.take()) != 0 {
		t.Errorf("a later run re-sent the trigger push (%d)", n)
	}
}

// Item 3/4: a user never gets two trigger pushes for one deal. When Expo may
// have taken the push (the response was lost) it is recorded "unknown" and
// not retried; a row left "pending" by a crash mid-send is not retried either.
func TestTriggerPushNeverDoubled(t *testing.T) {
	expo := newFakeExpo(t)
	database := openTestDB(t)
	now := time.Now()
	seedDeal(t, database, now)
	if _, err := database.Exec(`INSERT INTO notifications (id, user_id, triggered_event_id, status, kind, sent_at)
		VALUES ('n-crash','b','te-1','pending','trigger', ?)`, sqlTime(now)); err != nil {
		t.Fatal(err)
	}
	clock := now
	svc := testService(database, expo.srv.URL, &clock)
	ctx := context.Background()

	expo.set(false, true) // Expo reads the request, the response is lost
	if n := svc.sendOwedTriggerPushes(ctx); n != 0 {
		t.Fatalf("lost response: %d counted sent, want 0", n)
	}
	expo.mu.Lock()
	posts := expo.posts
	expo.mu.Unlock()
	if posts != 1 {
		t.Errorf("Expo got %d requests, want 1 (no retry once the request was written)", posts)
	}
	if got := statuses(t, database, KindTrigger); strings.Join(got["a"], ",") != "unknown" {
		t.Errorf("a's row = %v, want unknown", got["a"])
	}

	expo.set(false, false)
	clock = now.Add(2 * time.Hour)
	if n := svc.sendOwedTriggerPushes(ctx); n != 0 || len(expo.take()) != 0 {
		t.Errorf("a later run pushed again after unknown/pending outcomes (%d)", n)
	}
}

// Item 4 for reminders: a lost response is recorded "unknown" and the next
// run that day does not remind again.
func TestReminderNotRepeatedWhenOutcomeUnknown(t *testing.T) {
	expo := newFakeExpo(t)
	database := openTestDB(t)
	now := time.Now()
	seedDeal(t, database, now)
	svc := testService(database, expo.srv.URL, &now)

	expo.set(false, true)
	res, err := svc.SendReminders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Unknown != 2 || res.Failed != 0 {
		t.Errorf("result %+v, want 2 unknown and 0 failed", res)
	}
	expo.set(false, false)
	if _, err := svc.SendReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := expo.take(); len(got) != 0 {
		t.Errorf("second run re-sent %d reminders after an unknown outcome", len(got))
	}
}

// Item 5: the CLI's path (NotifyTriggered) is the worker's: same copy, deep
// link, paused users skipped, notifications recorded, and run twice it does
// not push twice.
func TestNotifyTriggeredMatchesCheckTriggers(t *testing.T) {
	expo := newFakeExpo(t)
	database := openTestDB(t)
	now := time.Date(2026, 11, 16, 7, 0, 0, 0, pacific)
	seedDeal(t, database, now)
	if _, err := database.Exec(`
		UPDATE users SET alerts_paused_until = ? WHERE id = 'b';
		UPDATE triggered_events SET payload = '{"source": "cli-trigger", "notify": true}',
			available_at = ?, expires_at = ? WHERE id = 'te-1';`,
		sqlTime(now.Add(24*time.Hour)),
		sqlTime(time.Date(2026, 11, 16, 10, 0, 0, 0, pacific)),
		sqlTime(time.Date(2026, 11, 16, 23, 59, 59, 0, pacific))); err != nil {
		t.Fatal(err)
	}
	svc := testService(database, expo.srv.URL, &now)
	ctx := context.Background()

	res, err := svc.NotifyTriggered(ctx, "te-1")
	if err != nil {
		t.Fatal(err)
	}
	got := expo.take()
	if res.Sent != 1 || len(got) != 1 || got[0].To != "ExponentPushToken[a]" {
		t.Fatalf("pushed %v (sent %d), want only a (b is paused)", tokensOf(got), res.Sent)
	}
	m := got[0]
	ev, err := svc.queries.GetEvent(ctx, "event-test-smoke")
	if err != nil {
		t.Fatal(err)
	}
	wantTitle, wantBody := unlockedMessage(dealCopy{Partner: ev.PartnerName, Offer: ev.OfferName,
		Condition: ev.TriggerCondition, HasExpiry: true,
		ExpiresAt:   time.Date(2026, 11, 16, 23, 59, 59, 0, pacific),
		AvailableAt: time.Date(2026, 11, 16, 10, 0, 0, 0, pacific)}, now)
	if m.Title != wantTitle || m.Body != wantBody {
		t.Errorf("push %q / %q, want the check-triggers copy %q / %q", m.Title, m.Body, wantTitle, wantBody)
	}
	if strings.Contains(m.Body, "24 hours") || !strings.Contains(m.Body, "starts at 10am") {
		t.Errorf("body %q", m.Body)
	}
	if m.Data["url"] != "freebie://deal/te-1" {
		t.Errorf("data.url = %v", m.Data["url"])
	}
	if rows := statuses(t, database, KindTrigger); strings.Join(rows["a"], ",") != "sent" || len(rows["b"]) != 0 {
		t.Errorf("notifications = %v, want a sent and nothing for paused b", rows)
	}
	if res, _ := svc.NotifyTriggered(ctx, "te-1"); res.Sent != 0 {
		t.Errorf("second NotifyTriggered sent %d, want 0", res.Sent)
	}
	// check-triggers' catch-up treats a --notify CLI deal like a real one.
	if n := svc.sendOwedTriggerPushes(ctx); n != 0 {
		t.Errorf("catch-up re-sent %d", n)
	}
}

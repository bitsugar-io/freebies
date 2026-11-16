package worker

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/notify"
)

type reminderEnv struct {
	svc   *Service
	db    *sql.DB
	sent  *atomic.Int64 // pushes the fake Expo server received
	clock time.Time
	day0  time.Time // midnight Pacific today, so SQL's "now" and the fake clock agree on the era
}

func newReminderEnv(t *testing.T) *reminderEnv {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	var sent atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msgs []notify.ExpoPushMessage
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &msgs)
		tickets := make([]notify.ExpoPushTicket, len(msgs))
		for i := range msgs {
			tickets[i] = notify.ExpoPushTicket{Status: "ok"}
		}
		sent.Add(int64(len(msgs)))
		_ = json.NewEncoder(w).Encode(notify.ExpoPushResponse{Data: tickets})
	}))
	t.Cleanup(srv.Close)

	if _, err := database.Exec(`
		INSERT INTO users (id, device_id, platform, push_token) VALUES ('u1','d1','ios','ExponentPushToken[u1]');
		INSERT INTO subscriptions (id, user_id, event_id) VALUES ('s1','u1','event-test-smoke');`); err != nil {
		t.Fatal(err)
	}
	n := time.Now().In(pacific)
	e := &reminderEnv{
		db:   database,
		sent: &sent,
		day0: time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, pacific),
	}
	e.svc = &Service{
		queries:  db.New(database),
		notifier: notify.NewExpoNotifierWithURL(srv.URL),
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		now:      func() time.Time { return e.clock },
	}
	return e
}

func sqlTime(t time.Time) string { return t.UTC().Format("2006-01-02 15:04:05") }

// deal inserts a triggered event triggered at day offset trig (9am), optionally
// unlocking at midnight of day unlock (-1 = immediately) and expiring at the
// end of day last.
func (e *reminderEnv) deal(t *testing.T, id string, trig, unlock, last int) {
	t.Helper()
	at := func(d, h, m, s int) time.Time {
		return time.Date(e.day0.Year(), e.day0.Month(), e.day0.Day()+d, h, m, s, 0, pacific)
	}
	var avail any
	if unlock >= 0 {
		avail = sqlTime(at(unlock, 0, 0, 0))
	}
	if _, err := e.db.Exec(`INSERT INTO triggered_events (id, event_id, triggered_at, expires_at, available_at)
		VALUES (?, 'event-test-smoke', ?, ?, ?)`,
		id, sqlTime(at(trig, 9, 0, 0)), sqlTime(at(last, 23, 59, 59)), avail); err != nil {
		t.Fatal(err)
	}
}

// run advances the clock to day offset d at hour h and returns pushes sent by that pass.
func (e *reminderEnv) run(t *testing.T, d, h int) int64 {
	t.Helper()
	e.clock = time.Date(e.day0.Year(), e.day0.Month(), e.day0.Day()+d, h, 0, 0, 0, pacific)
	before := e.sent.Load()
	if _, err := e.svc.SendReminders(t.Context()); err != nil {
		t.Fatal(err)
	}
	return e.sent.Load() - before
}

// A one-day deal is reminded once, not on every one of the day's runs.
func TestReminderOncePerDayForOneDayDeal(t *testing.T) {
	e := newReminderEnv(t)
	e.deal(t, "te-1", 0, -1, 0)
	got := []int64{e.run(t, 0, 9), e.run(t, 0, 13), e.run(t, 0, 19)}
	if got[0] != 1 || got[1] != 0 || got[2] != 0 {
		t.Errorf("pushes per run = %v, want [1 0 0]", got)
	}
	var n int
	_ = e.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE kind='reminder' AND status='sent'`).Scan(&n)
	if n != 1 {
		t.Errorf("recorded reminders = %d, want 1", n)
	}
}

// A delayed week-long deal: nothing before it unlocks, one reminder on the
// unlock day, none mid-window, one on its last day.
func TestReminderMultiDayOnlyUnlockAndFinalDay(t *testing.T) {
	e := newReminderEnv(t)
	e.deal(t, "te-w", 0, 2, 8) // triggered day 0, unlocks day 2, last day 8
	want := map[int]int64{0: 0, 1: 0, 2: 1, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0, 8: 1}
	for d := 0; d <= 8; d++ {
		var total int64
		for _, h := range []int{9, 13, 19} { // three runs a day
			total += e.run(t, d, h)
		}
		if total != want[d] {
			t.Errorf("day %d: %d pushes, want %d", d, total, want[d])
		}
	}
}

// A failed send is retried on the next run; a trigger push does not count as a
// reminder; another user's reminder does not suppress this one.
func TestReminderDedupSemantics(t *testing.T) {
	e := newReminderEnv(t)
	e.deal(t, "te-1", 0, -1, 0)
	e.clock = e.day0.Add(9 * time.Hour)
	ins := func(id, kind, status string) {
		t.Helper()
		if _, err := e.db.Exec(`INSERT INTO notifications (id, user_id, triggered_event_id, status, kind, sent_at)
			VALUES (?, 'u1', 'te-1', ?, ?, ?)`, id, status, kind, sqlTime(e.clock)); err != nil {
			t.Fatal(err)
		}
	}
	ins("n-trigger", KindTrigger, "sent")
	ins("n-failed", KindReminder, "failed")
	if got := e.run(t, 0, 10); got != 1 {
		t.Errorf("after trigger push + failed reminder: %d pushes, want 1", got)
	}
	if got := e.run(t, 0, 14); got != 0 {
		t.Errorf("after a sent reminder: %d pushes, want 0", got)
	}
}

func TestReminderDue(t *testing.T) {
	d := func(day, h int) time.Time { return time.Date(2026, 11, day, h, 0, 0, 0, pacific) }
	valid := func(t time.Time) sql.NullTime { return sql.NullTime{Time: t, Valid: true} }
	tests := []struct {
		name       string
		trig       time.Time
		avail, exp sql.NullTime
		now        time.Time
		want       bool
	}{
		{"same day", d(16, 7), sql.NullTime{}, valid(d(16, 23)), d(16, 20), true},
		{"mid window", d(16, 7), valid(d(18, 0)), valid(d(24, 0).Add(-time.Second)), d(20, 12), false},
		{"unlock day", d(16, 7), valid(d(18, 0)), valid(d(24, 0).Add(-time.Second)), d(18, 12), true},
		{"before unlock day", d(16, 7), valid(d(18, 0)), valid(d(24, 0).Add(-time.Second)), d(17, 12), false},
		{"final day", d(16, 7), valid(d(18, 0)), valid(d(24, 0).Add(-time.Second)), d(23, 12), true},
		{"midnight expiry is the day before", d(16, 7), sql.NullTime{}, valid(d(19, 0)), d(18, 12), true},
		{"no expiry only on the first day", d(16, 7), sql.NullTime{}, sql.NullTime{}, d(17, 12), false},
	}
	for _, tc := range tests {
		if got := reminderDue(tc.trig, tc.avail, tc.exp, tc.now); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

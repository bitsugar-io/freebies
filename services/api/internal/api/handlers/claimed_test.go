package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
	"github.com/retr0h/freebie/services/api/internal/db"
)

type claimedFixture struct {
	h        *handlers.Handler
	database *sql.DB
	eventID  string
}

func newClaimedFixture(t *testing.T) *claimedFixture {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	var eventID string
	if err := database.QueryRow(`SELECT id FROM events LIMIT 1`).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return &claimedFixture{h: handlers.New(database, logger), database: database, eventID: eventID}
}

func (f *claimedFixture) exec(t *testing.T, q string, args ...any) {
	t.Helper()
	if _, err := f.database.Exec(q, args...); err != nil {
		t.Fatal(err)
	}
}

func (f *claimedFixture) user(t *testing.T, id string) {
	f.exec(t, `INSERT INTO users (id, device_id, platform, token) VALUES (?, ?, 'ios', ?)`, id, "dev-"+id, "tok-"+id)
}

// trigger creates a triggered event whose window closes at expires (SQLite datetime expression).
func (f *claimedFixture) trigger(t *testing.T, id, expires string) {
	f.exec(t, `INSERT INTO triggered_events (id, event_id, expires_at) VALUES (?, ?, `+expires+`)`, id, f.eventID)
}

func (f *claimedFixture) dismiss(t *testing.T, user, te, typ, at string) {
	f.exec(t, `INSERT INTO dismissals (id, user_id, triggered_event_id, type, dismissed_at) VALUES (?, ?, ?, ?, ?)`,
		"d-"+user+"-"+te, user, te, typ, at)
}

func (f *claimedFixture) list(t *testing.T, user, query string) (int, []handlers.ClaimedDealResponse) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+user+"/claimed"+query, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userId", user)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	f.h.ListClaimedDeals(w, req)
	var got []handlers.ClaimedDealResponse
	if w.Code == http.StatusOK {
		if err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&got); err != nil {
			t.Fatal(err)
		}
	}
	return w.Code, got
}

func TestListClaimedDeals(t *testing.T) {
	f := newClaimedFixture(t)
	f.user(t, "u1")
	f.user(t, "u2")
	f.exec(t, `INSERT INTO subscriptions (id, user_id, event_id) VALUES ('s1', 'u1', ?)`, f.eventID)

	// Expired long ago, claimed first; and expired but claimed later; and one still active.
	f.trigger(t, "te-old", `datetime('now', '-3 days')`)
	f.trigger(t, "te-mid", `datetime('now', '-1 days')`)
	f.trigger(t, "te-live", `datetime('now', '+2 hours')`)
	f.trigger(t, "te-stop", `datetime('now', '-2 days')`)
	f.trigger(t, "te-other", `datetime('now', '-2 days')`)
	f.dismiss(t, "u1", "te-old", "got_it", "2026-11-01 12:00:00")
	f.dismiss(t, "u1", "te-mid", "got_it", "2026-11-10 12:00:00")
	f.dismiss(t, "u1", "te-live", "got_it", "2026-11-14 20:15:00")
	f.dismiss(t, "u1", "te-stop", "stop_reminding", "2026-11-12 12:00:00")
	f.dismiss(t, "u2", "te-other", "got_it", "2026-11-13 12:00:00")

	code, got := f.list(t, "u1", "")
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	var ids []string
	for _, c := range got {
		ids = append(ids, c.ID)
	}
	if want := "te-live,te-mid,te-old"; strings.Join(ids, ",") != want {
		t.Fatalf("ids = %v, want %s (expired claims kept, newest first, no stop_reminding, no other users)", ids, want)
	}
	first := got[0]
	if first.EventID != f.eventID || first.Event.ID != f.eventID {
		t.Fatalf("event ids = %q / %q, want %q", first.EventID, first.Event.ID, f.eventID)
	}
	if first.ClaimedAt != "2026-11-14T20:15:00Z" {
		t.Fatalf("claimedAt = %q", first.ClaimedAt)
	}
	if first.TriggeredAt == "" || first.Event.Sprite == nil {
		t.Fatalf("missing triggeredAt or sprite: %+v", first)
	}

	if _, other := f.list(t, "u2", ""); len(other) != 1 || other[0].ID != "te-other" {
		t.Fatalf("u2 claimed = %+v", other)
	}
}

func TestListClaimedDealsLimit(t *testing.T) {
	f := newClaimedFixture(t)
	f.user(t, "u1")
	for _, id := range []string{"a", "b", "c"} {
		f.trigger(t, "te-"+id, `datetime('now', '-1 days')`)
		f.dismiss(t, "u1", "te-"+id, "got_it", "2026-11-1"+map[string]string{"a": "0", "b": "1", "c": "2"}[id]+" 12:00:00")
	}
	if _, got := f.list(t, "u1", "?limit=2"); len(got) != 2 || got[0].ID != "te-c" {
		t.Fatalf("limit=2 -> %+v", got)
	}
	if _, got := f.list(t, "u1", "?limit=9999"); len(got) != 3 {
		t.Fatalf("limit clamps to max, got %d", len(got))
	}
	if code, _ := f.list(t, "u1", "?limit=0"); code != http.StatusBadRequest {
		t.Fatalf("limit=0 status = %d", code)
	}
	if code, got := f.list(t, "nobody", ""); code != http.StatusOK || len(got) != 0 {
		t.Fatalf("empty list: %d %v", code, got)
	}
}

package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
)

// TestAvailableAt: a delayed deal ("code unlocks Tue") carries availableAt in
// the active and claimed JSON; a normal deal omits it.
func TestAvailableAt(t *testing.T) {
	f := newClaimedFixture(t)
	f.user(t, "u1")
	f.exec(t, `INSERT INTO subscriptions (id, user_id, event_id) VALUES ('s1', 'u1', ?)`, f.eventID)
	f.exec(t, `INSERT INTO triggered_events (id, event_id, expires_at, available_at)
		VALUES ('te-late', ?, datetime('now', '+9 days'), '2026-11-17 08:00:00')`, f.eventID)
	f.exec(t, `INSERT INTO triggered_events (id, event_id, expires_at)
		VALUES ('te-now', ?, datetime('now', '+1 day'))`, f.eventID)
	f.dismiss(t, "u1", "te-late", "got_it", "2026-11-16 09:00:00")

	_, claimed := f.list(t, "u1", "")
	if len(claimed) != 1 || claimed[0].AvailableAt == nil ||
		*claimed[0].AvailableAt != "2026-11-17T08:00:00Z" {
		t.Fatalf("claimed availableAt = %+v", claimed)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/u1/deals", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userId", "u1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	f.h.ListActiveDeals(w, req)
	var deals []handlers.ActiveDealResponse
	if err := json.Unmarshal(w.Body.Bytes(), &deals); err != nil {
		t.Fatal(err)
	}
	got := map[string]*string{}
	for _, d := range deals {
		got[d.ID] = d.AvailableAt
	}
	if len(got) != 2 {
		t.Fatalf("want 2 active deals, got %v", got)
	}
	if got["te-late"] == nil || *got["te-late"] != "2026-11-17T08:00:00Z" {
		t.Errorf("te-late availableAt = %v", got["te-late"])
	}
	if got["te-now"] != nil {
		t.Errorf("te-now availableAt = %v, want omitted", *got["te-now"])
	}
}

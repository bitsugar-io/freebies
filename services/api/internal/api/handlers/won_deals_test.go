package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
)

// call runs one handler with chi URL params, returning the status and body.
func call(
	t *testing.T,
	fn http.HandlerFunc,
	method, path, body string,
	params map[string]string,
) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	fn(w, req)
	return w.Code, w.Body.Bytes()
}

func (f *claimedFixture) dismissVia(t *testing.T, user, te, typ string) (int, handlers.DismissalResponse) {
	t.Helper()
	code, body := call(t, f.h.CreateDismissal, http.MethodPost, "/api/v1/users/"+user+"/dismissals",
		`{"triggeredEventId":"`+te+`","type":"`+typ+`"}`, map[string]string{"userId": user})
	var got handlers.DismissalResponse
	if code < 300 {
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
	}
	return code, got
}

func (f *claimedFixture) active(t *testing.T, user string) []handlers.ActiveDealResponse {
	t.Helper()
	code, body := call(t, f.h.ListActiveDeals, http.MethodGet, "/api/v1/users/"+user+"/active-deals", "",
		map[string]string{"userId": user})
	if code != http.StatusOK {
		t.Fatalf("active-deals status = %d: %s", code, body)
	}
	var got []handlers.ActiveDealResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&got); err != nil {
		t.Fatal(err)
	}
	return got
}

// Silencing a deal then claiming it must work: the claim upgrades the dismissal, and a repeat of
// either is an idempotent 200 (a GOT IT that timed out on the phone but reached the server).
func TestCreateDismissalUpgradeAndRetry(t *testing.T) {
	f := newClaimedFixture(t)
	f.user(t, "u1")
	f.trigger(t, "te-1", `datetime('now', '+2 hours')`)

	code, d := f.dismissVia(t, "u1", "te-1", "stop_reminding")
	if code != http.StatusCreated || d.Type != "stop_reminding" {
		t.Fatalf("first silence: %d %+v", code, d)
	}
	if code, again := f.dismissVia(t, "u1", "te-1", "stop_reminding"); code != http.StatusOK || again.ID != d.ID {
		t.Fatalf("repeat silence: %d %+v, want 200 with the same row", code, again)
	}
	code, claimed := f.dismissVia(t, "u1", "te-1", "got_it")
	if code != http.StatusOK || claimed.Type != "got_it" || claimed.ID != d.ID {
		t.Fatalf("claim after silence: %d %+v, want 200 got_it on the same row", code, claimed)
	}
	if code, retry := f.dismissVia(t, "u1", "te-1", "got_it"); code != http.StatusOK || retry.Type != "got_it" {
		t.Fatalf("claim retry: %d %+v", code, retry)
	}
	// A claim is never downgraded by a late "stop reminders".
	if code, late := f.dismissVia(t, "u1", "te-1", "stop_reminding"); code != http.StatusOK || late.Type != "got_it" {
		t.Fatalf("silence after claim: %d %+v, want 200 still got_it", code, late)
	}
	if _, got := f.list(t, "u1", ""); len(got) != 1 || got[0].ID != "te-1" {
		t.Fatalf("claimed list = %+v, want te-1 once", got)
	}
}

// Follow, trigger, unfollow: the deal is still the user's.
func TestUnfollowKeepsWonDeal(t *testing.T) {
	f := newClaimedFixture(t)
	f.user(t, "u1")
	f.user(t, "u2")
	f.exec(t, `INSERT INTO subscriptions (id, user_id, event_id) VALUES ('s1', 'u1', ?)`, f.eventID)
	f.trigger(t, "te-won", `datetime('now', '+2 hours')`)
	f.trigger(t, "te-gone", `datetime('now', '-1 hours')`)

	if got := f.active(t, "u1"); len(got) != 1 || got[0].ID != "te-won" {
		t.Fatalf("before unfollow = %+v", got)
	}
	code, body := call(t, f.h.DeleteSubscription, http.MethodDelete, "/", "",
		map[string]string{"userId": "u1", "eventId": f.eventID})
	if code != http.StatusNoContent {
		t.Fatalf("unfollow: %d %s", code, body)
	}
	if got := f.active(t, "u1"); len(got) != 1 || got[0].ID != "te-won" {
		t.Fatalf("after unfollow = %+v, want te-won kept", got)
	}

	// A deal that triggers after the unfollow is not theirs.
	f.trigger(t, "te-later", `datetime('now', '+3 hours')`)
	if got := f.active(t, "u1"); len(got) != 1 {
		t.Fatalf("after a later trigger = %+v, want only te-won", got)
	}
	// Someone who never followed sees nothing.
	if got := f.active(t, "u2"); len(got) != 0 {
		t.Fatalf("u2 = %+v", got)
	}
}

// A pushed deal stays listed for the user who got the push, even with no follow.
func TestNotifiedDealIsListed(t *testing.T) {
	f := newClaimedFixture(t)
	f.user(t, "u1")
	f.trigger(t, "te-1", `datetime('now', '+2 hours')`)
	f.exec(t, `INSERT INTO notifications (id, user_id, triggered_event_id, status) VALUES ('n1', 'u1', 'te-1', 'sent')`)
	if got := f.active(t, "u1"); len(got) != 1 || got[0].ID != "te-1" {
		t.Fatalf("notified = %+v", got)
	}
}

func TestGetTriggered(t *testing.T) {
	f := newClaimedFixture(t)
	f.trigger(t, "te-1", `datetime('now', '+2 hours')`)
	code, body := call(t, f.h.GetTriggered, http.MethodGet, "/api/v1/triggered/te-1", "", map[string]string{"id": "te-1"})
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	var got handlers.LiveDealResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "te-1" || got.EventID != f.eventID || got.Event.ID != f.eventID || got.ExpiresAt == nil ||
		got.TriggeredAt == "" || got.Event.Sprite == nil {
		t.Fatalf("triggered = %+v", got)
	}
	if code, _ := call(t, f.h.GetTriggered, http.MethodGet, "/", "", map[string]string{"id": "nope"}); code != http.StatusNotFound {
		t.Fatalf("missing id status = %d", code)
	}
}

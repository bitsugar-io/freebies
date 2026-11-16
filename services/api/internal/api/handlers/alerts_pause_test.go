package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestAlertsPauseRoundTrip(t *testing.T) {
	h := setupTestHandler(t)
	r := chi.NewRouter()
	r.Get("/users/{id}", h.GetUser)
	r.Put("/users/{id}/alerts-pause", h.UpdateAlertsPause)

	// create a user
	w := httptest.NewRecorder()
	h.CreateUser(w, httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"deviceId":"dev-pause","platform":"ios"}`)))
	var u struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(w.Body).Decode(&u); err != nil || u.ID == "" {
		t.Fatalf("create user: %v %s", err, w.Body.String())
	}

	put := func(body string) (int, map[string]*string) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/users/"+u.ID+"/alerts-pause", strings.NewReader(body)))
		var out map[string]*string
		_ = json.NewDecoder(w.Body).Decode(&out)
		return w.Code, out
	}
	get := func() map[string]any {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/"+u.ID, nil))
		var out map[string]any
		_ = json.NewDecoder(w.Body).Decode(&out)
		return out
	}

	if _, ok := get()["alertsPausedUntil"]; ok {
		t.Fatal("alertsPausedUntil should be omitted by default")
	}

	until := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second).Format(time.RFC3339)
	code, out := put(`{"until":"` + until + `"}`)
	if code != http.StatusOK || out["pausedUntil"] == nil || *out["pausedUntil"] != until {
		t.Fatalf("pause: %d %v", code, out)
	}
	if got := get()["alertsPausedUntil"]; got != until {
		t.Fatalf("GET alertsPausedUntil = %v, want %s", got, until)
	}

	code, out = put(`{"until":"9999-12-31T00:00:00Z"}`)
	if code != http.StatusOK || *out["pausedUntil"] != "9999-12-31T00:00:00Z" {
		t.Fatalf("far future: %d %v", code, out)
	}
	if got := get()["alertsPausedUntil"]; got != "9999-12-31T00:00:00Z" {
		t.Fatalf("GET far future = %v", got)
	}

	code, out = put(`{"until":null}`)
	if code != http.StatusOK || out["pausedUntil"] != nil {
		t.Fatalf("resume: %d %v", code, out)
	}
	if _, ok := get()["alertsPausedUntil"]; ok {
		t.Fatal("alertsPausedUntil should be omitted after resume")
	}

	if code, _ := put(`{"until":"nope"}`); code != http.StatusBadRequest {
		t.Fatalf("bad until: %d", code)
	}
}

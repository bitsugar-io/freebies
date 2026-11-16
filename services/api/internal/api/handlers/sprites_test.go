package handlers_test

import (
	"context"
	"encoding/json"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/sprites"
)

func spriteRequest(file, query string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sprites/"+file+query, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("file", file)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestListSprites(t *testing.T) {
	h := setupTestHandler(t)
	w := httptest.NewRecorder()
	h.ListSprites(w, httptest.NewRequest(http.MethodGet, "/api/v1/sprites", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var got []sprites.Sprite
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range got {
		if s.ID == handlers.FallbackSpriteID {
			found = true
		}
	}
	if !found {
		t.Fatalf("fallback sprite %q missing from %d sprites", handlers.FallbackSpriteID, len(got))
	}
}

func TestGetSprite(t *testing.T) {
	h := setupTestHandler(t)

	t.Run("json", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.GetSprite(w, spriteRequest("gift", ""))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		var s sprites.Sprite
		if err := json.NewDecoder(w.Body).Decode(&s); err != nil {
			t.Fatal(err)
		}
		if s.Width != 16 || len(s.Pixels) != 16 {
			t.Fatalf("unexpected sprite %+v", s)
		}
	})

	t.Run("png with scale", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.GetSprite(w, spriteRequest("gift.png", "?scale=4"))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); ct != "image/png" {
			t.Fatalf("content-type = %q", ct)
		}
		img, err := png.Decode(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != 64 {
			t.Fatalf("width = %d, want 64", img.Bounds().Dx())
		}
	})

	t.Run("png default scale", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.GetSprite(w, spriteRequest("gift.png", ""))
		img, err := png.Decode(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != 128 {
			t.Fatalf("width = %d, want 128", img.Bounds().Dx())
		}
	})

	t.Run("bad scale", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.GetSprite(w, spriteRequest("gift.png", "?scale=big"))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.GetSprite(w, spriteRequest("nope.png", ""))
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
	})
}

func TestEventsCarryFallbackSprite(t *testing.T) {
	h := setupTestHandler(t)
	w := httptest.NewRecorder()
	h.ListEvents(w, httptest.NewRequest(http.MethodGet, "/api/v1/events", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var events []handlers.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&events); err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected seeded events")
	}
	for _, e := range events {
		if e.Sprite == nil {
			t.Fatalf("event %s has no sprite", e.ID)
		}
		want := handlers.FallbackSpriteID
		if e.SpriteID != nil {
			want = *e.SpriteID
		}
		if e.Sprite.ID != want {
			t.Fatalf("event %s sprite = %s, want %s", e.ID, e.Sprite.ID, want)
		}
	}
}

// The parsed sprites are cached: once loaded, requests do not touch the
// sprites table again until the TTL passes (here the table is gone and the
// list still serves), and one load serves both endpoints that need it.
func TestSpritesAreCached(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	h := handlers.New(database, slog.New(slog.NewTextHandler(io.Discard, nil)))

	list := func() (int, []sprites.Sprite) {
		w := httptest.NewRecorder()
		h.ListSprites(w, httptest.NewRequest(http.MethodGet, "/api/v1/sprites", nil))
		var got []sprites.Sprite
		_ = json.NewDecoder(w.Body).Decode(&got)
		return w.Code, got
	}
	code, first := list()
	if code != http.StatusOK || len(first) == 0 {
		t.Fatalf("first list: status %d, %d sprites", code, len(first))
	}
	for i := 1; i < len(first); i++ {
		if first[i-1].ID >= first[i].ID {
			t.Fatalf("sprites not in ID order: %q before %q", first[i-1].ID, first[i].ID)
		}
	}

	if _, err := database.Exec(`ALTER TABLE sprites RENAME TO sprites_gone`); err != nil {
		t.Fatal(err)
	}
	code, second := list()
	if code != http.StatusOK || len(second) != len(first) {
		t.Errorf("second list hit the database: status %d, %d sprites (want %d from cache)",
			code, len(second), len(first))
	}
}

// A failed load is an error response and is not cached.
func TestSpritesQueryErrorIsNotCached(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`ALTER TABLE sprites RENAME TO sprites_gone`); err != nil {
		t.Fatal(err)
	}
	h := handlers.New(database, slog.New(slog.NewTextHandler(io.Discard, nil)))
	w := httptest.NewRecorder()
	h.ListSprites(w, httptest.NewRequest(http.MethodGet, "/api/v1/sprites", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if _, err := database.Exec(`ALTER TABLE sprites_gone RENAME TO sprites`); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	h.ListSprites(w, httptest.NewRequest(http.MethodGet, "/api/v1/sprites", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("after recovery status = %d, want 200", w.Code)
	}
}

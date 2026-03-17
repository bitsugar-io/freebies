package handlers_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
	"github.com/retr0h/freebie/services/api/internal/db"
)

func setupTestHandler(t *testing.T) *handlers.Handler {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return handlers.New(database, logger)
}

func TestGetConfig(t *testing.T) {
	h := setupTestHandler(t)

	t.Run("returns features and screens from seed data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
		w := httptest.NewRecorder()

		h.GetConfig(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var config handlers.ConfigResponse
		if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(config.Features) != 8 {
			t.Fatalf("expected 8 feature flags, got %d", len(config.Features))
		}
		if config.Features["enable_mlb"] != true {
			t.Fatal("expected enable_mlb to be true")
		}
		if config.Features["enable_nhl"] != false {
			t.Fatal("expected enable_nhl to be false")
		}

		if len(config.Screens) != 3 {
			t.Fatalf("expected 3 screens, got %d", len(config.Screens))
		}
		if len(config.Screens["deals"]) != 1 {
			t.Fatalf("expected 1 block in deals, got %d", len(config.Screens["deals"]))
		}
		if config.Screens["deals"][0].Type != "active_deals" {
			t.Fatalf("expected active_deals block, got %s", config.Screens["deals"][0].Type)
		}
		if len(config.Screens["discover"]) != 2 {
			t.Fatalf("expected 2 blocks in discover, got %d", len(config.Screens["discover"]))
		}
		if len(config.Screens["profile"]) != 3 {
			t.Fatalf("expected 3 blocks in profile, got %d", len(config.Screens["profile"]))
		}

		cc := resp.Header.Get("Cache-Control")
		if cc != "public, max-age=60" {
			t.Fatalf("expected Cache-Control 'public, max-age=60', got %q", cc)
		}

		dealsConfig := config.Screens["deals"][0].Config
		if dealsConfig["layout"] != "list" {
			t.Fatalf("expected deals config layout=list, got %v", dealsConfig["layout"])
		}
	})
}

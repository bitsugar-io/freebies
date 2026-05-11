package api

import (
	"bytes"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/retr0h/freebie/services/api/internal/config"
	"github.com/retr0h/freebie/services/api/internal/db"
)

// TestHealthzNotLogged asserts the routing change that pulls /healthz out of
// the chi middleware.Logger stack. Without it, kubelet probes (every ~5s)
// drown real traffic in the pod log buffer.
func TestHealthzNotLogged(t *testing.T) {
	handler, logBuf := newTestServer(t)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("/healthz status = %d, want 200", rec.Code)
	}
	if got := logBuf.String(); strings.Contains(got, "/healthz") {
		t.Errorf("/healthz produced middleware log output:\n%s", got)
	}
}

// TestRealRouteIsLogged is the counterpart — proves we didn't accidentally
// drop the logger from everywhere. A regular API request must still log.
func TestRealRouteIsLogged(t *testing.T) {
	handler, logBuf := newTestServer(t)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/leagues", nil))

	if got := logBuf.String(); !strings.Contains(got, "/api/v1/leagues") {
		t.Errorf("/api/v1/leagues produced no middleware log output:\n%s", got)
	}
}

// newTestServer boots a Server backed by an in-memory SQLite DB and replaces
// chi's package-level DefaultLogger so its access-log output lands in a
// buffer instead of os.Stdout.
func newTestServer(t *testing.T) (http.Handler, *bytes.Buffer) {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var buf bytes.Buffer
	origLogger := middleware.DefaultLogger
	middleware.DefaultLogger = middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  log.New(&buf, "", 0),
		NoColor: true,
	})
	t.Cleanup(func() { middleware.DefaultLogger = origLogger })

	cfg := &config.Config{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	srv := NewServer(cfg, database, logger, nil)
	return srv.Router(), &buf
}

package handlers_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/api/handlers"
	"github.com/retr0h/freebie/services/api/internal/db"
)

var hexColor = regexp.MustCompile(`^#[0-9A-F]{6}$`)

func TestInSeason(t *testing.T) {
	tests := []struct {
		name       string
		day        string
		start, end string
		want       bool
	}{
		{"inside plain range", "2026-07-01", "03-20", "10-31", true},
		{"start boundary", "2026-03-20", "03-20", "10-31", true},
		{"end boundary", "2026-10-31", "03-20", "10-31", true},
		{"day before start", "2026-03-19", "03-20", "10-31", false},
		{"day after end", "2026-11-01", "03-20", "10-31", false},
		{"wrap: before new year", "2026-11-15", "09-01", "02-15", true},
		{"wrap: after new year", "2027-01-10", "09-01", "02-15", true},
		{"wrap: end boundary", "2027-02-15", "09-01", "02-15", true},
		{"wrap: just after end", "2027-02-16", "09-01", "02-15", false},
		{"wrap: summer gap", "2026-07-04", "09-01", "02-15", false},
		{"wrap: start boundary", "2026-09-01", "09-01", "02-15", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			day, err := time.Parse("2006-01-02", tc.day)
			if err != nil {
				t.Fatal(err)
			}
			if got := handlers.InSeason(day, tc.start, tc.end); got != tc.want {
				t.Fatalf("InSeason(%s, %s, %s) = %v, want %v", tc.day, tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestReturnLabel(t *testing.T) {
	override := "FEB 11"
	if got := handlers.ReturnLabel(true, "03-20", nil); got != nil {
		t.Fatalf("in season: want nil, got %q", *got)
	}
	if got := handlers.ReturnLabel(false, "03-20", nil); got == nil || *got != "MAR" {
		t.Fatalf("want MAR, got %v", got)
	}
	if got := handlers.ReturnLabel(false, "02-11", &override); got == nil || *got != "FEB 11" {
		t.Fatalf("want override, got %v", got)
	}
	if got := handlers.ReturnLabel(false, "bad", nil); got != nil {
		t.Fatalf("bad start: want nil, got %q", *got)
	}
}

func TestListTeams(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	h := handlers.New(database, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

	get := func(query string) (int, []handlers.TeamResponse) {
		w := httptest.NewRecorder()
		h.ListTeams(w, httptest.NewRequest(http.MethodGet, "/api/v1/teams"+query, nil))
		var out []handlers.TeamResponse
		if w.Code == http.StatusOK {
			if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
				t.Fatal(err)
			}
		}
		return w.Code, out
	}

	t.Run("2026-11-15", func(t *testing.T) {
		code, teams := get("?at=2026-11-15")
		if code != http.StatusOK || len(teams) != 14 {
			t.Fatalf("code %d, %d teams", code, len(teams))
		}
		wantIn := map[string]bool{"LAL": true, "LAC": true, "LAR": true, "LACH": true, "LAK": true,
			"ANA": true, "LAFC": true, "ACFC": true, "USC": true, "UCLA": true}
		// The Galaxy missed the 2026 playoffs (045): out of season, back in February.
		wantLabel := map[string]string{"LAD": "MAR", "LAA": "MAR", "LAS": "MAY", "LAG": "FEB"}
		seenOff := false
		for _, tm := range teams {
			if tm.InSeason != wantIn[tm.ID] {
				t.Errorf("%s inSeason = %v", tm.ID, tm.InSeason)
			}
			if tm.InSeason {
				if seenOff {
					t.Errorf("%s in season after an off-season team", tm.ID)
				}
				if tm.ReturnLabel != nil {
					t.Errorf("%s has returnLabel while in season", tm.ID)
				}
			} else {
				seenOff = true
				if tm.ReturnLabel == nil || *tm.ReturnLabel != wantLabel[tm.ID] {
					t.Errorf("%s returnLabel = %v, want %s", tm.ID, tm.ReturnLabel, wantLabel[tm.ID])
				}
			}
			if tm.LeagueName == "" || tm.Badge == nil || tm.BadgeSpriteID == nil || tm.Badge.ID != *tm.BadgeSpriteID {
				t.Errorf("%s missing leagueName/badge: %+v", tm.ID, tm)
			}
			if tm.Colors == nil || !hexColor.MatchString(tm.Colors.Primary) || !hexColor.MatchString(tm.Colors.Secondary) {
				t.Errorf("%s colors invalid: %+v", tm.ID, tm.Colors)
			}
			var n int
			if err := database.QueryRow(`SELECT COUNT(*) FROM events WHERE is_active = 1 AND team_id = ?`, tm.ID).Scan(&n); err != nil {
				t.Fatal(err)
			}
			if tm.ActiveDeals != n {
				t.Errorf("%s activeDeals = %d, want %d", tm.ID, tm.ActiveDeals, n)
			}
		}
	})

	t.Run("lakers badge and league name", func(t *testing.T) {
		_, teams := get("?at=2026-11-15")
		for _, tm := range teams {
			if tm.ID == "LAK" && (tm.LeagueName != "NHL" || *tm.BadgeSpriteID != "sport-hockey") {
				t.Errorf("LAK = %+v", tm)
			}
			if tm.ID == "LAL" && (tm.Colors == nil || tm.Colors.Primary != "#4428BC" || tm.Colors.Secondary != "#F8B800") {
				t.Errorf("LAL colors = %+v", tm.Colors)
			}
			if tm.ID == "USC" && tm.LeagueName != "NCAA" {
				t.Errorf("USC leagueName = %q", tm.LeagueName)
			}
		}
	})

	t.Run("angel city override label off-season", func(t *testing.T) {
		_, teams := get("?at=2026-12-25")
		for _, tm := range teams {
			if tm.ID == "ACFC" && (tm.InSeason || tm.ReturnLabel == nil || *tm.ReturnLabel != "FEB 11") {
				t.Errorf("ACFC = %+v", tm)
			}
		}
	})

	t.Run("bad at", func(t *testing.T) {
		if code, _ := get("?at=nope"); code != http.StatusBadRequest {
			t.Fatalf("got %d", code)
		}
	})
}

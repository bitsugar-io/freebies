package handlers_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/retr0h/freebie/services/api/internal/db"
)

// TestOnlyLAEventsActive guards the LA-only pivot: after all migrations every
// active event is in Los Angeles or Orange County and has a sprite.
func TestOnlyLAEventsActive(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	rows, err := database.Query(`SELECT id, region_code, COALESCE(sprite_id, '') FROM events WHERE is_active = 1`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var active []string
	for rows.Next() {
		var id, region, sprite string
		if err := rows.Scan(&id, &region, &sprite); err != nil {
			t.Fatal(err)
		}
		if region != "us-ca-la" && region != "us-ca-oc" {
			t.Errorf("active event %s is outside LA: region %q", id, region)
		}
		if sprite == "" {
			t.Errorf("active event %s has no sprite", id)
		}
		active = append(active, id)
	}
	sort.Strings(active)

	// Only deals that are in season, not in-venue-only, and whose metric a
	// source emits (see TestSeededActiveDealsCanFire in internal/triggers).
	want := []string{
		"event-nba-lac-chickfila", "event-nba-lal-jitb", "event-ncaa-ucla-fatsals", "event-ncaa-ucla-fatsals-hoops", "event-nfl-lach-firehouse",
		"event-nfl-lar-carlsjr", "event-nhl-ana-chickfila", "event-nhl-ana-mcdonalds", "event-nhl-lak-mcdonalds",
	}
	if !reflect.DeepEqual(active, want) {
		t.Fatalf("active events\n got %v\nwant %v", active, want)
	}

	var kings int
	if err := database.QueryRow(`SELECT is_active FROM events WHERE id = 'event-nhl-lak-elpolloloco'`).Scan(&kings); err != nil || kings != 0 {
		t.Fatalf("kings event should exist inactive, got active=%d err=%v", kings, err)
	}
}

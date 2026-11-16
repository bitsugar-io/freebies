package db

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
)

// Sprite migrations are generated as diffs (design skill, sprites.py), so each
// one's Down must undo only what its Up added. Walk them down one at a time and
// back up, checking the sprite set at every step.
func TestSpriteMigrationsRollBackIndependently(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}

	steps := []struct {
		downTo  int64
		gone    []string
		present []string
	}{
		{29, []string{"sport-hockey", "sport-basketball"}, []string{"street-tacos", "burger", "gift"}},
		{26, []string{"street-tacos"}, []string{"burger", "taco", "gift"}},
		{24, []string{"burger", "taco", "griddle-sandwich"}, []string{"gift"}},
	}
	for _, s := range steps {
		if err := goose.DownTo(database, "migrations", s.downTo); err != nil {
			t.Fatalf("down to %d: %v", s.downTo, err)
		}
		for _, id := range s.gone {
			if spriteExists(t, database, id) {
				t.Errorf("after down to %d: sprite %q should be gone", s.downTo, id)
			}
		}
		for _, id := range s.present {
			if !spriteExists(t, database, id) {
				t.Errorf("after down to %d: sprite %q should remain", s.downTo, id)
			}
		}
	}

	// At 024 the gift is back to its original palette and no deal points at a sprite.
	var palette string
	if err := database.QueryRow(`SELECT palette FROM sprites WHERE id = 'gift'`).Scan(&palette); err != nil {
		t.Fatal(err)
	}
	if palette != `{"k":"#000000","y":"#F8D878","r":"#F83800","R":"#A81000"}` {
		t.Errorf("gift palette not restored to 024: %s", palette)
	}
	var mapped int
	if err := database.QueryRow(`SELECT COUNT(*) FROM events WHERE sprite_id IS NOT NULL`).Scan(&mapped); err != nil {
		t.Fatal(err)
	}
	if mapped != 0 {
		t.Errorf("%d events still point at a sprite after rolling back to 024", mapped)
	}

	if err := Migrate(database); err != nil {
		t.Fatalf("migrating back up: %v", err)
	}
	for _, id := range []string{"gift", "burger", "street-tacos", "sport-soccer"} {
		if !spriteExists(t, database, id) {
			t.Errorf("after re-migrating: sprite %q missing", id)
		}
	}
}

func spriteExists(t *testing.T, database *sql.DB, id string) bool {
	t.Helper()
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM sprites WHERE id = ?`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n == 1
}

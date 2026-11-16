package sprites_test

import (
	"bytes"
	"context"
	"image/png"
	"testing"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/sprites"
)

func checker() sprites.Sprite {
	return sprites.Sprite{
		ID: "checker", Name: "Checker", Width: 2, Height: 2,
		Pixels:  []string{"a.", ".b"},
		Palette: map[string]string{"a": "#FF0000", "b": "#0000FF"},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*sprites.Sprite)
		wantErr bool
	}{
		{"valid", func(*sprites.Sprite) {}, false},
		{"wrong row count", func(s *sprites.Sprite) { s.Height = 3 }, true},
		{"ragged row", func(s *sprites.Sprite) { s.Pixels[1] = ".bb" }, true},
		{"unknown key", func(s *sprites.Sprite) { s.Pixels[0] = "z." }, true},
		{"bad color", func(s *sprites.Sprite) { s.Palette["a"] = "red" }, true},
		{"transparent key in palette", func(s *sprites.Sprite) { s.Palette["."] = "#000000" }, true},
		{"zero size", func(s *sprites.Sprite) { s.Width = 0 }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := checker()
			tt.mutate(&s)
			if err := s.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageScalesAndKeepsTransparency(t *testing.T) {
	img := checker().Image(3)
	if b := img.Bounds(); b.Dx() != 6 || b.Dy() != 6 {
		t.Fatalf("bounds = %v, want 6x6", b)
	}
	if c := img.NRGBAAt(2, 2); c.R != 255 || c.A != 255 {
		t.Fatalf("pixel (2,2) = %v, want opaque red", c)
	}
	if c := img.NRGBAAt(4, 1); c.A != 0 {
		t.Fatalf("pixel (4,1) = %v, want transparent", c)
	}
	if c := img.NRGBAAt(5, 5); c.B != 255 {
		t.Fatalf("pixel (5,5) = %v, want blue", c)
	}
}

func TestClampScale(t *testing.T) {
	for in, want := range map[int]int{-4: 1, 0: 1, 8: 8, 999: sprites.MaxScale} {
		if got := sprites.ClampScale(in); got != want {
			t.Errorf("ClampScale(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestWritePNG(t *testing.T) {
	var buf bytes.Buffer
	if err := checker().WritePNG(&buf, 4); err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("not a PNG: %v", err)
	}
	if img.Bounds().Dx() != 8 {
		t.Fatalf("width = %d, want 8", img.Bounds().Dx())
	}
}

// Every sprite seeded by migrations must decode and validate.
func TestSeededSpritesAreValid(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	rows, err := db.New(database).ListSprites(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("expected seeded sprites")
	}
	for _, row := range rows {
		if _, err := sprites.FromDB(row); err != nil {
			t.Errorf("%v", err)
		}
	}
}

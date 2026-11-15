// Package sprites parses, validates and renders pixel-art deal sprites.
//
// A sprite is stored as data (rows of palette keys plus a key -> color map)
// so the mobile app can draw it natively at any size and the API can render
// the same sprite to PNG for push notification attachments and sharing.
package sprites

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"strconv"

	"github.com/retr0h/freebie/services/api/internal/db"
)

// Transparent is the pixel key that draws nothing.
const Transparent = '.'

// MaxScale caps PNG rendering so a request can't ask for a huge image.
const MaxScale = 32

// Sprite is a decoded, render-ready sprite.
type Sprite struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Width   int               `json:"width"`
	Height  int               `json:"height"`
	Pixels  []string          `json:"pixels"`
	Palette map[string]string `json:"palette"`
}

// FromDB decodes a database row and validates it.
func FromDB(row db.Sprite) (Sprite, error) {
	s := Sprite{
		ID:     row.ID,
		Name:   row.Name,
		Width:  int(row.Width),
		Height: int(row.Height),
	}
	if err := json.Unmarshal([]byte(row.Pixels), &s.Pixels); err != nil {
		return Sprite{}, fmt.Errorf("sprite %q: decode pixels: %w", row.ID, err)
	}
	if err := json.Unmarshal([]byte(row.Palette), &s.Palette); err != nil {
		return Sprite{}, fmt.Errorf("sprite %q: decode palette: %w", row.ID, err)
	}
	if err := s.Validate(); err != nil {
		return Sprite{}, err
	}
	return s, nil
}

// Validate checks dimensions, palette keys and color values.
func (s Sprite) Validate() error {
	if s.Width <= 0 || s.Height <= 0 {
		return fmt.Errorf("sprite %q: width and height must be positive", s.ID)
	}
	if len(s.Pixels) != s.Height {
		return fmt.Errorf("sprite %q: has %d rows, want %d", s.ID, len(s.Pixels), s.Height)
	}
	for key, hex := range s.Palette {
		if len(key) != 1 || key[0] == Transparent {
			return fmt.Errorf("sprite %q: invalid palette key %q", s.ID, key)
		}
		if _, err := parseHex(hex); err != nil {
			return fmt.Errorf("sprite %q: palette %q: %w", s.ID, key, err)
		}
	}
	for y, row := range s.Pixels {
		if len(row) != s.Width {
			return fmt.Errorf("sprite %q: row %d is %d wide, want %d", s.ID, y, len(row), s.Width)
		}
		for x := 0; x < len(row); x++ {
			if row[x] == Transparent {
				continue
			}
			if _, ok := s.Palette[string(row[x])]; !ok {
				return fmt.Errorf("sprite %q: row %d col %d uses key %q not in palette", s.ID, y, x, row[x])
			}
		}
	}
	return nil
}

// Image renders the sprite with each pixel drawn as a scale x scale square.
// Transparent pixels stay fully transparent.
func (s Sprite) Image(scale int) *image.NRGBA {
	scale = ClampScale(scale)
	colors := make(map[byte]color.NRGBA, len(s.Palette))
	for key, hex := range s.Palette {
		c, _ := parseHex(hex) // validated already
		colors[key[0]] = c
	}
	img := image.NewNRGBA(image.Rect(0, 0, s.Width*scale, s.Height*scale))
	for y, row := range s.Pixels {
		for x := 0; x < len(row); x++ {
			c, ok := colors[row[x]]
			if !ok {
				continue
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					img.SetNRGBA(x*scale+dx, y*scale+dy, c)
				}
			}
		}
	}
	return img
}

// WritePNG encodes the sprite as PNG at the given scale.
func (s Sprite) WritePNG(w io.Writer, scale int) error {
	return png.Encode(w, s.Image(scale))
}

// ClampScale keeps a requested scale within 1..MaxScale.
func ClampScale(scale int) int {
	if scale < 1 {
		return 1
	}
	if scale > MaxScale {
		return MaxScale
	}
	return scale
}

func parseHex(hex string) (color.NRGBA, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return color.NRGBA{}, fmt.Errorf("color %q is not #RRGGBB", hex)
	}
	v, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("color %q is not #RRGGBB", hex)
	}
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}, nil
}

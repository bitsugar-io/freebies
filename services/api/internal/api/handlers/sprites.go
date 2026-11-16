package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/retr0h/freebie/services/api/internal/sprites"
	"github.com/retr0h/freebie/services/api/internal/ttlcache"
)

// FallbackSpriteID is drawn for any deal that has no sprite of its own.
const FallbackSpriteID = "gift"

// defaultPNGScale renders a 16x16 sprite at 128x128.
const defaultPNGScale = 8

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// spriteCacheTTL bounds how stale served sprites can be. Sprites change only
// through migrations, which restart the server, so a minute is conservative.
const spriteCacheTTL = time.Minute

// spriteSet is every valid sprite, parsed once: in ID order for the list
// endpoint and keyed by ID for lookups. Treat it as read-only.
type spriteSet struct {
	list []sprites.Sprite
	byID map[string]sprites.Sprite
}

// spriteCaches holds each Handler's parsed sprites. It lives beside Handler
// (rather than as a field) so this file stays self-contained.
var spriteCaches sync.Map // *Handler -> *ttlcache.Cache[*spriteSet]

func (h *Handler) spriteCache() *ttlcache.Cache[*spriteSet] {
	if c, ok := spriteCaches.Load(h); ok {
		return c.(*ttlcache.Cache[*spriteSet])
	}
	c, _ := spriteCaches.LoadOrStore(h, ttlcache.New[*spriteSet](spriteCacheTTL, 1))
	return c.(*ttlcache.Cache[*spriteSet])
}

// loadSpriteSet returns the cached sprites, running one query when the cache is
// cold or older than spriteCacheTTL. Invalid rows are logged and skipped so
// one bad sprite can't break a deals response. It returns nil on a query error
// (which is not cached).
func (h *Handler) loadSpriteSet(ctx context.Context) *spriteSet {
	set, err := h.spriteCache().Get(ctx, "all", func(ctx context.Context) (*spriteSet, error) {
		rows, err := h.queries.ListSprites(ctx)
		if err != nil {
			return nil, err
		}
		out := &spriteSet{
			list: make([]sprites.Sprite, 0, len(rows)),
			byID: make(map[string]sprites.Sprite, len(rows)),
		}
		for _, row := range rows { // ordered by id
			s, err := sprites.FromDB(row)
			if err != nil {
				h.logger.Warn("skipping invalid sprite", "error", err)
				continue
			}
			out.list = append(out.list, s)
			out.byID[s.ID] = s
		}
		return out, nil
	})
	if err != nil {
		h.logger.Error("failed to list sprites", "error", err)
		return nil
	}
	return set
}

// loadSprites returns every valid sprite keyed by ID, or nil on error. The map
// is shared; do not modify it.
func (h *Handler) loadSprites(ctx context.Context) map[string]sprites.Sprite {
	set := h.loadSpriteSet(ctx)
	if set == nil {
		return nil
	}
	return set.byID
}

// attachSprites fills in each event's Sprite from its SpriteID, falling back
// to the gift sprite.
func (h *Handler) attachSprites(ctx context.Context, events ...*EventResponse) {
	if len(events) == 0 {
		return
	}
	all := h.loadSprites(ctx)
	for _, e := range events {
		id := FallbackSpriteID
		if e.SpriteID != nil {
			id = *e.SpriteID
		}
		s, ok := all[id]
		if !ok {
			s, ok = all[FallbackSpriteID]
		}
		if ok {
			e.Sprite = &s
		}
	}
}

// ListSprites returns every sprite so the app can cache them.
func (h *Handler) ListSprites(w http.ResponseWriter, r *http.Request) {
	set := h.loadSpriteSet(r.Context())
	if set == nil {
		respondError(w, http.StatusInternalServerError, "failed to list sprites")
		return
	}
	respondJSON(w, http.StatusOK, set.list)
}

// GetSprite serves /sprites/{id} as JSON and /sprites/{id}.png as an image.
// PNG takes ?scale=N (1-32, default 8): each sprite pixel becomes N x N.
func (h *Handler) GetSprite(w http.ResponseWriter, r *http.Request) {
	file := chi.URLParam(r, "file")
	id, asPNG := strings.CutSuffix(file, ".png")

	row, err := h.queries.GetSprite(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		respondError(w, http.StatusNotFound, "sprite not found")
		return
	}
	if err != nil {
		h.logger.Error("failed to get sprite", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to get sprite")
		return
	}
	s, err := sprites.FromDB(row)
	if err != nil {
		h.logger.Error("invalid sprite", "error", err)
		respondError(w, http.StatusInternalServerError, "invalid sprite")
		return
	}

	if !asPNG {
		respondJSON(w, http.StatusOK, s)
		return
	}

	scale := defaultPNGScale
	if v := r.URL.Query().Get("scale"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			respondError(w, http.StatusBadRequest, "scale must be an integer")
			return
		}
		scale = sprites.ClampScale(n)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if err := s.WritePNG(w, scale); err != nil {
		h.logger.Error("failed to encode sprite", "error", err)
	}
}

package handlers

import (
	"encoding/json"
	"net/http"
)

// GetConfig returns feature flags and screen block layouts.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	flags, err := h.queries.ListFeatureFlags(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load config")
		return
	}

	features := make(map[string]bool, len(flags))
	for _, f := range flags {
		features[f.Key] = f.Enabled == 1
	}

	blocks, err := h.queries.ListAllEnabledScreenBlocks(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load config")
		return
	}

	screens := make(map[string][]ScreenBlock)
	for _, b := range blocks {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(b.Config), &config); err != nil {
			config = make(map[string]interface{})
		}

		screens[b.Screen] = append(screens[b.Screen], ScreenBlock{
			Type:   b.Type,
			Key:    b.Key,
			Config: config,
		})
	}

	w.Header().Set("Cache-Control", "public, max-age=60")
	respondJSON(w, http.StatusOK, ConfigResponse{
		Features: features,
		Screens:  screens,
	})
}

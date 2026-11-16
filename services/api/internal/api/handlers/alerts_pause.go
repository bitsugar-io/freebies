package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/retr0h/freebie/services/api/internal/db"
)

// UpdateAlertsPause sets or clears the time until which push alerts are paused.
// Body: {"until": "<RFC3339>" | null}. Responds {"pausedUntil": "<RFC3339>" | null}.
func (h *Handler) UpdateAlertsPause(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Until *string `json:"until"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var until sql.NullTime
	var out *string
	if req.Until != nil {
		t, err := time.Parse(time.RFC3339, *req.Until)
		if err != nil {
			respondError(w, http.StatusBadRequest, "until must be RFC3339 or null")
			return
		}
		t = t.UTC()
		until = sql.NullTime{Time: t, Valid: true}
		s := t.Format(time.RFC3339)
		out = &s
	}

	if _, err := h.queries.GetUser(r.Context(), id); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		h.logger.Error("failed to get user", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to update alerts pause")
		return
	}

	if err := h.queries.UpdateUserAlertsPause(r.Context(), db.UpdateUserAlertsPauseParams{
		AlertsPausedUntil: until,
		ID:                id,
	}); err != nil {
		h.logger.Error("failed to update alerts pause", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to update alerts pause")
		return
	}
	respondJSON(w, http.StatusOK, map[string]*string{"pausedUntil": out})
}

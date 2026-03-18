package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/retr0h/freebie/services/api/internal/db"
)

type Handler struct {
	queries *db.Queries
	logger  *slog.Logger
}

func New(database *sql.DB, logger *slog.Logger) *Handler {
	return &Handler{
		queries: db.New(database),
		logger:  logger,
	}
}

// Response helpers
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// ListLeagues returns all leagues
func (h *Handler) ListLeagues(w http.ResponseWriter, r *http.Request) {
	leagues, err := h.queries.ListLeagues(r.Context())
	if err != nil {
		h.logger.Error("failed to list leagues", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list leagues")
		return
	}

	response := make([]LeagueResponse, len(leagues))
	for i, l := range leagues {
		response[i] = LeagueResponse{
			ID:           l.ID,
			Name:         l.Name,
			Icon:         l.Icon,
			DisplayOrder: int(l.DisplayOrder),
		}
	}
	respondJSON(w, http.StatusOK, response)
}

func eventToResponse(e db.Event) EventResponse {
	var regionCode *string
	var regionName *string
	if e.RegionCode.Valid {
		regionCode = &e.RegionCode.String
		if name, ok := RegionNames[e.RegionCode.String]; ok {
			regionName = &name
		}
	}
	var teamColor *string
	if e.TeamColor.Valid {
		teamColor = &e.TeamColor.String
	}
	var icon *string
	if e.Icon.Valid {
		icon = &e.Icon.String
	}
	var triggerRule *string
	if e.TriggerRule.Valid {
		triggerRule = &e.TriggerRule.String
	}
	var offerUrl *string
	if e.OfferUrl.Valid {
		offerUrl = &e.OfferUrl.String
	}
	var affiliateUrl *string
	if e.AffiliateUrl.Valid {
		affiliateUrl = &e.AffiliateUrl.String
	}
	var affiliateTagline *string
	if e.AffiliateTagline.Valid {
		affiliateTagline = &e.AffiliateTagline.String
	}
	return EventResponse{
		ID:               e.ID,
		OfferID:          e.OfferID,
		TeamID:           e.TeamID,
		TeamName:         e.TeamName,
		League:           e.League,
		TeamColor:        teamColor,
		Icon:             icon,
		PartnerName:      e.PartnerName,
		OfferName:        e.OfferName,
		OfferDescription: e.OfferDescription,
		TriggerCondition: e.TriggerCondition,
		TriggerRule:      triggerRule,
		RegionCode:       regionCode,
		RegionName:       regionName,
		OfferUrl:         offerUrl,
		AffiliateUrl:     affiliateUrl,
		AffiliateTagline: affiliateTagline,
		IsActive:         e.IsActive == 1,
	}
}

// ListEvents returns all active events
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.queries.ListActiveEvents(r.Context())
	if err != nil {
		h.logger.Error("failed to list events", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list events")
		return
	}

	response := make([]EventResponse, len(events))
	for i, e := range events {
		response[i] = eventToResponse(e)
	}
	respondJSON(w, http.StatusOK, response)
}

// GetEvent returns a single event
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	event, err := h.queries.GetEvent(r.Context(), id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "event not found")
		return
	}
	if err != nil {
		h.logger.Error("failed to get event", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to get event")
		return
	}
	respondJSON(w, http.StatusOK, eventToResponse(event))
}

func userToResponse(u db.User, includeToken bool) UserResponse {
	var pushToken *string
	if u.PushToken.Valid {
		pushToken = &u.PushToken.String
	}
	var token *string
	if includeToken && u.Token.Valid {
		token = &u.Token.String
	}
	return UserResponse{
		ID:        u.ID,
		DeviceID:  u.DeviceID,
		Platform:  u.Platform,
		PushToken: pushToken,
		Token:     token,
	}
}

// CreateUser creates or returns existing user by device ID
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DeviceID == "" {
		respondError(w, http.StatusBadRequest, "deviceId is required")
		return
	}

	if req.Platform == "" {
		req.Platform = "unknown"
	}

	// Check if user exists
	existing, err := h.queries.GetUserByDeviceID(r.Context(), req.DeviceID)
	if err == nil {
		// User exists, return it with token (so they can recover if they lost it)
		respondJSON(w, http.StatusOK, userToResponse(existing, true))
		return
	}

	// Generate auth token
	token := uuid.New().String() + uuid.New().String()

	// Create new user
	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		ID:        uuid.New().String(),
		DeviceID:  req.DeviceID,
		Platform:  req.Platform,
		PushToken: sql.NullString{String: req.PushToken, Valid: req.PushToken != ""},
		Token:     sql.NullString{String: token, Valid: true},
	})
	if err != nil {
		h.logger.Error("failed to create user", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, userToResponse(user, true))
}

// GetUser returns a user by ID
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.queries.GetUser(r.Context(), id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		h.logger.Error("failed to get user", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	respondJSON(w, http.StatusOK, userToResponse(user, false))
}

// UpdatePushToken updates a user's push token
func (h *Handler) UpdatePushToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		PushToken string `json:"pushToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.queries.UpdateUserPushToken(r.Context(), db.UpdateUserPushTokenParams{
		PushToken: sql.NullString{String: req.PushToken, Valid: req.PushToken != ""},
		ID:        id,
	})
	if err != nil {
		h.logger.Error("failed to update push token", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to update push token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserStats returns statistics for a user
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Get deals claimed count
	claimedCount, err := h.queries.CountUserClaimedDeals(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to count claimed deals", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to get user stats")
		return
	}

	// Get subscriptions count
	subsCount, err := h.queries.CountUserSubscriptions(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to count subscriptions", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to get user stats")
		return
	}

	respondJSON(w, http.StatusOK, UserStatsResponse{
		DealsClaimed:       claimedCount,
		SubscriptionsCount: subsCount,
	})
}

// ListSubscriptions returns a user's subscriptions with event details
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	rows, err := h.queries.ListUserSubscriptions(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list subscriptions", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}

	response := make([]SubscriptionResponse, len(rows))
	for i, row := range rows {
		var regionCode *string
		if row.RegionCode.Valid {
			regionCode = &row.RegionCode.String
		}
		var teamColor *string
		if row.TeamColor.Valid {
			teamColor = &row.TeamColor.String
		}
		var icon *string
		if row.Icon.Valid {
			icon = &row.Icon.String
		}
		var triggerRule *string
		if row.TriggerRule.Valid {
			triggerRule = &row.TriggerRule.String
		}
		response[i] = SubscriptionResponse{
			ID:      row.ID,
			UserID:  row.UserID,
			EventID: row.EventID,
			Event: EventResponse{
				ID:               row.ID_2,
				OfferID:          row.OfferID,
				TeamID:           row.TeamID,
				TeamName:         row.TeamName,
				League:           row.League,
				TeamColor:        teamColor,
				Icon:             icon,
				PartnerName:      row.PartnerName,
				OfferName:        row.OfferName,
				OfferDescription: row.OfferDescription,
				TriggerCondition: row.TriggerCondition,
				TriggerRule:      triggerRule,
				RegionCode:       regionCode,
				IsActive:         row.IsActive == 1,
			},
		}
	}
	respondJSON(w, http.StatusOK, response)
}

// CreateSubscription subscribes a user to an event
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EventID == "" {
		respondError(w, http.StatusBadRequest, "eventId is required")
		return
	}

	// Verify user exists
	_, err := h.queries.GetUser(r.Context(), userID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	// Verify event exists
	event, err := h.queries.GetEvent(r.Context(), req.EventID)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "event not found")
		return
	}

	// Check if already subscribed
	_, err = h.queries.GetSubscription(r.Context(), db.GetSubscriptionParams{
		UserID:  userID,
		EventID: req.EventID,
	})
	if err == nil {
		// Already subscribed
		respondError(w, http.StatusConflict, "already subscribed")
		return
	}

	// Create subscription
	sub, err := h.queries.CreateSubscription(r.Context(), db.CreateSubscriptionParams{
		ID:      uuid.New().String(),
		UserID:  userID,
		EventID: req.EventID,
	})
	if err != nil {
		h.logger.Error("failed to create subscription", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}

	respondJSON(w, http.StatusCreated, SubscriptionResponse{
		ID:      sub.ID,
		UserID:  sub.UserID,
		EventID: sub.EventID,
		Event:   eventToResponse(event),
	})
}

// DeleteSubscription unsubscribes a user from an event
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	eventID := chi.URLParam(r, "eventId")

	err := h.queries.DeleteSubscription(r.Context(), db.DeleteSubscriptionParams{
		UserID:  userID,
		EventID: eventID,
	})
	if err != nil {
		h.logger.Error("failed to delete subscription", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListActiveDeals returns all active triggered events for a user's subscriptions
func (h *Handler) ListActiveDeals(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	rows, err := h.queries.ListActiveTriggeredEventsForUser(r.Context(), db.ListActiveTriggeredEventsForUserParams{
		UserID:   userID,
		UserID_2: userID,
	})
	if err != nil {
		h.logger.Error("failed to list active deals", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list active deals")
		return
	}

	response := make([]ActiveDealResponse, len(rows))
	for i, row := range rows {
		var expiresAt *string
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
			expiresAt = &t
		}

		var regionCode *string
		var regionName *string
		if row.RegionCode.Valid {
			regionCode = &row.RegionCode.String
			if name, ok := RegionNames[row.RegionCode.String]; ok {
				regionName = &name
			}
		}

		var teamColor *string
		if row.TeamColor.Valid {
			teamColor = &row.TeamColor.String
		}

		var icon *string
		if row.Icon.Valid {
			icon = &row.Icon.String
		}

		var triggerRule *string
		if row.TriggerRule.Valid {
			triggerRule = &row.TriggerRule.String
		}

		var offerUrl *string
		if row.OfferUrl.Valid {
			offerUrl = &row.OfferUrl.String
		}

		var affiliateUrl *string
		if row.AffiliateUrl.Valid {
			affiliateUrl = &row.AffiliateUrl.String
		}

		var affiliateTagline *string
		if row.AffiliateTagline.Valid {
			affiliateTagline = &row.AffiliateTagline.String
		}

		var dismissalType *string
		if row.DismissalType.Valid {
			dismissalType = &row.DismissalType.String
		}

		response[i] = ActiveDealResponse{
			ID:          row.ID,
			EventID:     row.EventID,
			TriggeredAt: row.TriggeredAt.Format("2006-01-02T15:04:05Z07:00"),
			ExpiresAt:   expiresAt,
			Event: EventResponse{
				ID:               row.ID_2,
				OfferID:          row.OfferID,
				TeamID:           row.TeamID,
				TeamName:         row.TeamName,
				League:           row.League,
				TeamColor:        teamColor,
				Icon:             icon,
				PartnerName:      row.PartnerName,
				OfferName:        row.OfferName,
				OfferDescription: row.OfferDescription,
				TriggerCondition: row.TriggerCondition,
				TriggerRule:      triggerRule,
				RegionCode:       regionCode,
				RegionName:       regionName,
				OfferUrl:         offerUrl,
				AffiliateUrl:     affiliateUrl,
				AffiliateTagline: affiliateTagline,
				IsActive:         row.IsActive == 1,
			},
			IsDismissed:   row.IsDismissed == 1,
			DismissalType: dismissalType,
		}
	}
	respondJSON(w, http.StatusOK, response)
}

// CreateDismissal marks a deal as dismissed/acknowledged
func (h *Handler) CreateDismissal(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	var req CreateDismissalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TriggeredEventID == "" {
		respondError(w, http.StatusBadRequest, "triggeredEventId is required")
		return
	}

	if req.Type == "" {
		req.Type = "got_it"
	}

	if req.Type != "got_it" && req.Type != "stop_reminding" {
		respondError(w, http.StatusBadRequest, "type must be 'got_it' or 'stop_reminding'")
		return
	}

	// Check if already dismissed
	_, err := h.queries.GetDismissal(r.Context(), db.GetDismissalParams{
		UserID:           userID,
		TriggeredEventID: req.TriggeredEventID,
	})
	if err == nil {
		respondError(w, http.StatusConflict, "already dismissed")
		return
	}

	dismissal, err := h.queries.CreateDismissal(r.Context(), db.CreateDismissalParams{
		ID:               uuid.New().String(),
		UserID:           userID,
		TriggeredEventID: req.TriggeredEventID,
		Type:             req.Type,
	})
	if err != nil {
		h.logger.Error("failed to create dismissal", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create dismissal")
		return
	}

	respondJSON(w, http.StatusCreated, DismissalResponse{
		ID:               dismissal.ID,
		UserID:           dismissal.UserID,
		TriggeredEventID: dismissal.TriggeredEventID,
		Type:             dismissal.Type,
		DismissedAt:      dismissal.DismissedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteDismissal removes a dismissal (undo)
func (h *Handler) DeleteDismissal(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	triggeredEventID := chi.URLParam(r, "triggeredEventId")

	err := h.queries.DeleteDismissal(r.Context(), db.DeleteDismissalParams{
		UserID:           userID,
		TriggeredEventID: triggeredEventID,
	})
	if err != nil {
		h.logger.Error("failed to delete dismissal", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to delete dismissal")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

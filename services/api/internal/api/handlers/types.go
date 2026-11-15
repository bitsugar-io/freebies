package handlers

import "github.com/retr0h/freebie/services/api/internal/sprites"

// LeagueResponse represents a league in API responses
type LeagueResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	DisplayOrder int    `json:"displayOrder"`
}

// EventResponse represents an event/deal in API responses
type EventResponse struct {
	ID               string  `json:"id"`
	OfferID          string  `json:"offerId"`
	TeamID           string  `json:"teamId"`
	TeamName         string  `json:"teamName"`
	League           string  `json:"league"`
	TeamColor        *string `json:"teamColor,omitempty"`
	Icon             *string `json:"icon,omitempty"`
	PartnerName      string  `json:"partnerName"`
	OfferName        string  `json:"offerName"`
	OfferDescription string  `json:"offerDescription"`
	TriggerCondition string  `json:"triggerCondition"`
	TriggerRule      *string `json:"triggerRule,omitempty"`
	RegionCode       *string `json:"regionCode,omitempty"`
	RegionName       *string `json:"regionName,omitempty"`
	OfferUrl         *string `json:"offerUrl,omitempty"`
	AffiliateUrl     *string `json:"affiliateUrl,omitempty"`
	AffiliateTagline *string `json:"affiliateTagline,omitempty"`
	IsActive         bool    `json:"isActive"`
	SpriteID         *string `json:"spriteId,omitempty"`
	// Sprite is the pixel art for the deal; the "gift" fallback when unset.
	Sprite *sprites.Sprite `json:"sprite,omitempty"`
}

// RegionNames maps region codes to human-readable names. The product is
// LA-only; unknown codes (legacy events in other regions) simply have no name
// and callers must treat a missing entry as "no region label".
var RegionNames = map[string]string{
	"us-ca-la": "Los Angeles",
	"us-ca-oc": "Orange County",
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
	DeviceID  string `json:"deviceId"`
	Platform  string `json:"platform"`
	PushToken string `json:"pushToken,omitempty"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        string  `json:"id"`
	DeviceID  string  `json:"deviceId"`
	Platform  string  `json:"platform"`
	PushToken *string `json:"pushToken,omitempty"`
	Token     *string `json:"token,omitempty"`
	// AlertsPausedUntil is RFC3339; omitted when alerts are not paused.
	AlertsPausedUntil *string `json:"alertsPausedUntil,omitempty"`
}

// UserStatsResponse contains user statistics
type UserStatsResponse struct {
	DealsClaimed       int64 `json:"dealsClaimed"`
	SubscriptionsCount int64 `json:"subscriptionsCount"`
}

// SubscriptionResponse represents a subscription in API responses
type SubscriptionResponse struct {
	ID      string        `json:"id"`
	UserID  string        `json:"userId"`
	EventID string        `json:"eventId"`
	Event   EventResponse `json:"event"`
}

// CreateSubscriptionRequest is the request body for creating a subscription
type CreateSubscriptionRequest struct {
	EventID string `json:"eventId"`
}

// ActiveDealResponse represents an active triggered deal
type ActiveDealResponse struct {
	ID          string  `json:"id"`
	EventID     string  `json:"eventId"`
	TriggeredAt string  `json:"triggeredAt"`
	ExpiresAt   *string `json:"expiresAt,omitempty"`
	// AvailableAt is when a delayed deal's code unlocks ("Code unlocks Tue");
	// omitted when the deal is usable as soon as it triggers.
	AvailableAt   *string       `json:"availableAt,omitempty"`
	Event         EventResponse `json:"event"`
	IsDismissed   bool          `json:"isDismissed"`
	DismissalType *string       `json:"dismissalType,omitempty"`
}

// ClaimedDealResponse is one deal the user claimed, kept after its window closes.
type ClaimedDealResponse struct {
	// ID is the triggered event id, so it opens the same deal screen as an active deal.
	ID          string `json:"id"`
	EventID     string `json:"eventId"`
	ClaimedAt   string `json:"claimedAt"`
	TriggeredAt string `json:"triggeredAt"`
	// AvailableAt is when a delayed deal's code unlocked; omitted otherwise.
	AvailableAt *string       `json:"availableAt,omitempty"`
	Event       EventResponse `json:"event"`
}

// CreateDismissalRequest is the request body for dismissing a deal
type CreateDismissalRequest struct {
	TriggeredEventID string `json:"triggeredEventId"`
	Type             string `json:"type"` // "got_it" or "stop_reminding"
}

// DismissalResponse represents a dismissal in API responses
type DismissalResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	TriggeredEventID string `json:"triggeredEventId"`
	Type             string `json:"type"`
	DismissedAt      string `json:"dismissedAt"`
}

// ConfigResponse is the top-level config response
type ConfigResponse struct {
	Features map[string]bool          `json:"features"`
	Screens  map[string][]ScreenBlock `json:"screens"`
}

// ScreenBlock represents a UI block in a screen layout
type ScreenBlock struct {
	Type   string                 `json:"type"`
	Key    string                 `json:"key"`
	Config map[string]interface{} `json:"config"`
}

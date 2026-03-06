package models

// User represents a user subscribed to notifications
type User struct {
	UserID     string `yaml:"userId"`
	PhoneE164  string `yaml:"phoneE164"`
	RegionCode string `yaml:"regionCode"`
	TeamID     string `yaml:"teamId"` // MLB team they care about
}

// Offer represents a partner offer
type Offer struct {
	OfferID     string `yaml:"offerId"`
	PartnerName string `yaml:"partnerName"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Rule represents a trigger rule for offers
type Rule struct {
	RuleID    string                 `yaml:"ruleId"`
	OfferID   string                 `yaml:"offerId"`
	TeamID    string                 `yaml:"teamId"`    // Which MLB team
	Condition map[string]interface{} `yaml:"condition"` // Rule conditions (e.g., win: true, strikeouts: {gte: 7})
}

// Event represents an MLB game event
type Event struct {
	EventID    string                 `json:"eventId"`
	EventType  string                 `json:"eventType"` // Always "mlb_game" for now
	RegionCode string                 `json:"regionCode"`
	DateISO    string                 `json:"dateISO"` // YYYY-MM-DD
	Data       map[string]interface{} `json:"data"`    // Game data (team_id, win, strikeouts, players, etc.)
}

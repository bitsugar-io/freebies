package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const expoPushURL = "https://exp.host/--/api/v2/push/send"

// ExpoPushMessage represents a push notification to send via Expo
type ExpoPushMessage struct {
	To       string                 `json:"to"`
	Title    string                 `json:"title,omitempty"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Sound    string                 `json:"sound,omitempty"`
	Badge    *int                   `json:"badge,omitempty"`
	TTL      int                    `json:"ttl,omitempty"`
	Priority string                 `json:"priority,omitempty"` // default, normal, high
}

// ExpoPushResponse represents the response from Expo push API
type ExpoPushResponse struct {
	Data []ExpoPushTicket `json:"data"`
}

type ExpoPushTicket struct {
	ID      string `json:"id,omitempty"`
	Status  string `json:"status"` // "ok" or "error"
	Message string `json:"message,omitempty"`
	Details struct {
		Error string `json:"error,omitempty"`
	} `json:"details,omitempty"`
}

// ExpoNotifier sends push notifications via Expo's push service
type ExpoNotifier struct {
	client *http.Client
}

// NewExpoNotifier creates a new Expo push notifier
func NewExpoNotifier() *ExpoNotifier {
	return &ExpoNotifier{
		client: &http.Client{},
	}
}

// Send sends a push notification to the specified Expo push token
func (n *ExpoNotifier) Send(token, title, body string, data map[string]interface{}) (*ExpoPushTicket, error) {
	msg := ExpoPushMessage{
		To:       token,
		Title:    title,
		Body:     body,
		Data:     data,
		Sound:    "default",
		Priority: "high",
	}

	return n.sendMessage(msg)
}

// SendBatch sends multiple push notifications at once
func (n *ExpoNotifier) SendBatch(messages []ExpoPushMessage) ([]ExpoPushTicket, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("marshaling messages: %w", err)
	}

	req, err := http.NewRequest("POST", expoPushURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var pushResp ExpoPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pushResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return pushResp.Data, nil
}

func (n *ExpoNotifier) sendMessage(msg ExpoPushMessage) (*ExpoPushTicket, error) {
	tickets, err := n.SendBatch([]ExpoPushMessage{msg})
	if err != nil {
		return nil, err
	}
	if len(tickets) == 0 {
		return nil, fmt.Errorf("no ticket returned")
	}
	return &tickets[0], nil
}

// IsValidExpoToken checks if a token looks like a valid Expo push token
func IsValidExpoToken(token string) bool {
	// Expo push tokens start with "ExponentPushToken[" or "ExpoPushToken["
	return len(token) > 20 &&
		(token[:18] == "ExponentPushToken[" || token[:14] == "ExpoPushToken[")
}

package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/retr0h/freebie/services/api/internal/httputil"
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
	url    string
}

// NewExpoNotifier creates a new Expo push notifier
func NewExpoNotifier() *ExpoNotifier {
	return &ExpoNotifier{client: &http.Client{}, url: expoPushURL}
}

// NewExpoNotifierWithURL creates a new Expo push notifier with a custom URL
func NewExpoNotifierWithURL(url string) *ExpoNotifier {
	return &ExpoNotifier{client: &http.Client{}, url: url}
}

// Send sends a push notification to the specified Expo push token
func (n *ExpoNotifier) Send(ctx context.Context, token, title, body string, data map[string]interface{}) (*ExpoPushTicket, error) {
	msg := ExpoPushMessage{
		To:       token,
		Title:    title,
		Body:     body,
		Data:     data,
		Sound:    "default",
		Priority: "high",
	}

	return n.sendMessage(ctx, msg)
}

// SendBatch sends multiple push notifications at once
func (n *ExpoNotifier) SendBatch(ctx context.Context, messages []ExpoPushMessage) ([]ExpoPushTicket, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	payload, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("marshaling messages: %w", err)
	}

	newReq := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", n.url, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		return req, nil
	}

	resp, err := httputil.Do(ctx, n.client, newReq, nil)
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

func (n *ExpoNotifier) sendMessage(ctx context.Context, msg ExpoPushMessage) (*ExpoPushTicket, error) {
	tickets, err := n.SendBatch(ctx, []ExpoPushMessage{msg})
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

const (
	// MaxBatchSize is the maximum number of notifications per Expo API request
	MaxBatchSize = 100
	// DefaultWorkers is the default number of concurrent workers for sending
	DefaultWorkers = 10
)

// BatchResult contains the results of a batch send operation
type BatchResult struct {
	Sent   int64
	Failed int64
	Errors []error
}

// SendBatchConcurrent sends notifications in batches using multiple workers
func (n *ExpoNotifier) SendBatchConcurrent(ctx context.Context, logger *slog.Logger, messages []ExpoPushMessage, workers int) *BatchResult {
	if workers <= 0 {
		workers = DefaultWorkers
	}

	result := &BatchResult{}
	if len(messages) == 0 {
		return result
	}

	// Split messages into batches
	var batches [][]ExpoPushMessage
	for i := 0; i < len(messages); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(messages) {
			end = len(messages)
		}
		batches = append(batches, messages[i:end])
	}

	logger.Info("sending notifications",
		"total", len(messages),
		"batches", len(batches),
		"workers", workers,
	)

	// Create work channel
	batchChan := make(chan []ExpoPushMessage, len(batches))
	for _, batch := range batches {
		batchChan <- batch
	}
	close(batchChan)

	// Collect errors
	var errorsMu sync.Mutex
	var errors []error

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				tickets, err := n.SendBatch(ctx, batch)
				if err != nil {
					atomic.AddInt64(&result.Failed, int64(len(batch)))
					errorsMu.Lock()
					errors = append(errors, err)
					errorsMu.Unlock()
					continue
				}

				// Count successes and failures
				for _, ticket := range tickets {
					if ticket.Status == "ok" {
						atomic.AddInt64(&result.Sent, 1)
					} else {
						atomic.AddInt64(&result.Failed, 1)
					}
				}
			}
		}()
	}

	wg.Wait()
	result.Errors = errors
	return result
}

// DeduplicateMessages removes duplicate messages based on the To field,
// keeping only the first occurrence for each recipient.
func DeduplicateMessages(messages []ExpoPushMessage) []ExpoPushMessage {
	seen := make(map[string]bool, len(messages))
	result := make([]ExpoPushMessage, 0, len(messages))
	for _, msg := range messages {
		if !seen[msg.To] {
			seen[msg.To] = true
			result = append(result, msg)
		}
	}
	return result
}

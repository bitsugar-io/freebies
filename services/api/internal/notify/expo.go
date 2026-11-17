package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

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

// ErrDeviceNotRegistered is the Expo ticket error for a token whose app was
// uninstalled or whose push permission was revoked. Stop sending to it.
const ErrDeviceNotRegistered = "DeviceNotRegistered"

// ErrMaybeDelivered marks a send whose outcome is unknown: the request reached
// Expo but no usable answer came back (a timeout or reset after the request
// was written, a 5xx other than 503, an unreadable 200). Expo may have queued
// the pushes, so they are not sent again.
var ErrMaybeDelivered = httputil.ErrMaybeDelivered

// sendTimeout bounds one Expo request so a hung connection cannot stall a run.
const sendTimeout = 30 * time.Second

// ExpoNotifier sends push notifications via Expo's push service
type ExpoNotifier struct {
	client *http.Client
	url    string
	retry  httputil.RetryOptions
}

// Expo sends are not idempotent: a POST Expo received is a push on a phone.
// Retry only what Expo cannot have processed (see httputil.RetryOptions).
func newExpoNotifier(url string) *ExpoNotifier {
	return &ExpoNotifier{
		client: &http.Client{Timeout: sendTimeout},
		url:    url,
		retry:  httputil.RetryOptions{NonIdempotent: true},
	}
}

// NewExpoNotifier creates a new Expo push notifier
func NewExpoNotifier() *ExpoNotifier { return newExpoNotifier(expoPushURL) }

// NewExpoNotifierWithURL creates a new Expo push notifier with a custom URL
func NewExpoNotifierWithURL(url string) *ExpoNotifier { return newExpoNotifier(url) }

// SetRetryDelay sets the base backoff between retries (tests use a short one).
func (n *ExpoNotifier) SetRetryDelay(d time.Duration) { n.retry.BaseDelay = d }

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

// SendBatch sends up to MaxBatchSize push notifications in one request. Expo
// returns one ticket per message, in message order. An error wrapping
// ErrMaybeDelivered means Expo may have accepted the batch; any other error
// means it did not.
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

	resp, err := httputil.Do(ctx, n.client, newReq, &n.retry)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode >= 500 && resp.StatusCode != http.StatusServiceUnavailable:
		// A 500, 502 or 504 does not say whether Expo queued the pushes.
		return nil, fmt.Errorf("%w: unexpected status: %d", ErrMaybeDelivered, resp.StatusCode)
	default:
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var pushResp ExpoPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pushResp); err != nil {
		return nil, fmt.Errorf("%w: decoding response: %w", ErrMaybeDelivered, err)
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

// Delivery is what became of one message. The values are the statuses the
// worker writes to notifications.status.
type Delivery string

const (
	// Delivered: Expo returned an "ok" ticket for the message.
	Delivered Delivery = "sent"
	// NotDelivered: Expo rejected it (an error ticket) or never received it.
	// Safe to send again.
	NotDelivered Delivery = "failed"
	// MaybeDelivered: the outcome is unknown (see ErrMaybeDelivered). Treat it
	// as sent: a missed push is better than a duplicate one.
	MaybeDelivered Delivery = "unknown"
)

// MessageResult is the outcome of one message.
type MessageResult struct {
	To       string
	Delivery Delivery
	// ErrorCode is Expo's details.error on an error ticket, e.g.
	// ErrDeviceNotRegistered; empty otherwise.
	ErrorCode string
}

// BatchResult contains the results of a batch send operation
type BatchResult struct {
	Sent    int64
	Failed  int64
	Unknown int64
	Errors  []error
	// Results has one entry per input message, in input order.
	Results []MessageResult
}

// SendBatchConcurrent sends notifications in batches using multiple workers.
// Results[i] is the outcome of messages[i]: Expo answers each batch with its
// tickets in message order, so ticket j of a batch belongs to message j of
// that batch.
func (n *ExpoNotifier) SendBatchConcurrent(ctx context.Context, logger *slog.Logger, messages []ExpoPushMessage, workers int) *BatchResult {
	if workers <= 0 {
		workers = DefaultWorkers
	}

	result := &BatchResult{Results: make([]MessageResult, len(messages))}
	if len(messages) == 0 {
		return result
	}
	// A batch that is never attempted (context cancelled) was not delivered.
	for i, m := range messages {
		result.Results[i] = MessageResult{To: m.To, Delivery: NotDelivered}
	}

	// Split messages into batches, remembering where each starts.
	type batch struct {
		start int
		msgs  []ExpoPushMessage
	}
	var batches []batch
	for i := 0; i < len(messages); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(messages))
		batches = append(batches, batch{start: i, msgs: messages[i:end]})
	}

	logger.Info("sending notifications",
		"total", len(messages),
		"batches", len(batches),
		"workers", workers,
	)

	batchChan := make(chan batch, len(batches))
	for _, b := range batches {
		batchChan <- b
	}
	close(batchChan)

	var errorsMu sync.Mutex
	var errs []error

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for b := range batchChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				tickets, err := n.SendBatch(ctx, b.msgs)
				if err == nil && len(tickets) != len(b.msgs) {
					// Without one ticket per message we cannot tell whose
					// push went out.
					err = fmt.Errorf("%w: %d tickets for %d messages", ErrMaybeDelivered, len(tickets), len(b.msgs))
				}
				if err != nil {
					d := NotDelivered
					if errors.Is(err, ErrMaybeDelivered) {
						d = MaybeDelivered
					}
					for j := range b.msgs {
						result.Results[b.start+j].Delivery = d
					}
					if d == MaybeDelivered {
						atomic.AddInt64(&result.Unknown, int64(len(b.msgs)))
					} else {
						atomic.AddInt64(&result.Failed, int64(len(b.msgs)))
					}
					errorsMu.Lock()
					errs = append(errs, err)
					errorsMu.Unlock()
					continue
				}

				for j, ticket := range tickets {
					r := &result.Results[b.start+j]
					if ticket.Status == "ok" {
						r.Delivery = Delivered
						atomic.AddInt64(&result.Sent, 1)
					} else {
						r.Delivery = NotDelivered
						r.ErrorCode = ticket.Details.Error
						atomic.AddInt64(&result.Failed, 1)
					}
				}
			}
		}()
	}

	wg.Wait()
	result.Errors = errs
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

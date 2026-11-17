package notify_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/notify"
)

func TestSendBatch_RetriesOnServerError(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"ticket-1","status":"ok"}]}`))
	}))
	defer srv.Close()

	notifier := notify.NewExpoNotifierWithURL(srv.URL)
	notifier.SetRetryDelay(time.Millisecond)
	messages := []notify.ExpoPushMessage{{To: "ExponentPushToken[test]", Body: "test"}}

	tickets, err := notifier.SendBatch(context.Background(), messages)
	require.NoError(t, err)
	assert.Len(t, tickets, 1)
	assert.Equal(t, "ok", tickets[0].Status)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&attempts), int32(3))
}

func TestDeduplicateMessages(t *testing.T) {
	messages := []notify.ExpoPushMessage{
		{To: "ExponentPushToken[aaa]", Body: "hello"},
		{To: "ExponentPushToken[bbb]", Body: "hello"},
		{To: "ExponentPushToken[aaa]", Body: "hello"}, // duplicate
	}

	deduped := notify.DeduplicateMessages(messages)
	assert.Len(t, deduped, 2)

	tokens := make(map[string]bool)
	for _, m := range deduped {
		tokens[m.To] = true
	}
	assert.True(t, tokens["ExponentPushToken[aaa]"])
	assert.True(t, tokens["ExponentPushToken[bbb]"])
}

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

// Each result belongs to its own message, across batches: Expo's tickets come
// back in message order per batch, and that order is kept.
func TestSendBatchConcurrent_ResultPerMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch []notify.ExpoPushMessage
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &batch)
		tickets := make([]notify.ExpoPushTicket, len(batch))
		for i, m := range batch {
			tickets[i].Status = "ok"
			if m.Body == "dead" {
				tickets[i].Status = "error"
				tickets[i].Details.Error = notify.ErrDeviceNotRegistered
			}
		}
		_ = json.NewEncoder(w).Encode(notify.ExpoPushResponse{Data: tickets})
	}))
	defer srv.Close()

	var messages []notify.ExpoPushMessage
	for i := range 250 { // three batches
		body := "live"
		if i%7 == 3 {
			body = "dead"
		}
		messages = append(messages, notify.ExpoPushMessage{To: fmt.Sprintf("tok-%d", i), Body: body})
	}
	res := notify.NewExpoNotifierWithURL(srv.URL).SendBatchConcurrent(context.Background(), discard, messages, 3)
	require.Len(t, res.Results, len(messages))
	var dead int64
	for i, m := range messages {
		r := res.Results[i]
		assert.Equal(t, m.To, r.To)
		if m.Body == "dead" {
			dead++
			assert.Equal(t, notify.NotDelivered, r.Delivery, m.To)
			assert.Equal(t, notify.ErrDeviceNotRegistered, r.ErrorCode, m.To)
		} else {
			assert.Equal(t, notify.Delivered, r.Delivery, m.To)
		}
	}
	assert.Equal(t, dead, res.Failed)
	assert.Equal(t, int64(len(messages))-dead, res.Sent)
}

// Expo read the request but the response was lost (the connection dropped).
// The push may be on the phone already, so it is not sent again and its
// outcome is unknown, not failed.
func TestSendBatchConcurrent_NoResendAfterLostResponse(t *testing.T) {
	var posts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		_, _ = io.ReadAll(r.Body)
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			_ = conn.Close()
		}
	}))
	defer srv.Close()

	n := notify.NewExpoNotifierWithURL(srv.URL)
	n.SetRetryDelay(time.Millisecond)
	res := n.SendBatchConcurrent(context.Background(), discard,
		[]notify.ExpoPushMessage{{To: "a"}, {To: "b"}}, 1)
	assert.Equal(t, int32(1), posts.Load(), "a request Expo may have processed must not be re-sent")
	assert.Equal(t, int64(2), res.Unknown)
	assert.Equal(t, int64(0), res.Failed)
	for _, r := range res.Results {
		assert.Equal(t, notify.MaybeDelivered, r.Delivery)
	}
}

// A 500 does not say whether Expo queued the pushes: no retry, outcome unknown.
// A 503 that never clears means Expo did not take them: failed, safe to retry later.
func TestSendBatchConcurrent_StatusOutcomes(t *testing.T) {
	for _, tc := range []struct {
		status    int
		wantPosts int32
		want      notify.Delivery
	}{
		{http.StatusInternalServerError, 1, notify.MaybeDelivered},
		{http.StatusBadGateway, 1, notify.MaybeDelivered},
		{http.StatusServiceUnavailable, 4, notify.NotDelivered},
		{http.StatusTooManyRequests, 4, notify.NotDelivered},
		{http.StatusBadRequest, 1, notify.NotDelivered},
	} {
		var posts atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			posts.Add(1)
			w.WriteHeader(tc.status)
		}))
		n := notify.NewExpoNotifierWithURL(srv.URL)
		n.SetRetryDelay(time.Millisecond)
		res := n.SendBatchConcurrent(context.Background(), discard, []notify.ExpoPushMessage{{To: "a"}}, 1)
		srv.Close()
		assert.Equal(t, tc.wantPosts, posts.Load(), "status %d posts", tc.status)
		assert.Equal(t, tc.want, res.Results[0].Delivery, "status %d", tc.status)
	}
}

package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/notify"
)

func TestSendBatch_RetriesOnServerError(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"ticket-1","status":"ok"}]}`))
	}))
	defer srv.Close()

	notifier := notify.NewExpoNotifierWithURL(srv.URL)
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

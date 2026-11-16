package espn_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// A slow scoreboard must not block a different URL, and concurrent callers of
// the same URL share one request.
func TestClientFetchesDifferentURLsConcurrently(t *testing.T) {
	release := make(chan struct{})
	var slowHits, fastHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("dates") == "20261114" {
			slowHits.Add(1)
			<-release
		} else {
			fastHits.Add(1)
		}
		_, _ = w.Write([]byte(`{"events":[]}`))
	}))
	t.Cleanup(srv.Close)
	c := espn.NewClientWith(srv.URL, srv.Client())
	feed := espn.Feed{Path: "basketball/nba"}

	var wg sync.WaitGroup
	for range 4 { // same slow URL
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Games(t.Context(), feed, day(t, "2026-11-14"))
			assert.NoError(t, err)
		}()
	}
	// Wait for the slow request to be in flight, then fetch another day.
	require.Eventually(t, func() bool { return slowHits.Load() == 1 }, 2*time.Second, 5*time.Millisecond)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := c.Games(t.Context(), feed, day(t, "2026-11-20"))
		assert.NoError(t, err)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("a different URL blocked behind the slow one")
	}
	close(release)
	wg.Wait()
	assert.EqualValues(t, 1, slowHits.Load(), "same URL must share one request")
}

package httputil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := &RetryOptions{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	resp, err := Do(context.Background(), srv.Client(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	}, opts)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
	assert.Equal(t, int32(1), attempts.Load())
}

func TestDo_RetriesOn500(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := &RetryOptions{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	resp, err := Do(context.Background(), srv.Client(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	}, opts)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
	assert.Equal(t, int32(3), attempts.Load())
}

func TestDo_RetriesOn429(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := &RetryOptions{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	resp, err := Do(context.Background(), srv.Client(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	}, opts)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
	assert.Equal(t, int32(2), attempts.Load())
}

func TestDo_NoRetryOn4xx(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	opts := &RetryOptions{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	resp, err := Do(context.Background(), srv.Client(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	}, opts)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()
	assert.Equal(t, int32(1), attempts.Load())
}

func TestDo_ExhaustsRetries(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	opts := &RetryOptions{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	resp, err := Do(context.Background(), srv.Client(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	}, opts)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	_ = resp.Body.Close()
	assert.Equal(t, int32(4), attempts.Load())
}

func TestDo_RespectsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &RetryOptions{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	_, err := Do(ctx, srv.Client(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	}, opts)

	require.Error(t, err)
}

// roundTripFunc lets a test stand in for the network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// A non-idempotent request that failed before it was written (connection
// refused) is retried; one that failed after it was written is not, and the
// error says it may have been delivered.
func TestDo_NonIdempotentRetriesOnlyUnsentRequests(t *testing.T) {
	newReq := func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, "http://expo.invalid/push", strings.NewReader("{}"))
	}
	opts := &RetryOptions{MaxRetries: 3, BaseDelay: time.Millisecond, NonIdempotent: true}

	var refused atomic.Int32
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		refused.Add(1)
		return nil, errors.New("dial tcp: connection refused")
	})}
	_, err := Do(context.Background(), client, newReq, opts)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrMaybeDelivered)
	assert.Equal(t, int32(4), refused.Load(), "connection refused before sending is retried")

	var lost atomic.Int32
	client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		lost.Add(1)
		if tr := httptrace.ContextClientTrace(r.Context()); tr != nil && tr.WroteRequest != nil {
			tr.WroteRequest(httptrace.WroteRequestInfo{})
		}
		return nil, errors.New("read tcp: i/o timeout")
	})}
	_, err = Do(context.Background(), client, newReq, opts)
	require.ErrorIs(t, err, ErrMaybeDelivered)
	assert.Equal(t, int32(1), lost.Load(), "a request that was written is not re-sent")
}

func TestDo_NonIdempotentStatuses(t *testing.T) {
	for status, want := range map[int]int32{500: 1, 502: 1, 504: 1, 503: 4, 429: 4} {
		var attempts atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts.Add(1)
			w.WriteHeader(status)
		}))
		resp, err := Do(context.Background(), srv.Client(), func() (*http.Request, error) {
			return http.NewRequest(http.MethodPost, srv.URL, strings.NewReader("{}"))
		}, &RetryOptions{MaxRetries: 3, BaseDelay: time.Millisecond, NonIdempotent: true})
		require.NoError(t, err)
		_ = resp.Body.Close()
		srv.Close()
		assert.Equal(t, want, attempts.Load(), "status %d", status)
	}
}

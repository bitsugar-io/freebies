package httputil

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryTransport_RetriesExtraStatus(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) <= 2 {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{Transport: &RetryTransport{Options: &RetryOptions{
		MaxRetries:    3,
		BaseDelay:     10 * time.Millisecond,
		RetryStatuses: []int{http.StatusNotAcceptable},
	}}}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestRetryTransport_NoRetryOn406ByDefault(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusNotAcceptable)
	}))
	defer srv.Close()

	client := &http.Client{Transport: &RetryTransport{Options: &RetryOptions{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
	}}}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode)
	assert.Equal(t, int32(1), attempts.Load())
}

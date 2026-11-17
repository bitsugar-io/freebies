package httputil

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptrace"
	"slices"
	"sync/atomic"
	"time"
)

// ErrMaybeDelivered wraps a transport error on a NonIdempotent request that
// happened after the request was written: the server may have acted on it, so
// Do does not send it again.
var ErrMaybeDelivered = errors.New("request may have been delivered")

// RetryOptions configures retry behavior.
type RetryOptions struct {
	MaxRetries int           // default 3
	BaseDelay  time.Duration // default 1s
	// RetryStatuses lists extra status codes to retry beyond the defaults
	// (5xx and 429), e.g. 406 from the MLB Stats API edge.
	RetryStatuses []int
	// NonIdempotent marks a request that must not run twice, such as a push
	// send. Do then retries only when the server cannot have acted on it: a
	// transport error before the request was fully written (connection
	// refused, DNS), 429 and 503. A transport error after the request was
	// written (timeout, reset) returns an error wrapping ErrMaybeDelivered at
	// once, and other 5xx are returned as they are, without a retry.
	NonIdempotent bool
}

func (o *RetryOptions) maxRetries() int {
	if o != nil && o.MaxRetries > 0 {
		return o.MaxRetries
	}
	return 3
}

func (o *RetryOptions) baseDelay() time.Duration {
	if o != nil && o.BaseDelay > 0 {
		return o.BaseDelay
	}
	return time.Second
}

// Do executes an HTTP request with retry and exponential backoff.
// newReq is called for each attempt to produce a fresh request (avoids consumed-body issues).
// Retries on network errors, 5xx, 429, and opts.RetryStatuses. Does not
// retry other 4xx. See RetryOptions.NonIdempotent for requests that must not
// be sent twice.
func Do(
	ctx context.Context,
	client *http.Client,
	newReq func() (*http.Request, error),
	opts *RetryOptions,
) (*http.Response, error) {
	maxRetries := opts.maxRetries()
	base := opts.baseDelay()

	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before each attempt (including the first).
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		req, err := newReq()
		if err != nil {
			return nil, err
		}
		var wrote atomic.Bool
		reqCtx := ctx
		if opts != nil && opts.NonIdempotent {
			reqCtx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
				WroteRequest: func(info httptrace.WroteRequestInfo) {
					if info.Err == nil {
						wrote.Store(true)
					}
				},
			})
		}
		req = req.WithContext(reqCtx)

		resp, err := client.Do(req)
		if err != nil {
			if wrote.Load() {
				return nil, fmt.Errorf("%w: %w", ErrMaybeDelivered, err)
			}
			lastErr = err
			lastResp = nil
			if attempt < maxRetries {
				if sleepErr := backoff(ctx, base, attempt); sleepErr != nil {
					return nil, sleepErr
				}
			}
			continue
		}

		// Non-retryable status: return immediately.
		if !opts.isRetryable(resp.StatusCode) {
			return resp, nil
		}

		// Retryable status on last attempt: return response as-is.
		if attempt == maxRetries {
			return resp, nil
		}

		// Retryable status with retries remaining: close body and retry.
		_ = resp.Body.Close()
		lastErr = nil
		lastResp = nil

		if sleepErr := backoff(ctx, base, attempt); sleepErr != nil {
			return nil, sleepErr
		}
	}

	if lastResp != nil {
		return lastResp, nil
	}
	return nil, lastErr
}

func (o *RetryOptions) isRetryable(status int) bool {
	if o != nil && o.NonIdempotent {
		// Only statuses that mean the request was not processed.
		return status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable ||
			slices.Contains(o.RetryStatuses, status)
	}
	if status == http.StatusTooManyRequests || status >= 500 {
		return true
	}
	return o != nil && slices.Contains(o.RetryStatuses, status)
}

// backoff sleeps for base * 2^attempt with +/-25% jitter, respecting context cancellation.
func backoff(ctx context.Context, base time.Duration, attempt int) error {
	delay := base * (1 << uint(attempt))

	// Apply +/-25% jitter.
	jitter := 0.75 + rand.Float64()*0.5 // [0.75, 1.25)
	delay = time.Duration(float64(delay) * jitter)

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

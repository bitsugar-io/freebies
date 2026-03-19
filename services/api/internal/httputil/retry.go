package httputil

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"
)

// RetryOptions configures retry behavior.
type RetryOptions struct {
	MaxRetries int           // default 3
	BaseDelay  time.Duration // default 1s
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
// Retries on network errors, 5xx, and 429. Does not retry other 4xx.
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
		req = req.WithContext(ctx)

		resp, err := client.Do(req)
		if err != nil {
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
		if !isRetryable(resp.StatusCode) {
			return resp, nil
		}

		// Retryable status on last attempt: return response as-is.
		if attempt == maxRetries {
			return resp, nil
		}

		// Retryable status with retries remaining: close body and retry.
		resp.Body.Close()
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

func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
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

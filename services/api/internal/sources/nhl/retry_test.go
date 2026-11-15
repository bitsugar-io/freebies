package nhl

import (
	"net/http"
	"time"

	"github.com/retr0h/freebie/services/api/internal/httputil"
)

func fastRetry() http.RoundTripper {
	return &httputil.RetryTransport{Options: &httputil.RetryOptions{
		MaxRetries: 3, BaseDelay: time.Millisecond,
	}}
}

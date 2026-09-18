package httputil

import "net/http"

// RetryTransport is an http.RoundTripper that retries requests using Do.
// Use it to add retries to clients that accept an *http.Client, such as
// third-party SDKs. Requests with a body that cannot be replayed (no
// GetBody) are sent once without retry.
type RetryTransport struct {
	Base    http.RoundTripper // defaults to http.DefaultTransport
	Options *RetryOptions
}

// RoundTrip implements http.RoundTripper.
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return base.RoundTrip(req)
	}

	client := &http.Client{Transport: base}
	return Do(req.Context(), client, func() (*http.Request, error) {
		r := req.Clone(req.Context())
		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			r.Body = body
		}
		return r, nil
	}, t.Options)
}

package client

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	mathrand "math/rand/v2"
	"net/http"
	"time"
)

// retryTransport is an http.RoundTripper that:
//   - Injects a unique X-Request-ID header on every request for traceability.
//   - Retries requests that return retryable HTTP status codes (429, 5xx) using
//     exponential backoff capped at 30 seconds.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
}

// newRequestID generates a random hex-encoded request identifier.
func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// RoundTrip executes the request, retrying on retryable status codes.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Inject X-Request-ID before the first attempt.
	if req.Header.Get("X-Request-ID") == "" {
		// Clone the header map so we don't mutate the caller's request.
		req = req.Clone(req.Context())
		req.Header.Set("X-Request-ID", newRequestID())
	}

	// Snapshot the body so it can be replayed on retries.
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("retryTransport: read request body: %w", err)
		}
		req.Body.Close()
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		// Restore body for each attempt.
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if !IsRetryable(resp.StatusCode) || attempt == t.maxRetries {
			break
		}

		// Drain and close the response body before retrying to release the
		// underlying TCP connection back to the pool.
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		delay := computeDelay(t.baseDelay, attempt)
		log.Printf("[DEBUG] retryTransport: attempt %d/%d returned %d — retrying in %v (X-Request-ID: %s)",
			attempt+1, t.maxRetries, resp.StatusCode, delay, req.Header.Get("X-Request-ID"))

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
		}
	}

	return resp, nil
}

// computeDelay returns the exponential backoff delay for a given attempt,
// capped at 30 seconds, with +/-20% random jitter to prevent thundering herds.
func computeDelay(baseDelay time.Duration, attempt int) time.Duration {
	delay := baseDelay * (1 << uint(attempt))
	const maxDelay = 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	// Apply +/-20% jitter to prevent thundering herds.
	jitterRange := float64(delay) * 0.2
	jitter := (mathrand.Float64()*2 - 1) * jitterRange // [-20%, +20%]
	delay = time.Duration(float64(delay) + jitter)
	if delay < 0 {
		delay = 0
	}
	return delay
}

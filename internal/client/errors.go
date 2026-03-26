package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APISubError is one error entry in a FlashBlade API error response.
type APISubError struct {
	Message string `json:"message"`
}

// APIError represents an error returned by the FlashBlade REST API.
type APIError struct {
	StatusCode int
	Message    string
	Errors     []APISubError
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("FlashBlade API error (HTTP %d): %s", e.StatusCode, e.Errors[0].Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("FlashBlade API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("FlashBlade API error (HTTP %d)", e.StatusCode)
}

// apiErrorBody is used for JSON parsing of error bodies from the FlashBlade API.
type apiErrorBody struct {
	Errors []APISubError `json:"errors"`
}

// ParseAPIError reads the HTTP response and returns an *APIError if the status
// code indicates an error (>= 400). Returns nil for successful responses.
func ParseAPIError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	var eb apiErrorBody
	_ = json.Unmarshal(body, &eb)

	return &APIError{
		StatusCode: resp.StatusCode,
		Errors:     eb.Errors,
	}
}

// IsNotFound returns true when err is an *APIError with HTTP 404.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

// IsRetryable returns true for HTTP status codes that represent transient
// failures: 429 (rate limit) and 5xx server errors.
func IsRetryable(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

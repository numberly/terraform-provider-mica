package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("failed to read response body: %v", err),
		}
	}

	var eb apiErrorBody
	_ = json.Unmarshal(body, &eb)

	return &APIError{
		StatusCode: resp.StatusCode,
		Errors:     eb.Errors,
	}
}

// IsNotFound returns true when err is an *APIError with HTTP 404,
// or HTTP 400 with a sub-error message ending with "does not exist"
// (FlashBlade returns 400 instead of 404 for some resources).
//
// The match is scoped to the first sub-error message (Errors[0].Message)
// using HasSuffix to avoid false positives on validation errors that
// contain "does not exist" mid-sentence (e.g., "parameter does not exist in schema").
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.StatusCode == http.StatusNotFound {
		return true
	}
	// FlashBlade returns 400 instead of 404 for some resources.
	// Match only when the sub-error message ends with "does not exist." or "does not exist".
	if apiErr.StatusCode == http.StatusBadRequest && len(apiErr.Errors) > 0 {
		msg := strings.TrimSpace(apiErr.Errors[0].Message)
		return strings.HasSuffix(msg, "does not exist.") || strings.HasSuffix(msg, "does not exist")
	}
	return false
}

// IsConflict returns true when err is an *APIError with HTTP 409.
func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusConflict
}

// IsUnprocessable returns true when err is an *APIError with HTTP 422.
func IsUnprocessable(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusUnprocessableEntity
}

// IsRetryable returns true for HTTP status codes that represent transient
// failures: 429 (rate limit) and 5xx server errors.
func IsRetryable(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

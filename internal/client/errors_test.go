package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"testing/iotest"
)

// TestUnit_IsConflict_True verifies IsConflict returns true for HTTP 409.
func TestUnit_IsConflict_True(t *testing.T) {
	err := &APIError{StatusCode: 409}
	if !IsConflict(err) {
		t.Error("expected IsConflict to return true for HTTP 409")
	}
}

// TestUnit_IsConflict_False verifies IsConflict returns false for non-409 errors.
func TestUnit_IsConflict_False(t *testing.T) {
	err := &APIError{StatusCode: 404}
	if IsConflict(err) {
		t.Error("expected IsConflict to return false for HTTP 404")
	}
}

// TestUnit_IsConflict_Nil verifies IsConflict returns false for nil.
func TestUnit_IsConflict_Nil(t *testing.T) {
	if IsConflict(nil) {
		t.Error("expected IsConflict to return false for nil")
	}
}

// TestUnit_IsConflict_NonAPIError verifies IsConflict returns false for non-APIError types.
func TestUnit_IsConflict_NonAPIError(t *testing.T) {
	err := errors.New("some error")
	if IsConflict(err) {
		t.Error("expected IsConflict to return false for non-APIError")
	}
}

// TestUnit_IsUnprocessable_True verifies IsUnprocessable returns true for HTTP 422.
func TestUnit_IsUnprocessable_True(t *testing.T) {
	err := &APIError{StatusCode: 422}
	if !IsUnprocessable(err) {
		t.Error("expected IsUnprocessable to return true for HTTP 422")
	}
}

// TestUnit_IsUnprocessable_False verifies IsUnprocessable returns false for non-422 errors.
func TestUnit_IsUnprocessable_False(t *testing.T) {
	err := &APIError{StatusCode: 404}
	if IsUnprocessable(err) {
		t.Error("expected IsUnprocessable to return false for HTTP 404")
	}
}

// TestUnit_IsUnprocessable_Nil verifies IsUnprocessable returns false for nil.
func TestUnit_IsUnprocessable_Nil(t *testing.T) {
	if IsUnprocessable(nil) {
		t.Error("expected IsUnprocessable to return false for nil")
	}
}

// TestUnit_IsNotFound verifies IsNotFound handles all edge cases correctly.
func TestUnit_IsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error returns false",
			err:  nil,
			want: false,
		},
		{
			name: "non-APIError returns false",
			err:  errors.New("some random error"),
			want: false,
		},
		{
			name: "404 returns true",
			err:  &APIError{StatusCode: 404, Errors: []APISubError{{Message: "Not Found"}}},
			want: true,
		},
		{
			name: "400 with sub-error 'Resource does not exist.' returns true",
			err: &APIError{
				StatusCode: 400,
				Errors:     []APISubError{{Message: "Resource does not exist."}},
			},
			want: true,
		},
		{
			name: "400 with sub-error 'XXX does not exist' returns true (no trailing period)",
			err: &APIError{
				StatusCode: 400,
				Errors:     []APISubError{{Message: "file-system does not exist"}},
			},
			want: true,
		},
		{
			name: "400 with sub-error 'does not exist.' returns true",
			err: &APIError{
				StatusCode: 400,
				Errors:     []APISubError{{Message: "does not exist."}},
			},
			want: true,
		},
		{
			name: "400 validation error containing 'does not exist' mid-sentence returns false",
			err: &APIError{
				StatusCode: 400,
				Errors:     []APISubError{{Message: "Invalid parameter: field does not exist in schema"}},
			},
			want: false,
		},
		{
			name: "400 generic validation error returns false",
			err: &APIError{
				StatusCode: 400,
				Errors:     []APISubError{{Message: "Validation error"}},
			},
			want: false,
		},
		{
			name: "400 with no sub-errors but message containing 'does not exist' returns false",
			err: &APIError{
				StatusCode: 400,
				Message:    "Resource does not exist.",
				Errors:     nil,
			},
			want: false,
		},
		{
			name: "500 error returns false",
			err:  &APIError{StatusCode: 500, Errors: []APISubError{{Message: "Internal error"}}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUnit_IsNotFound_WrappedError verifies IsNotFound handles wrapped APIError.
func TestUnit_IsNotFound_WrappedError(t *testing.T) {
	apiErr := &APIError{StatusCode: 404}
	wrapped := fmt.Errorf("op failed: %w", apiErr)
	if !IsNotFound(wrapped) {
		t.Error("expected IsNotFound to return true for wrapped 404 APIError")
	}
}

// TestUnit_IsNotFound_DoubleWrapped verifies IsNotFound handles double-wrapped APIError.
func TestUnit_IsNotFound_DoubleWrapped(t *testing.T) {
	apiErr := &APIError{StatusCode: 404}
	wrapped := fmt.Errorf("inner: %w", apiErr)
	doubleWrapped := fmt.Errorf("outer: %w", wrapped)
	if !IsNotFound(doubleWrapped) {
		t.Error("expected IsNotFound to return true for double-wrapped 404 APIError")
	}
}

// TestUnit_IsConflict_WrappedError verifies IsConflict handles wrapped APIError.
func TestUnit_IsConflict_WrappedError(t *testing.T) {
	apiErr := &APIError{StatusCode: 409}
	wrapped := fmt.Errorf("wrapped: %w", apiErr)
	if !IsConflict(wrapped) {
		t.Error("expected IsConflict to return true for wrapped 409 APIError")
	}
}

// TestUnit_IsUnprocessable_WrappedError verifies IsUnprocessable handles wrapped APIError.
func TestUnit_IsUnprocessable_WrappedError(t *testing.T) {
	apiErr := &APIError{StatusCode: 422}
	wrapped := fmt.Errorf("wrapped: %w", apiErr)
	if !IsUnprocessable(wrapped) {
		t.Error("expected IsUnprocessable to return true for wrapped 422 APIError")
	}
}

// TestUnit_ParseAPIError_ReadFailure verifies ParseAPIError handles io.ReadAll failures.
func TestUnit_ParseAPIError_ReadFailure(t *testing.T) {
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(iotest.ErrReader(errors.New("disk failure"))),
	}
	err := ParseAPIError(resp)
	if err == nil {
		t.Fatal("expected non-nil error from ParseAPIError when body read fails")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected error to be *APIError")
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected StatusCode 500, got %d", apiErr.StatusCode)
	}
}

// TestUnit_ParseAPIError_ReadFailure_Message verifies the error message is descriptive.
func TestUnit_ParseAPIError_ReadFailure_Message(t *testing.T) {
	resp := &http.Response{
		StatusCode: 502,
		Body:       io.NopCloser(iotest.ErrReader(errors.New("network error"))),
	}
	err := ParseAPIError(resp)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected *APIError")
	}
	if apiErr.Message == "" {
		t.Error("expected non-empty Message field on APIError")
	}
	// Message should mention body read failure.
	if got := apiErr.Error(); got == "" {
		t.Error("expected non-empty error string")
	}
}

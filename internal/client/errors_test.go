package client

import (
	"errors"
	"testing"
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

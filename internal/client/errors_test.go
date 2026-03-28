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

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateQueryParams_AcceptsKnownParams(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test?names=foo&ids=bar", nil)
	w := httptest.NewRecorder()

	ok := ValidateQueryParams(w, req, []string{"names", "ids"})
	if !ok {
		t.Fatal("expected ValidateQueryParams to return true for known params")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestValidateQueryParams_AcceptsGlobalParams(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test?names=foo&limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	ok := ValidateQueryParams(w, req, []string{"names"})
	if !ok {
		t.Fatal("expected ValidateQueryParams to return true for global params")
	}
}

func TestValidateQueryParams_RejectsUnknownParam(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test?names=foo&bogus=bar", nil)
	w := httptest.NewRecorder()

	ok := ValidateQueryParams(w, req, []string{"names", "ids"})
	if ok {
		t.Fatal("expected ValidateQueryParams to return false for unknown param")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	// Verify error body.
	var body map[string][]map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	errs := body["errors"]
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0]["message"], "bogus") {
		t.Fatalf("expected error message to mention 'bogus', got %q", errs[0]["message"])
	}
}

func TestValidateQueryParams_AllowsEmptyQuery(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ok := ValidateQueryParams(w, req, []string{"names"})
	if !ok {
		t.Fatal("expected ValidateQueryParams to return true for empty query")
	}
}

func TestRequireQueryParam_ReturnsValueWhenPresent(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test?names=myfs", nil)
	w := httptest.NewRecorder()

	val, ok := RequireQueryParam(w, req, "names")
	if !ok {
		t.Fatal("expected RequireQueryParam to return true when param is present")
	}
	if val != "myfs" {
		t.Fatalf("expected value 'myfs', got %q", val)
	}
}

func TestRequireQueryParam_ReturnsFalseWhenMissing(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	val, ok := RequireQueryParam(w, req, "names")
	if ok {
		t.Fatal("expected RequireQueryParam to return false when param is missing")
	}
	if val != "" {
		t.Fatalf("expected empty value, got %q", val)
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

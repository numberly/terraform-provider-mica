package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_OAuth2TokenSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/1.0/token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"test-access-token","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	ts := client.NewFlashBladeTokenSource(context.Background(), srv.URL, "my-api-token", srv.Client())

	tok, err := ts.Token()
	if err != nil {
		t.Fatalf("Token(): %v", err)
	}
	if tok.AccessToken != "test-access-token" {
		t.Errorf("expected access_token 'test-access-token', got %q", tok.AccessToken)
	}
	if !tok.Valid() {
		t.Error("expected token to be valid")
	}
}

func TestUnit_OAuth2TokenSource_Caching(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"cached-token","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	ts := client.NewFlashBladeTokenSource(context.Background(), srv.URL, "my-api-token", srv.Client())

	// First call — fetches from server.
	if _, err := ts.Token(); err != nil {
		t.Fatalf("first Token(): %v", err)
	}
	// Second call — should use cache.
	if _, err := ts.Token(); err != nil {
		t.Fatalf("second Token(): %v", err)
	}

	if n := atomic.LoadInt32(&callCount); n != 1 {
		t.Errorf("expected 1 HTTP call (caching), got %d", n)
	}
}

func TestUnit_LoginWithAPIToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/login" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("api-token") != "my-api-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("x-auth-token", "session-token-abc")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tok, err := client.LoginWithAPIToken(context.Background(), srv.Client(), srv.URL, "my-api-token")
	if err != nil {
		t.Fatalf("LoginWithAPIToken(): %v", err)
	}
	if tok != "session-token-abc" {
		t.Errorf("expected session token 'session-token-abc', got %q", tok)
	}
}

func TestUnit_APIError_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"file system not found"}]}`))
	}))
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:       srv.URL,
		RetryBaseDelay: 1,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := c.HTTPClient().Get(srv.URL + "/api/2.22/file-systems?names=missing")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	apiErr := client.ParseAPIError(resp)
	if apiErr == nil {
		t.Fatal("expected APIError from 404 response, got nil")
	}
	if !client.IsNotFound(apiErr) {
		t.Errorf("expected IsNotFound=true for 404, got false; err=%v", apiErr)
	}
}

func TestUnit_OAuth2TokenSource_ErrorSanitized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"token expired"}`))
	}))
	defer srv.Close()

	ts := client.NewFlashBladeTokenSource(context.Background(), srv.URL, "bad-token", srv.Client())

	_, err := ts.Token()
	if err == nil {
		t.Fatal("expected error from 401 response, got nil")
	}

	errMsg := err.Error()
	// Must contain the sanitized info.
	if !contains(errMsg, "unexpected HTTP 401") {
		t.Errorf("error should mention 'unexpected HTTP 401', got: %s", errMsg)
	}
	// Must NOT leak response body content.
	if contains(errMsg, "invalid_grant") {
		t.Errorf("error must not contain response body content 'invalid_grant', got: %s", errMsg)
	}
	if contains(errMsg, "token expired") {
		t.Errorf("error must not contain response body content 'token expired', got: %s", errMsg)
	}
}

func TestUnit_OAuth2TokenSource_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	ts := client.NewFlashBladeTokenSource(context.Background(), srv.URL, "my-api-token", srv.Client())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := ts.FetchTokenWithContext(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if !contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got: %v", err)
	}
}

func TestUnit_OAuth2TokenSource_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	ts := client.NewFlashBladeTokenSource(context.Background(), srv.URL, "my-api-token", srv.Client())

	_, err := ts.Token()
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	errMsg := err.Error()
	if !contains(errMsg, "parse response") {
		t.Errorf("expected 'parse response' in error, got: %s", errMsg)
	}
	// Must not leak the body.
	if contains(errMsg, "not-json") {
		t.Errorf("error must not contain raw body 'not-json', got: %s", errMsg)
	}
}

// contains is a helper to avoid importing strings in tests.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestUnit_APIError_Retryable(t *testing.T) {
	retryableCodes := []int{http.StatusTooManyRequests, http.StatusServiceUnavailable, 500, 501, 502, 503}
	for _, code := range retryableCodes {
		if !client.IsRetryable(code) {
			t.Errorf("expected IsRetryable=true for %d", code)
		}
	}

	nonRetryableCodes := []int{http.StatusBadRequest, http.StatusNotFound, http.StatusUnauthorized, http.StatusForbidden}
	for _, code := range nonRetryableCodes {
		if client.IsRetryable(code) {
			t.Errorf("expected IsRetryable=false for %d", code)
		}
	}
}

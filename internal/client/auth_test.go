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

	ts := client.NewFlashBladeTokenSource(srv.URL, "my-api-token", srv.Client())

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

	ts := client.NewFlashBladeTokenSource(srv.URL, "my-api-token", srv.Client())

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

	tok, err := client.LoginWithAPIToken(context.TODO(), srv.Client(), srv.URL, "my-api-token")
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

	c, err := client.NewClient(client.Config{
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

package client_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_RetryTransport_429(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint:       srv.URL,
		RetryBaseDelay: 1, // 1ms for tests
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := c.HTTPClient().Get(srv.URL + "/api/2.22/file-systems")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&callCount) < 2 {
		t.Errorf("expected at least 2 calls (1 retry), got %d", callCount)
	}
}

func TestUnit_RetryTransport_503(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint:       srv.URL,
		RetryBaseDelay: 1,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := c.HTTPClient().Get(srv.URL + "/test")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
}

func TestUnit_RetryTransport_MaxRetries(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	maxRetries := 3
	c, err := client.NewClient(client.Config{
		Endpoint:       srv.URL,
		MaxRetries:     maxRetries,
		RetryBaseDelay: 1,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := c.HTTPClient().Get(srv.URL + "/test")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 after max retries exhausted, got %d", resp.StatusCode)
	}
	// 1 original + maxRetries
	expected := int32(maxRetries + 1)
	if atomic.LoadInt32(&callCount) != expected {
		t.Errorf("expected %d calls, got %d", expected, callCount)
	}
}

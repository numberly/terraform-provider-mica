// Package testmock provides a reusable mock HTTP server that simulates the
// FlashBlade REST API for use in provider-level and unit tests.
package testmock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// MockServer wraps an httptest.Server with a configurable ServeMux and
// convenience methods for test setup.
type MockServer struct {
	// Server is the underlying test HTTP server.
	Server *httptest.Server
	// Mux is the request multiplexer — register resource handlers against it.
	Mux *http.ServeMux
}

// NewMockServer creates and starts a new mock HTTP server with built-in
// handlers for /login and /api/api_version.
func NewMockServer() *MockServer {
	mux := http.NewServeMux()
	ms := &MockServer{
		Mux: mux,
	}

	// Register built-in handlers.
	mux.HandleFunc("/api/login", ms.handleLogin)
	mux.HandleFunc("/api/api_version", ms.handleAPIVersion)

	ms.Server = httptest.NewServer(mux)
	return ms
}

// URL returns the base URL of the mock server.
func (ms *MockServer) URL() string {
	return ms.Server.URL
}

// Close shuts down the mock server.
func (ms *MockServer) Close() {
	ms.Server.Close()
}

// RegisterHandler adds a handler for the given URL pattern to the server's mux.
func (ms *MockServer) RegisterHandler(pattern string, handler http.HandlerFunc) {
	ms.Mux.HandleFunc(pattern, handler)
}

// handleLogin handles POST /login by returning a mock session token.
func (ms *MockServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("x-auth-token", "mock-session-token")
	w.WriteHeader(http.StatusOK)
}

// handleAPIVersion handles GET /api/api_version by returning a versions list
// that includes the target API version "2.22".
func (ms *MockServer) handleAPIVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": []string{"2.12", "2.15", "2.22"},
	})
}

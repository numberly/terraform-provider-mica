package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// arrayConnectionKeyStore holds at most one connection key (singleton per array).
type arrayConnectionKeyStore struct {
	mu      sync.Mutex
	current *client.ArrayConnectionKey
	nextID  int
}

// RegisterArrayConnectionKeyHandlers registers GET/POST handlers for
// /api/2.22/array-connections/connection-key against the provided ServeMux.
// The store pointer is returned for test setup (Seed).
func RegisterArrayConnectionKeyHandlers(mux *http.ServeMux) *arrayConnectionKeyStore {
	store := &arrayConnectionKeyStore{nextID: 1}
	mux.HandleFunc("/api/2.22/array-connections/connection-key", store.handle)
	return store
}

// Seed sets the current key directly (used by tests to pre-populate state).
func (s *arrayConnectionKeyStore) Seed(key *client.ArrayConnectionKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = key
}

func (s *arrayConnectionKeyStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET /api/2.22/array-connections/connection-key.
// Returns the current key as a plain JSON object (not a list envelope).
// If no key has been set, returns a zero-value object with HTTP 200.
func (s *arrayConnectionKeyStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var items []client.ArrayConnectionKey
	if s.current != nil {
		items = append(items, *s.current)
	}
	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/array-connections/connection-key.
// Generates a new synthetic key, overwrites the current one, and returns it as a plain JSON object.
func (s *arrayConnectionKeyStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := &client.ArrayConnectionKey{
		ConnectionKey: fmt.Sprintf("conn-key-%d", s.nextID),
		Created:       1000000000000,
		Expires:       1000003600000,
	}
	s.nextID++
	s.current = key

	WriteJSONListResponse(w, http.StatusOK, []client.ArrayConnectionKey{*key})
}

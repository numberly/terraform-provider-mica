package handlers

import (
	"net/http"
	"sync"
)

// objectStoreUserStore is the thread-safe in-memory state for object store user handlers.
type objectStoreUserStore struct {
	mu       sync.Mutex
	byName   map[string]bool
	accounts *objectStoreAccountStore
}

// RegisterObjectStoreUserHandlers registers GET/POST handlers for /api/2.22/object-store-users.
// Returns the store for cross-reference if needed.
func RegisterObjectStoreUserHandlers(mux *http.ServeMux, accounts *objectStoreAccountStore) *objectStoreUserStore {
	store := &objectStoreUserStore{
		byName:   make(map[string]bool),
		accounts: accounts,
	}
	mux.HandleFunc("/api/2.22/object-store-users", store.handle)
	return store
}

func (s *objectStoreUserStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *objectStoreUserStore) handleGet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.byName[name] {
		// Return empty items list (provider synthesizes 404 from this).
		WriteJSONListResponse(w, http.StatusOK, []map[string]any{})
		return
	}

	WriteJSONListResponse(w, http.StatusOK, []map[string]any{
		{"name": name},
	})
}

func (s *objectStoreUserStore) handlePost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.byName[name] = true

	WriteJSONListResponse(w, http.StatusOK, []map[string]any{
		{"name": name},
	})
}

func (s *objectStoreUserStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.byName, name)
	w.WriteHeader(http.StatusOK)
}

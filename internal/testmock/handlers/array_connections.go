package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// arrayConnectionStore is the thread-safe in-memory state for array connection handlers.
type arrayConnectionStore struct {
	mu     sync.Mutex
	byName map[string]*client.ArrayConnection
	nextID int
}

// RegisterArrayConnectionHandlers registers CRUD handlers for /api/2.22/array-connections
// against the provided ServeMux. The store pointer is returned for test setup.
func RegisterArrayConnectionHandlers(mux *http.ServeMux) *arrayConnectionStore {
	store := &arrayConnectionStore{
		byName: make(map[string]*client.ArrayConnection),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.22/array-connections", store.handle)
	return store
}

// Seed inserts an array connection directly into the store (used by tests to pre-populate data).
// The connection is keyed by conn.Remote.Name.
func (s *arrayConnectionStore) Seed(conn *client.ArrayConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[conn.Remote.Name] = conn
}

// handle dispatches array connection requests by HTTP method.
func (s *arrayConnectionStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodPatch:
		s.handlePatch(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET /api/2.22/array-connections with optional ?remote_names= filter.
// When the filter finds no match, returns an empty list with HTTP 200 (not 404).
func (s *arrayConnectionStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"remote_names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	filter := r.URL.Query().Get("remote_names")

	var items []client.ArrayConnection
	if filter != "" {
		if conn, ok := s.byName[filter]; ok {
			items = append(items, *conn)
		}
	} else {
		for _, conn := range s.byName {
			items = append(items, *conn)
		}
	}

	if items == nil {
		items = []client.ArrayConnection{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/array-connections?remote_names={remoteName}.
// Returns 409 if a connection for that remote already exists.
func (s *arrayConnectionStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"remote_names"}) {
		return
	}

	remoteName, ok := RequireQueryParam(w, r, "remote_names")
	if !ok {
		return
	}

	var body client.ArrayConnectionPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[remoteName]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("array connection for remote %q already exists", remoteName))
		return
	}

	id := fmt.Sprintf("array-connection-%d", s.nextID)
	s.nextID++

	conn := &client.ArrayConnection{
		ID:                   id,
		Remote:               client.NamedReference{Name: remoteName},
		ManagementAddress:    body.ManagementAddress,
		Encrypted:            body.Encrypted,
		ReplicationAddresses: body.ReplicationAddresses,
		Throttle:             body.Throttle,
		Status:               "connected",
		Type:                 "async-replication",
	}
	if body.CACertificateGroup != nil {
		conn.CACertificateGroup = body.CACertificateGroup
	}

	s.byName[remoteName] = conn

	WriteJSONListResponse(w, http.StatusOK, []client.ArrayConnection{*conn})
}

// handlePatch handles PATCH /api/2.22/array-connections?remote_names={remoteName}.
// Applies non-nil pointer fields. Returns 404 if the connection is not found.
func (s *arrayConnectionStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"remote_names"}) {
		return
	}

	remoteName, ok := RequireQueryParam(w, r, "remote_names")
	if !ok {
		return
	}

	var body client.ArrayConnectionPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.byName[remoteName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("array connection for remote %q not found", remoteName))
		return
	}

	if body.ManagementAddress != nil {
		conn.ManagementAddress = *body.ManagementAddress
	}
	if body.Encrypted != nil {
		conn.Encrypted = *body.Encrypted
	}
	// **NamedReference: outer ptr non-nil means field was sent.
	// Inner ptr nil means set to null (clear); inner ptr non-nil means set to value.
	if body.CACertificateGroup != nil {
		conn.CACertificateGroup = *body.CACertificateGroup
	}
	if body.ReplicationAddresses != nil {
		conn.ReplicationAddresses = *body.ReplicationAddresses
	}
	if body.Throttle != nil {
		conn.Throttle = body.Throttle
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ArrayConnection{*conn})
}

// handleDelete handles DELETE /api/2.22/array-connections?remote_names={remoteName}.
func (s *arrayConnectionStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"remote_names"}) {
		return
	}

	remoteName, ok := RequireQueryParam(w, r, "remote_names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[remoteName]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("array connection for remote %q not found", remoteName))
		return
	}

	delete(s.byName, remoteName)

	w.WriteHeader(http.StatusOK)
}

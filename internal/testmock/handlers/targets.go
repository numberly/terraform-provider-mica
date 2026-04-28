package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// targetStore is the thread-safe in-memory state for target handlers.
type targetStore struct {
	mu     sync.Mutex
	byName map[string]*client.Target
	nextID int
}

// RegisterTargetHandlers registers CRUD handlers for /api/2.22/targets
// against the provided ServeMux. The store pointer is returned for test setup.
func RegisterTargetHandlers(mux *http.ServeMux) *targetStore {
	store := &targetStore{
		byName: make(map[string]*client.Target),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.22/targets", store.handle)
	return store
}

// Seed adds a target directly to the store for test setup.
func (s *targetStore) Seed(t *client.Target) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[t.Name] = t
}

func (s *targetStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/targets with optional ?names= param.
func (s *targetStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.Target

	if namesFilter != "" {
		tgt, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *tgt)
		}
	} else {
		for _, tgt := range s.byName {
			items = append(items, *tgt)
		}
	}

	if items == nil {
		items = []client.Target{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/targets?names={name}.
// Requires non-empty address in body. Returns 409 if name already exists.
func (s *targetStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.TargetPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if body.Address == "" {
		WriteJSONError(w, http.StatusBadRequest, "address is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("target %q already exists", name))
		return
	}

	id := fmt.Sprintf("tgt-%d", s.nextID)
	s.nextID++

	tgt := &client.Target{
		ID:            id,
		Name:          name,
		Address:       body.Address,
		Status:        "connected",
		StatusDetails: "",
	}

	s.byName[name] = tgt

	WriteJSONListResponse(w, http.StatusOK, []client.Target{*tgt})
}

// handlePatch handles PATCH /api/2.22/targets?names={name}.
// Applies non-nil pointer fields. Returns 404 if not found.
func (s *targetStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.TargetPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tgt, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("target %q not found", name))
		return
	}

	if body.Address != nil {
		tgt.Address = *body.Address
	}
	// **NamedReference: outer ptr non-nil means field was sent.
	// Inner ptr nil means set to null (clear); inner ptr non-nil means set to value.
	if body.CACertificateGroup != nil {
		tgt.CACertificateGroup = *body.CACertificateGroup
	}

	WriteJSONListResponse(w, http.StatusOK, []client.Target{*tgt})
}

// handleDelete handles DELETE /api/2.22/targets?names={name}.
func (s *targetStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("target %q not found", name))
		return
	}

	delete(s.byName, name)

	w.WriteHeader(http.StatusOK)
}

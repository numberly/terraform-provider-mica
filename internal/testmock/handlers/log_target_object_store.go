package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// logTargetObjectStoreStore is the thread-safe in-memory state for log target object store handlers.
type logTargetObjectStoreStore struct {
	mu     sync.Mutex
	byName map[string]*client.LogTargetObjectStore
	nextID int
}

// RegisterLogTargetObjectStoreHandlers registers CRUD handlers for /api/2.22/log-targets/object-store
// against the provided ServeMux. The store pointer is returned for seeding in tests.
func RegisterLogTargetObjectStoreHandlers(mux *http.ServeMux) *logTargetObjectStoreStore {
	store := &logTargetObjectStoreStore{
		byName: make(map[string]*client.LogTargetObjectStore),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.22/log-targets/object-store", store.handle)
	return store
}

// Seed adds a log target object store directly to the store for test setup.
func (s *logTargetObjectStoreStore) Seed(item *client.LogTargetObjectStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[item.Name] = item
}

func (s *logTargetObjectStoreStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/log-targets/object-store with optional ?names= param.
// Returns empty list with HTTP 200 when not found (matches real FlashBlade API behavior).
func (s *logTargetObjectStoreStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.LogTargetObjectStore

	if namesFilter != "" {
		item, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *item)
		}
	} else {
		for _, item := range s.byName {
			items = append(items, *item)
		}
	}

	if items == nil {
		items = []client.LogTargetObjectStore{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/log-targets/object-store.
// Requires ?names= query param. Returns 409 on name conflict.
func (s *logTargetObjectStoreStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.LogTargetObjectStorePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("log target object store %q already exists", name))
		return
	}

	id := fmt.Sprintf("ltos-%d", s.nextID)
	s.nextID++

	item := &client.LogTargetObjectStore{
		ID:            id,
		Name:          name,
		Bucket:        body.Bucket,
		LogNamePrefix: body.LogNamePrefix,
		LogRotate:     body.LogRotate,
	}

	s.byName[name] = item

	WriteJSONListResponse(w, http.StatusOK, []client.LogTargetObjectStore{*item})
}

// handlePatch handles PATCH /api/2.22/log-targets/object-store?names={name}.
// Applies non-nil fields from the body. Returns 404 if the item does not exist.
func (s *logTargetObjectStoreStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.LogTargetObjectStorePatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("log target object store %q not found", name))
		return
	}

	if body.Bucket != nil {
		item.Bucket = *body.Bucket
	}
	if body.LogNamePrefix != nil {
		item.LogNamePrefix = *body.LogNamePrefix
	}
	if body.LogRotate != nil {
		item.LogRotate = *body.LogRotate
	}

	WriteJSONListResponse(w, http.StatusOK, []client.LogTargetObjectStore{*item})
}

// handleDelete handles DELETE /api/2.22/log-targets/object-store?names={name}.
// Returns 404 if the item does not exist.
func (s *logTargetObjectStoreStore) handleDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("log target object store %q not found", name))
		return
	}

	delete(s.byName, name)

	w.WriteHeader(http.StatusOK)
}

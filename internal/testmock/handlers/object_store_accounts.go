package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// objectStoreAccountStore is the thread-safe in-memory state for object store account handlers.
type objectStoreAccountStore struct {
	mu     sync.Mutex
	byName map[string]*client.ObjectStoreAccount
	byID   map[string]*client.ObjectStoreAccount
}

// RegisterObjectStoreAccountHandlers registers CRUD handlers for /api/2.22/object-store-accounts
// against the provided ServeMux. The handlers share in-memory state and are thread-safe.
// The store pointer is returned so bucket handlers can cross-reference accounts.
func RegisterObjectStoreAccountHandlers(mux *http.ServeMux) *objectStoreAccountStore {
	store := &objectStoreAccountStore{
		byName: make(map[string]*client.ObjectStoreAccount),
		byID:   make(map[string]*client.ObjectStoreAccount),
	}
	mux.HandleFunc("/api/2.22/object-store-accounts", store.handle)
	return store
}

// handle dispatches object store account requests by HTTP method.
func (s *objectStoreAccountStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/object-store-accounts with optional ?names= param.
func (s *objectStoreAccountStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.ObjectStoreAccount

	if namesFilter != "" {
		acct, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *acct)
		}
	} else {
		for _, acct := range s.byID {
			items = append(items, *acct)
		}
	}

	if items == nil {
		items = []client.ObjectStoreAccount{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/object-store-accounts?names={name}.
// The account name comes from the ?names= query parameter.
func (s *objectStoreAccountStore) handlePost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.ObjectStoreAccountPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("object store account %q already exists", name))
		return
	}

	acct := &client.ObjectStoreAccount{
		ID:               uuid.New().String(),
		Name:             name,
		Created:          time.Now().UnixMilli(),
		QuotaLimit:       body.QuotaLimit,
		HardLimitEnabled: body.HardLimitEnabled,
		Space:            client.Space{},
	}

	s.byName[acct.Name] = acct
	s.byID[acct.ID] = acct

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccount{*acct})
}

// handlePatch handles PATCH /api/2.22/object-store-accounts?names={name}.
func (s *objectStoreAccountStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	acct, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store account %q not found", name))
		return
	}

	// Use a raw map to decode only provided fields (true PATCH semantics).
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["quota_limit"]; ok {
		var quotaLimit string
		if err := json.Unmarshal(v, &quotaLimit); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid quota_limit field")
			return
		}
		acct.QuotaLimit = quotaLimit
	}

	if v, ok := rawPatch["hard_limit_enabled"]; ok {
		var hardLimitEnabled bool
		if err := json.Unmarshal(v, &hardLimitEnabled); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid hard_limit_enabled field")
			return
		}
		acct.HardLimitEnabled = hardLimitEnabled
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccount{*acct})
}

// handleDelete handles DELETE /api/2.22/object-store-accounts?names={name}.
// Single-phase delete (no soft-delete for object store accounts).
func (s *objectStoreAccountStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	acct, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store account %q not found", name))
		return
	}

	delete(s.byName, acct.Name)
	delete(s.byID, acct.ID)

	w.WriteHeader(http.StatusOK)
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// directoryServiceRolesStore is the thread-safe in-memory state for directory service role handlers.
type directoryServiceRolesStore struct {
	mu     sync.Mutex
	byName map[string]*client.DirectoryServiceRole
	nextID int
}

// RegisterDirectoryServiceRolesHandlers registers CRUD handlers for /api/2.22/directory-services/roles
// against the provided ServeMux. The store pointer is returned for test setup.
func RegisterDirectoryServiceRolesHandlers(mux *http.ServeMux) *directoryServiceRolesStore {
	s := &directoryServiceRolesStore{
		byName: make(map[string]*client.DirectoryServiceRole),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.22/directory-services/roles", s.handle)
	return s
}

// Seed adds a directory service role directly to the store for test setup.
// ID is auto-generated if empty.
func (s *directoryServiceRolesStore) Seed(role *client.DirectoryServiceRole) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if role.ID == "" {
		role.ID = fmt.Sprintf("dsr-%d", s.nextID)
		s.nextID++
	}
	s.byName[role.Name] = role
}

func (s *directoryServiceRolesStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/directory-services/roles with optional ?names= filter.
// Returns HTTP 200 with empty list when name not found (matches real API behaviour;
// lets getOneByName[T] detect not-found via list length).
func (s *directoryServiceRolesStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var items []client.DirectoryServiceRole
	if n := r.URL.Query().Get("names"); n != "" {
		if role, ok := s.byName[n]; ok {
			items = append(items, *role)
		}
	} else {
		for _, role := range s.byName {
			items = append(items, *role)
		}
	}
	if items == nil {
		items = []client.DirectoryServiceRole{}
	}
	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/directory-services/roles?names={name}.
// Requires ?names= query param. Returns 409 when name already exists.
func (s *directoryServiceRolesStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.DirectoryServiceRolePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("role %q already exists", name))
		return
	}

	role := &client.DirectoryServiceRole{
		ID:                       fmt.Sprintf("dsr-%d", s.nextID),
		Name:                     name,
		Group:                    body.Group,
		GroupBase:                body.GroupBase,
		ManagementAccessPolicies: body.ManagementAccessPolicies,
		Role:                     body.Role,
	}
	s.nextID++
	s.byName[name] = role
	WriteJSONListResponse(w, http.StatusOK, []client.DirectoryServiceRole{*role})
}

// handlePatch handles PATCH /api/2.22/directory-services/roles?names={name}.
// Rejects management_access_policies in body with 400 (readonly per swagger — mutations go
// through /management-access-policies/directory-services/roles endpoint instead).
// Applies group and group_base when non-nil.
func (s *directoryServiceRolesStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}
	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	// Decode raw first to enforce readonly-on-PATCH contract for management_access_policies.
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}
	if _, present := raw["management_access_policies"]; present {
		WriteJSONError(w, http.StatusBadRequest,
			"management_access_policies is readonly on PATCH; use /management-access-policies/directory-services/roles membership endpoint")
		return
	}

	// Re-decode mutable fields from the raw map.
	var body client.DirectoryServiceRolePatch
	if raw["group"] != nil {
		var v string
		if err := json.Unmarshal(raw["group"], &v); err == nil {
			body.Group = &v
		}
	}
	if raw["group_base"] != nil {
		var v string
		if err := json.Unmarshal(raw["group_base"], &v); err == nil {
			body.GroupBase = &v
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("role %q not found", name))
		return
	}
	if body.Group != nil {
		role.Group = *body.Group
	}
	if body.GroupBase != nil {
		role.GroupBase = *body.GroupBase
	}
	WriteJSONListResponse(w, http.StatusOK, []client.DirectoryServiceRole{*role})
}

// handleDelete handles DELETE /api/2.22/directory-services/roles?names={name}.
// Returns 404 if the role does not exist.
func (s *directoryServiceRolesStore) handleDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("role %q not found", name))
		return
	}
	delete(s.byName, name)
	w.WriteHeader(http.StatusOK)
}

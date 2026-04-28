package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// directoryServicesStore is the thread-safe in-memory state for directory service handlers.
type directoryServicesStore struct {
	mu     sync.Mutex
	byName map[string]*client.DirectoryService
	nextID int
}

// RegisterDirectoryServicesHandlers registers handlers for /api/2.22/directory-services
// against the provided ServeMux. The store pointer is returned for test setup.
// Endpoint supports GET + PATCH only (confirmed against api_references/2.22.md line 431 —
// no POST, no DELETE on the directory-services collection).
func RegisterDirectoryServicesHandlers(mux *http.ServeMux) *directoryServicesStore {
	store := &directoryServicesStore{
		byName: make(map[string]*client.DirectoryService),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.22/directory-services", store.handle)
	return store
}

// Seed adds a directory service directly to the store for test setup.
func (s *directoryServicesStore) Seed(ds *client.DirectoryService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ds.ID == "" {
		ds.ID = fmt.Sprintf("ds-%d", s.nextID)
		s.nextID++
	}
	s.byName[ds.Name] = ds
}

// handle dispatches directory service requests. Endpoint supports GET + PATCH only.
func (s *directoryServicesStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPatch:
		s.handlePatch(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET /api/2.22/directory-services with optional ?names= filter.
// Returns HTTP 200 with empty list when name not found (matches real API behaviour;
// lets getOneByName[T] detect not-found via list length).
func (s *directoryServicesStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")
	var items []client.DirectoryService
	if namesFilter != "" {
		ds, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *ds)
		}
	} else {
		for _, ds := range s.byName {
			items = append(items, *ds)
		}
	}
	if items == nil {
		items = []client.DirectoryService{}
	}
	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePatch handles PATCH /api/2.22/directory-services?names={name}.
// Applies non-nil pointer fields; returns 404 when name missing from store.
// Supports **NamedReference clear-or-set semantics for ca_certificate and ca_certificate_group:
//   - outer nil ptr = field omitted (not sent in PATCH body)
//   - outer non-nil + inner nil = set to null (clear reference)
//   - outer non-nil + inner non-nil = set to named reference value
func (s *directoryServicesStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}
	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.DirectoryServicePatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ds, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("directory service %q not found", name))
		return
	}

	if body.Enabled != nil {
		ds.Enabled = *body.Enabled
	}
	if body.URIs != nil {
		ds.URIs = *body.URIs
	}
	if body.BaseDN != nil {
		ds.BaseDN = *body.BaseDN
	}
	if body.BindUser != nil {
		ds.BindUser = *body.BindUser
	}
	// bind_password is write-only — never stored back, never returned.
	// **NamedReference: outer ptr non-nil means field was sent.
	// Inner ptr nil means set to null (clear); inner ptr non-nil means set to value.
	if body.CACertificate != nil {
		ds.CACertificate = *body.CACertificate
	}
	if body.CACertificateGroup != nil {
		ds.CACertificateGroup = *body.CACertificateGroup
	}
	if body.Management != nil {
		if body.Management.UserLoginAttribute != nil {
			ds.Management.UserLoginAttribute = *body.Management.UserLoginAttribute
		}
		if body.Management.UserObjectClass != nil {
			ds.Management.UserObjectClass = *body.Management.UserObjectClass
		}
		if body.Management.SSHPublicKeyAttribute != nil {
			ds.Management.SSHPublicKeyAttribute = *body.Management.SSHPublicKeyAttribute
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.DirectoryService{*ds})
}

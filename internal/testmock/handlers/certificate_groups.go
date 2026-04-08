package handlers

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// certificateGroupStore is the thread-safe in-memory state for certificate group handlers.
type certificateGroupStore struct {
	mu      sync.Mutex
	groups  map[string]*client.CertificateGroup        // keyed by group name
	members map[string][]client.CertificateGroupMember // keyed by group name
	nextID  int
}

// RegisterCertificateGroupHandlers registers CRUD handlers for:
//   - /api/2.22/certificate-groups/certificates (member GET/POST/DELETE)
//   - /api/2.22/certificate-groups (group GET/POST/DELETE — no PATCH in API)
//
// The certificates endpoint is registered before the groups endpoint to avoid
// ServeMux prefix collision (longer path wins in Go's ServeMux).
// Returns the store so tests can call Seed and SeedMember.
func RegisterCertificateGroupHandlers(mux *http.ServeMux) *certificateGroupStore {
	store := &certificateGroupStore{
		groups:  make(map[string]*client.CertificateGroup),
		members: make(map[string][]client.CertificateGroupMember),
	}
	mux.HandleFunc("/api/2.22/certificate-groups/certificates", store.handleCertificates)
	mux.HandleFunc("/api/2.22/certificate-groups", store.handleGroup)
	return store
}

// Seed adds a certificate group directly to the store for test setup.
func (s *certificateGroupStore) Seed(group *client.CertificateGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups[group.Name] = group
}

// SeedMember adds a certificate group member directly to the store for test setup.
func (s *certificateGroupStore) SeedMember(groupName string, member client.CertificateGroupMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[groupName] = append(s.members[groupName], member)
}

// handleGroup dispatches certificate group requests by HTTP method.
func (s *certificateGroupStore) handleGroup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGroupGet(w, r)
	case http.MethodPost:
		s.handleGroupPost(w, r)
	case http.MethodDelete:
		s.handleGroupDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGroupGet handles GET /api/2.22/certificate-groups.
// When ?names= filter matches nothing, returns empty list with HTTP 200 (not 404).
func (s *certificateGroupStore) handleGroupGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	namesFilter := q.Get("names")

	var items []client.CertificateGroup

	if namesFilter != "" {
		if group, ok := s.groups[namesFilter]; ok {
			items = append(items, *group)
		}
	} else {
		for _, group := range s.groups {
			items = append(items, *group)
		}
	}

	if items == nil {
		items = []client.CertificateGroup{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleGroupPost handles POST /api/2.22/certificate-groups.
func (s *certificateGroupStore) handleGroupPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	// Drain body — no fields to parse for this resource.
	_, _ = io.ReadAll(r.Body)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("certificate group %q already exists", name))
		return
	}

	s.nextID++
	id := fmt.Sprintf("certgroup-%d", s.nextID)

	group := &client.CertificateGroup{
		ID:     id,
		Name:   name,
		Realms: []string{},
	}

	s.groups[name] = group

	WriteJSONListResponse(w, http.StatusOK, []client.CertificateGroup{*group})
}

// handleGroupDelete handles DELETE /api/2.22/certificate-groups.
func (s *certificateGroupStore) handleGroupDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[name]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("certificate group %q not found", name))
		return
	}

	delete(s.groups, name)
	delete(s.members, name)

	w.WriteHeader(http.StatusOK)
}

// handleCertificates dispatches certificate member requests by HTTP method.
func (s *certificateGroupStore) handleCertificates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleMemberGet(w, r)
	case http.MethodPost:
		s.handleMemberPost(w, r)
	case http.MethodDelete:
		s.handleMemberDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMemberGet handles GET /api/2.22/certificate-groups/certificates.
// When certificate_group_names filter matches nothing, returns empty list with HTTP 200.
func (s *certificateGroupStore) handleMemberGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"certificate_group_names", "certificate_names", "continuation_token"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	groupNamesFilter := q.Get("certificate_group_names")

	var items []client.CertificateGroupMember

	if groupNamesFilter != "" {
		if members, ok := s.members[groupNamesFilter]; ok {
			items = append(items, members...)
		}
	} else {
		for _, members := range s.members {
			items = append(items, members...)
		}
	}

	if items == nil {
		items = []client.CertificateGroupMember{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleMemberPost handles POST /api/2.22/certificate-groups/certificates.
func (s *certificateGroupStore) handleMemberPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"certificate_group_names", "certificate_names"}) {
		return
	}

	groupName, ok := RequireQueryParam(w, r, "certificate_group_names")
	if !ok {
		return
	}

	certName := r.URL.Query().Get("certificate_names")
	if certName == "" {
		WriteJSONError(w, http.StatusBadRequest, "certificate_names query parameter is required")
		return
	}

	// Drain body — no fields needed.
	_, _ = io.ReadAll(r.Body)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.groups[groupName]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("certificate group %q not found", groupName))
		return
	}

	member := client.CertificateGroupMember{
		Certificate: client.NamedReference{Name: certName},
		Group:       client.NamedReference{Name: groupName},
	}

	s.members[groupName] = append(s.members[groupName], member)

	WriteJSONListResponse(w, http.StatusOK, []client.CertificateGroupMember{member})
}

// handleMemberDelete handles DELETE /api/2.22/certificate-groups/certificates.
func (s *certificateGroupStore) handleMemberDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"certificate_group_names", "certificate_names"}) {
		return
	}

	groupName, ok := RequireQueryParam(w, r, "certificate_group_names")
	if !ok {
		return
	}

	certName := r.URL.Query().Get("certificate_names")
	if certName == "" {
		WriteJSONError(w, http.StatusBadRequest, "certificate_names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	members, exists := s.members[groupName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("certificate group %q not found or has no members", groupName))
		return
	}

	found := false
	for i, m := range members {
		if m.Certificate.Name == certName {
			s.members[groupName] = append(members[:i], members[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("certificate %q not found in group %q", certName, groupName))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CertificateGroupStoreFacade exposes Seed and SeedMember on the unexported store for use in tests
// from packages outside the handlers package.
type CertificateGroupStoreFacade struct {
	store *certificateGroupStore
}

// NewCertificateGroupStoreFacade wraps the internal store so cross-package tests can call Seed and SeedMember.
func NewCertificateGroupStoreFacade(store *certificateGroupStore) *CertificateGroupStoreFacade {
	return &CertificateGroupStoreFacade{store: store}
}

// Seed adds a certificate group to the store.
func (f *CertificateGroupStoreFacade) Seed(group *client.CertificateGroup) {
	f.store.Seed(group)
}

// SeedMember adds a certificate group member to the store.
func (f *CertificateGroupStoreFacade) SeedMember(groupName string, member client.CertificateGroupMember) {
	f.store.SeedMember(groupName, member)
}

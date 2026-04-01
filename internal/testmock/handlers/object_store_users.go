package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// objectStoreUserStore is the thread-safe in-memory state for object store user handlers.
type objectStoreUserStore struct {
	mu       sync.Mutex
	byName   map[string]*client.ObjectStoreUser
	policies map[string][]string // userName -> []policyName
	accounts *objectStoreAccountStore
}

// RegisterObjectStoreUserHandlers registers GET/POST/DELETE handlers for
// /api/2.22/object-store-users and its sub-path
// /api/2.22/object-store-users/object-store-access-policies.
// Returns the store for cross-reference or test setup.
func RegisterObjectStoreUserHandlers(mux *http.ServeMux, accounts *objectStoreAccountStore) *objectStoreUserStore {
	store := &objectStoreUserStore{
		byName:   make(map[string]*client.ObjectStoreUser),
		policies: make(map[string][]string),
		accounts: accounts,
	}
	mux.HandleFunc("/api/2.22/object-store-users", store.handle)
	mux.HandleFunc("/api/2.22/object-store-users/object-store-access-policies", store.handlePolicies)
	return store
}

// AddPolicyForTest pre-populates the policy store for use in unit tests.
func (s *objectStoreUserStore) AddPolicyForTest(userName, policyName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[userName] = append(s.policies[userName], policyName)
}

// AddUserForTest pre-populates the user store for use in unit tests.
func (s *objectStoreUserStore) AddUserForTest(name string, fullAccess bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[name] = &client.ObjectStoreUser{
		ID:         uuid.NewString(),
		Name:       name,
		FullAccess: fullAccess,
	}
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

	u := s.byName[name]
	if u == nil {
		// Return empty items list (provider synthesizes 404 from this).
		WriteJSONListResponse(w, http.StatusOK, []map[string]any{})
		return
	}

	WriteJSONListResponse(w, http.StatusOK, []map[string]any{
		{"name": u.Name, "id": u.ID, "full_access": u.FullAccess},
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

	if s.byName[name] != nil {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("object store user %q already exists", name))
		return
	}

	u := &client.ObjectStoreUser{
		ID:         uuid.NewString(),
		Name:       name,
		FullAccess: false,
	}
	s.byName[name] = u

	WriteJSONListResponse(w, http.StatusOK, []map[string]any{
		{"name": u.Name, "id": u.ID, "full_access": u.FullAccess},
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
	delete(s.policies, name)
	w.WriteHeader(http.StatusOK)
}

// handlePolicies routes GET/POST/DELETE for /object-store-users/object-store-access-policies.
func (s *objectStoreUserStore) handlePolicies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handlePoliciesGet(w, r)
	case http.MethodPost:
		s.handlePoliciesPost(w, r)
	case http.MethodDelete:
		s.handlePoliciesDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *objectStoreUserStore) handlePoliciesGet(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("member_names")
	if userName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pols := s.policies[userName]
	items := make([]client.ObjectStoreUserPolicyMember, 0, len(pols))
	for _, pol := range pols {
		items = append(items, client.ObjectStoreUserPolicyMember{
			Member: client.NamedReference{Name: userName},
			Policy: client.NamedReference{Name: pol},
		})
	}
	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *objectStoreUserStore) handlePoliciesPost(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("member_names")
	policyName := r.URL.Query().Get("policy_names")
	if userName == "" || policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names and policy_names query parameters are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.policies[userName] {
		if existing == policyName {
			WriteJSONError(w, http.StatusConflict, fmt.Sprintf("policy %q already attached to user %q", policyName, userName))
			return
		}
	}

	s.policies[userName] = append(s.policies[userName], policyName)

	member := client.ObjectStoreUserPolicyMember{
		Member: client.NamedReference{Name: userName},
		Policy: client.NamedReference{Name: policyName},
	}
	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreUserPolicyMember{member})
}

func (s *objectStoreUserStore) handlePoliciesDelete(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("member_names")
	policyName := r.URL.Query().Get("policy_names")
	if userName == "" || policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names and policy_names query parameters are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.policies[userName]
	filtered := existing[:0]
	for _, p := range existing {
		if p != policyName {
			filtered = append(filtered, p)
		}
	}
	s.policies[userName] = filtered

	w.WriteHeader(http.StatusOK)
}

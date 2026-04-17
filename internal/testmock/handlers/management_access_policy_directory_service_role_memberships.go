package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// mapDsrMembershipsStore is the thread-safe mock state for the
// management-access-policy / directory-service-role association set.
type mapDsrMembershipsStore struct {
	mu  sync.Mutex
	set map[string]struct{} // key: "<policy>|<role>"
}

// RegisterManagementAccessPolicyDirectoryServiceRoleMembershipsHandlers registers
// GET/POST/DELETE handlers for /api/2.22/management-access-policies/directory-services/roles.
// Returns the store so tests can Seed pre-existing associations.
func RegisterManagementAccessPolicyDirectoryServiceRoleMembershipsHandlers(mux *http.ServeMux) *mapDsrMembershipsStore {
	s := &mapDsrMembershipsStore{set: make(map[string]struct{})}
	mux.HandleFunc("/api/2.22/management-access-policies/directory-services/roles", s.handle)
	return s
}

// Seed pre-creates an association pair for test setup.
func (s *mapDsrMembershipsStore) Seed(policyName, roleName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set[dsrMemberKey(policyName, roleName)] = struct{}{}
}

// dsrMemberKey produces the composite key for a policy+role pair.
func dsrMemberKey(policyName, roleName string) string {
	return policyName + "|" + roleName
}

// handle dispatches membership requests by HTTP method.
func (s *mapDsrMembershipsStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet returns the pair when found, or empty list + 200 on miss.
// Both policy_names and role_names are optional individually; when both are provided,
// the response is filtered to that exact pair.
func (s *mapDsrMembershipsStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "role_names"}) {
		return
	}
	pName := r.URL.Query().Get("policy_names")
	rName := r.URL.Query().Get("role_names")

	s.mu.Lock()
	defer s.mu.Unlock()

	items := []client.ManagementAccessPolicyDirectoryServiceRoleMembership{}
	if pName != "" && rName != "" {
		if _, ok := s.set[dsrMemberKey(pName, rName)]; ok {
			items = append(items, client.ManagementAccessPolicyDirectoryServiceRoleMembership{
				Policy: client.NamedReference{Name: pName, ID: "map-mock"},
				Role:   client.NamedReference{Name: rName, ID: "dsr-mock"},
			})
		}
	}
	// When only one param is provided we return an empty list (tests always specify both).
	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost is idempotent: creates the pair if absent, returns 200 with the pair either way.
// This resolves Q3 from 50-CONTEXT.md — Terraform replays never produce 409 conflicts.
func (s *mapDsrMembershipsStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "role_names"}) {
		return
	}
	pName, okP := RequireQueryParam(w, r, "policy_names")
	if !okP {
		return
	}
	rName, okR := RequireQueryParam(w, r, "role_names")
	if !okR {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create-or-return: idempotent.
	s.set[dsrMemberKey(pName, rName)] = struct{}{}
	items := []client.ManagementAccessPolicyDirectoryServiceRoleMembership{{
		Policy: client.NamedReference{Name: pName, ID: "map-mock"},
		Role:   client.NamedReference{Name: rName, ID: "dsr-mock"},
	}}
	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleDelete removes the pair from the set. Idempotent — missing pair is silently ignored.
func (s *mapDsrMembershipsStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "role_names"}) {
		return
	}
	pName, okP := RequireQueryParam(w, r, "policy_names")
	if !okP {
		return
	}
	rName, okR := RequireQueryParam(w, r, "role_names")
	if !okR {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.set, dsrMemberKey(pName, rName))
	fmt.Fprint(w, "")
}

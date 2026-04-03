package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// tlsPolicyStore is the thread-safe in-memory state for TLS policy handlers.
type tlsPolicyStore struct {
	mu      sync.Mutex
	policies map[string]*client.TlsPolicy        // keyed by policy name
	members  map[string][]client.TlsPolicyMember // keyed by policy name
	nextID  int
}

// RegisterTlsPolicyHandlers registers CRUD handlers for:
//   - /api/2.22/tls-policies (policy CRUD)
//   - /api/2.22/tls-policies/members (member GET list)
//   - /api/2.22/network-interfaces/tls-policies (member POST/DELETE)
//
// Returns the store so tests can call Seed and SeedMember.
func RegisterTlsPolicyHandlers(mux *http.ServeMux) *tlsPolicyStore {
	store := &tlsPolicyStore{
		policies: make(map[string]*client.TlsPolicy),
		members:  make(map[string][]client.TlsPolicyMember),
	}
	// Register member endpoints before policy endpoint to avoid ServeMux prefix collision.
	mux.HandleFunc("/api/2.22/tls-policies/members", store.handleMember)
	mux.HandleFunc("/api/2.22/tls-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/network-interfaces/tls-policies", store.handleNITlsPolicies)
	return store
}

// Seed adds a TLS policy directly to the store for test setup.
func (s *tlsPolicyStore) Seed(policy *client.TlsPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[policy.Name] = policy
}

// SeedMember adds a TLS policy member directly to the store for test setup.
func (s *tlsPolicyStore) SeedMember(policyName string, member client.TlsPolicyMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[policyName] = append(s.members[policyName], member)
}

// handlePolicy dispatches TLS policy requests by HTTP method.
func (s *tlsPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handlePolicyGet(w, r)
	case http.MethodPost:
		s.handlePolicyPost(w, r)
	case http.MethodPatch:
		s.handlePolicyPatch(w, r)
	case http.MethodDelete:
		s.handlePolicyDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePolicyGet handles GET /api/2.22/tls-policies.
// When ?names= filter matches nothing, returns empty list with HTTP 200 (not 404).
func (s *tlsPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "names", "effective", "purity_defined"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	namesFilter := q.Get("names")

	var items []client.TlsPolicy

	if namesFilter != "" {
		if policy, ok := s.policies[namesFilter]; ok {
			items = append(items, *policy)
		}
	} else {
		for _, policy := range s.policies {
			items = append(items, *policy)
		}
	}

	if items == nil {
		items = []client.TlsPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePolicyPost handles POST /api/2.22/tls-policies.
func (s *tlsPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.TlsPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("TLS policy %q already exists", name))
		return
	}

	s.nextID++
	id := fmt.Sprintf("tls-%d", s.nextID)

	policy := &client.TlsPolicy{
		ID:                               id,
		Name:                             name,
		ApplianceCertificate:             body.ApplianceCertificate,
		ClientCertificatesRequired:       body.ClientCertificatesRequired,
		DisabledTlsCiphers:               body.DisabledTlsCiphers,
		Enabled:                          body.Enabled,
		EnabledTlsCiphers:                body.EnabledTlsCiphers,
		IsLocal:                          true,
		MinTlsVersion:                    body.MinTlsVersion,
		PolicyType:                       "tls",
		TrustedClientCertificateAuthority: body.TrustedClientCertificateAuthority,
		VerifyClientCertificateTrust:     body.VerifyClientCertificateTrust,
	}

	s.policies[name] = policy

	WriteJSONListResponse(w, http.StatusOK, []client.TlsPolicy{*policy})
}

// handlePolicyPatch handles PATCH /api/2.22/tls-policies.
func (s *tlsPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.TlsPolicyPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("TLS policy %q not found", name))
		return
	}

	if body.ApplianceCertificate != nil {
		policy.ApplianceCertificate = *body.ApplianceCertificate
	}
	if body.ClientCertificatesRequired != nil {
		policy.ClientCertificatesRequired = *body.ClientCertificatesRequired
	}
	if body.DisabledTlsCiphers != nil {
		policy.DisabledTlsCiphers = *body.DisabledTlsCiphers
	}
	if body.Enabled != nil {
		policy.Enabled = *body.Enabled
	}
	if body.EnabledTlsCiphers != nil {
		policy.EnabledTlsCiphers = *body.EnabledTlsCiphers
	}
	if body.MinTlsVersion != nil {
		policy.MinTlsVersion = *body.MinTlsVersion
	}
	if body.TrustedClientCertificateAuthority != nil {
		policy.TrustedClientCertificateAuthority = *body.TrustedClientCertificateAuthority
	}
	if body.VerifyClientCertificateTrust != nil {
		policy.VerifyClientCertificateTrust = *body.VerifyClientCertificateTrust
	}

	WriteJSONListResponse(w, http.StatusOK, []client.TlsPolicy{*policy})
}

// handlePolicyDelete handles DELETE /api/2.22/tls-policies.
func (s *tlsPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("TLS policy %q not found", name))
		return
	}

	delete(s.policies, name)
	delete(s.members, name)

	w.WriteHeader(http.StatusOK)
}

// handleMember dispatches GET /api/2.22/tls-policies/members requests.
func (s *tlsPolicyStore) handleMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.handleMemberGet(w, r)
}

// handleMemberGet handles GET /api/2.22/tls-policies/members.
func (s *tlsPolicyStore) handleMemberGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	policyNamesFilter := q.Get("policy_names")

	var items []client.TlsPolicyMember

	if policyNamesFilter != "" {
		if members, ok := s.members[policyNamesFilter]; ok {
			items = append(items, members...)
		}
	} else {
		for _, members := range s.members {
			items = append(items, members...)
		}
	}

	if items == nil {
		items = []client.TlsPolicyMember{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleNITlsPolicies dispatches POST/DELETE /api/2.22/network-interfaces/tls-policies requests.
func (s *tlsPolicyStore) handleNITlsPolicies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleMemberPost(w, r)
	case http.MethodDelete:
		s.handleMemberDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMemberPost handles POST /api/2.22/network-interfaces/tls-policies.
func (s *tlsPolicyStore) handleMemberPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names"}) {
		return
	}

	policyName, ok := RequireQueryParam(w, r, "policy_names")
	if !ok {
		return
	}

	memberName := r.URL.Query().Get("member_names")
	if memberName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names query parameter is required")
		return
	}

	// Drain body (may be empty or {}).
	_, _ = io.ReadAll(r.Body)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policyName]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("TLS policy %q not found", policyName))
		return
	}

	member := client.TlsPolicyMember{
		Policy: client.NamedReference{Name: policyName},
		Member: client.NamedReference{Name: memberName},
	}

	s.members[policyName] = append(s.members[policyName], member)

	WriteJSONListResponse(w, http.StatusOK, []client.TlsPolicyMember{member})
}

// TlsPolicyStoreFacade exposes Seed and SeedMember on the unexported store for use in tests
// from packages outside the handlers package.
type TlsPolicyStoreFacade struct {
	store *tlsPolicyStore
}

// NewTlsPolicyStoreFacade wraps the internal store so cross-package tests can call Seed and SeedMember.
func NewTlsPolicyStoreFacade(store *tlsPolicyStore) *TlsPolicyStoreFacade {
	return &TlsPolicyStoreFacade{store: store}
}

// Seed adds a TLS policy to the store.
func (f *TlsPolicyStoreFacade) Seed(policy *client.TlsPolicy) {
	f.store.Seed(policy)
}

// SeedMember adds a TLS policy member to the store.
func (f *TlsPolicyStoreFacade) SeedMember(policyName string, member client.TlsPolicyMember) {
	f.store.SeedMember(policyName, member)
}

// handleMemberDelete handles DELETE /api/2.22/network-interfaces/tls-policies.
func (s *tlsPolicyStore) handleMemberDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names"}) {
		return
	}

	policyName, ok := RequireQueryParam(w, r, "policy_names")
	if !ok {
		return
	}

	memberName := r.URL.Query().Get("member_names")
	if memberName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	members, exists := s.members[policyName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("TLS policy %q not found or has no members", policyName))
		return
	}

	found := false
	for i, m := range members {
		if m.Member.Name == memberName {
			s.members[policyName] = append(members[:i], members[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("member %q not found in TLS policy %q", memberName, policyName))
		return
	}

	w.WriteHeader(http.StatusOK)
}

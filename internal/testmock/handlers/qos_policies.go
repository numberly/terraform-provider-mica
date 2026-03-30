package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// qosPolicyStore is the thread-safe in-memory state for QoS policy handlers.
type qosPolicyStore struct {
	mu       sync.Mutex
	policies map[string]*client.QosPolicy          // keyed by policy name
	members  map[string][]client.QosPolicyMember    // keyed by policy name
	nextID   int
}

// RegisterQosPolicyHandlers registers CRUD handlers for
// /api/2.22/qos-policies and /api/2.22/qos-policies/members
// against the provided ServeMux. The returned store pointer can be used for test setup.
func RegisterQosPolicyHandlers(mux *http.ServeMux) *qosPolicyStore {
	store := &qosPolicyStore{
		policies: make(map[string]*client.QosPolicy),
		members:  make(map[string][]client.QosPolicyMember),
	}
	mux.HandleFunc("/api/2.22/qos-policies/members", store.handleMember)
	mux.HandleFunc("/api/2.22/qos-policies", store.handlePolicy)
	return store
}

// Seed adds a QoS policy directly to the store for test setup.
func (s *qosPolicyStore) Seed(policy *client.QosPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[policy.Name] = policy
}

// SeedMember adds a QoS policy member directly to the store for test setup.
func (s *qosPolicyStore) SeedMember(policyName string, member client.QosPolicyMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[policyName] = append(s.members[policyName], member)
}

// handlePolicy dispatches QoS policy requests by HTTP method.
func (s *qosPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
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

// handlePolicyGet handles GET /api/2.22/qos-policies.
func (s *qosPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	namesFilter := q.Get("names")

	var items []client.QosPolicy

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
		items = []client.QosPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePolicyPost handles POST /api/2.22/qos-policies.
func (s *qosPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	var body client.QosPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if body.Name == "" {
		WriteJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[body.Name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("QoS policy %q already exists", body.Name))
		return
	}

	s.nextID++
	id := fmt.Sprintf("qos-%d", s.nextID)

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	policy := &client.QosPolicy{
		ID:                  id,
		Name:                body.Name,
		Enabled:             enabled,
		IsLocal:             true,
		MaxTotalBytesPerSec: body.MaxTotalBytesPerSec,
		MaxTotalOpsPerSec:   body.MaxTotalOpsPerSec,
		PolicyType:          "bandwidth-limit",
	}

	s.policies[body.Name] = policy

	WriteJSONListResponse(w, http.StatusOK, []client.QosPolicy{*policy})
}

// handlePolicyPatch handles PATCH /api/2.22/qos-policies.
func (s *qosPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.QosPolicyPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("QoS policy %q not found", name))
		return
	}

	if body.Enabled != nil {
		policy.Enabled = *body.Enabled
	}
	if body.MaxTotalBytesPerSec != nil {
		policy.MaxTotalBytesPerSec = *body.MaxTotalBytesPerSec
	}
	if body.MaxTotalOpsPerSec != nil {
		policy.MaxTotalOpsPerSec = *body.MaxTotalOpsPerSec
	}
	if body.Name != nil {
		oldName := policy.Name
		policy.Name = *body.Name
		if *body.Name != oldName {
			s.policies[*body.Name] = policy
			delete(s.policies, oldName)
			// Move members to the new key.
			if members, ok := s.members[oldName]; ok {
				s.members[*body.Name] = members
				delete(s.members, oldName)
			}
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.QosPolicy{*policy})
}

// handlePolicyDelete handles DELETE /api/2.22/qos-policies.
func (s *qosPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("QoS policy %q not found", name))
		return
	}

	delete(s.policies, name)
	delete(s.members, name)

	w.WriteHeader(http.StatusOK)
}

// handleMember dispatches QoS policy member requests by HTTP method.
func (s *qosPolicyStore) handleMember(w http.ResponseWriter, r *http.Request) {
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

// handleMemberGet handles GET /api/2.22/qos-policies/members.
func (s *qosPolicyStore) handleMemberGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names", "member_types"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	policyNamesFilter := q.Get("policy_names")

	var items []client.QosPolicyMember

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
		items = []client.QosPolicyMember{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleMemberPost handles POST /api/2.22/qos-policies/members.
func (s *qosPolicyStore) handleMemberPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_types"}) {
		return
	}

	policyName, ok := RequireQueryParam(w, r, "policy_names")
	if !ok {
		return
	}

	var body client.QosPolicyMemberPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policyName]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("QoS policy %q not found", policyName))
		return
	}

	member := client.QosPolicyMember{
		Member: body.Member,
		Policy: client.NamedReference{Name: policyName},
	}

	s.members[policyName] = append(s.members[policyName], member)

	WriteJSONListResponse(w, http.StatusOK, []client.QosPolicyMember{member})
}

// handleMemberDelete handles DELETE /api/2.22/qos-policies/members.
func (s *qosPolicyStore) handleMemberDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names", "member_types"}) {
		return
	}

	policyName, ok := RequireQueryParam(w, r, "policy_names")
	if !ok {
		return
	}

	q := r.URL.Query()
	memberName := q.Get("member_names")
	if memberName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	members, exists := s.members[policyName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("QoS policy %q not found or has no members", policyName))
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("member %q not found in QoS policy %q", memberName, policyName))
		return
	}

	w.WriteHeader(http.StatusOK)
}

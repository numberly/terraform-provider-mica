package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// auditObjectStorePolicyStore is the thread-safe in-memory state for audit object store policy handlers.
type auditObjectStorePolicyStore struct {
	mu      sync.Mutex
	byName  map[string]*client.AuditObjectStorePolicy
	members map[string][]client.AuditObjectStorePolicyMember // keyed by policy name
	nextID  int
}

// RegisterAuditObjectStorePolicyHandlers registers CRUD handlers for
// /api/2.22/audit-object-store-policies against the provided ServeMux.
// The returned store pointer can be used for test setup via Seed.
func RegisterAuditObjectStorePolicyHandlers(mux *http.ServeMux) *auditObjectStorePolicyStore {
	store := &auditObjectStorePolicyStore{
		byName:  make(map[string]*client.AuditObjectStorePolicy),
		members: make(map[string][]client.AuditObjectStorePolicyMember),
		nextID:  1,
	}
	mux.HandleFunc("/api/2.22/audit-object-store-policies/members", store.handleMember)
	mux.HandleFunc("/api/2.22/audit-object-store-policies", store.handle)
	return store
}

// Seed adds an audit object store policy directly to the store for test setup.
func (s *auditObjectStorePolicyStore) Seed(policy *client.AuditObjectStorePolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[policy.Name] = policy
}

// SeedMember adds a policy member directly to the store for test setup.
func (s *auditObjectStorePolicyStore) SeedMember(policyName string, member client.AuditObjectStorePolicyMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[policyName] = append(s.members[policyName], member)
}

// RemoveMember removes a member from the store for test setup (simulates out-of-band deletion).
func (s *auditObjectStorePolicyStore) RemoveMember(policyName, memberName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	members := s.members[policyName]
	for i, m := range members {
		if m.Member.Name == memberName {
			s.members[policyName] = append(members[:i], members[i+1:]...)
			return
		}
	}
}

func (s *auditObjectStorePolicyStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/audit-object-store-policies.
// Returns empty list (HTTP 200) when not found — matches real FlashBlade API behavior.
func (s *auditObjectStorePolicyStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.AuditObjectStorePolicy

	if namesFilter != "" {
		policy, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *policy)
		}
	} else {
		for _, policy := range s.byName {
			items = append(items, *policy)
		}
	}

	if items == nil {
		items = []client.AuditObjectStorePolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/audit-object-store-policies?names={name}.
func (s *auditObjectStorePolicyStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.AuditObjectStorePolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("audit object store policy %q already exists", name))
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	logTargets := body.LogTargets
	if logTargets == nil {
		logTargets = []client.NamedReference{}
	}

	policy := &client.AuditObjectStorePolicy{
		ID:         fmt.Sprintf("audit-policy-%d", s.nextID),
		Name:       name,
		Enabled:    enabled,
		IsLocal:    true,
		PolicyType: "audit",
		LogTargets: logTargets,
	}
	s.nextID++

	s.byName[name] = policy

	WriteJSONListResponse(w, http.StatusOK, []client.AuditObjectStorePolicy{*policy})
}

// handlePatch handles PATCH /api/2.22/audit-object-store-policies?names={name}.
func (s *auditObjectStorePolicyStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("audit object store policy %q not found", name))
		return
	}

	if v, ok := rawPatch["enabled"]; ok {
		var enabled bool
		if err := json.Unmarshal(v, &enabled); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid enabled field")
			return
		}
		policy.Enabled = enabled
	}

	if v, ok := rawPatch["log_targets"]; ok {
		var logTargets []client.NamedReference
		if err := json.Unmarshal(v, &logTargets); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid log_targets field")
			return
		}
		if logTargets == nil {
			logTargets = []client.NamedReference{}
		}
		policy.LogTargets = logTargets
	}

	WriteJSONListResponse(w, http.StatusOK, []client.AuditObjectStorePolicy{*policy})
}

// handleDelete handles DELETE /api/2.22/audit-object-store-policies?names={name}.
func (s *auditObjectStorePolicyStore) handleDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("audit object store policy %q not found", name))
		return
	}

	delete(s.byName, name)
	delete(s.members, name)

	w.WriteHeader(http.StatusOK)
}

// ---------- member handlers ---------------------------------------------------

// handleMember dispatches member requests by HTTP method.
func (s *auditObjectStorePolicyStore) handleMember(w http.ResponseWriter, r *http.Request) {
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

// handleMemberGet handles GET /api/2.22/audit-object-store-policies/members.
func (s *auditObjectStorePolicyStore) handleMemberGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names", "member_ids"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policyNamesFilter := r.URL.Query().Get("policy_names")

	var items []client.AuditObjectStorePolicyMember

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
		items = []client.AuditObjectStorePolicyMember{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleMemberPost handles POST /api/2.22/audit-object-store-policies/members.
func (s *auditObjectStorePolicyStore) handleMemberPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names", "member_ids"}) {
		return
	}

	policyName := r.URL.Query().Get("policy_names")
	memberName := r.URL.Query().Get("member_names")
	if policyName == "" || memberName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names and member_names query parameters are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[policyName]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("audit object store policy %q not found", policyName))
		return
	}

	for _, m := range s.members[policyName] {
		if m.Member.Name == memberName {
			WriteJSONError(w, http.StatusConflict, fmt.Sprintf("member %q already in policy %q", memberName, policyName))
			return
		}
	}

	member := client.AuditObjectStorePolicyMember{
		Member: client.NamedReference{Name: memberName},
		Policy: client.NamedReference{Name: policyName},
	}

	s.members[policyName] = append(s.members[policyName], member)

	WriteJSONListResponse(w, http.StatusOK, []client.AuditObjectStorePolicyMember{member})
}

// handleMemberDelete handles DELETE /api/2.22/audit-object-store-policies/members.
func (s *auditObjectStorePolicyStore) handleMemberDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names", "policy_ids", "member_names", "member_ids"}) {
		return
	}

	policyName := r.URL.Query().Get("policy_names")
	memberName := r.URL.Query().Get("member_names")
	if policyName == "" || memberName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names and member_names query parameters are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	members, exists := s.members[policyName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("policy %q not found or has no members", policyName))
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("member %q not found in policy %q", memberName, policyName))
		return
	}

	w.WriteHeader(http.StatusOK)
}

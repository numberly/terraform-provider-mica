package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// snapshotPolicyStore is the thread-safe in-memory state for snapshot policy handlers.
// Rules are embedded in the policy object — there is no separate rules endpoint.
type snapshotPolicyStore struct {
	mu       sync.Mutex
	policies map[string]*client.SnapshotPolicy // policyName -> policy (includes Rules array)
}

// RegisterSnapshotPolicyHandlers registers CRUD handlers for snapshot policies.
// Returns the store pointer so resource tests can cross-reference state.
func RegisterSnapshotPolicyHandlers(mux *http.ServeMux) *snapshotPolicyStore {
	store := &snapshotPolicyStore{
		policies: make(map[string]*client.SnapshotPolicy),
	}
	mux.HandleFunc("/api/2.22/policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/policies/file-systems", store.handleFileSystems)
	return store
}

// handlePolicy dispatches snapshot policy requests by HTTP method.
func (s *snapshotPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
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

// handleFileSystems handles GET /api/2.22/policies/file-systems?policy_names={name}.
// Returns the list of file systems attached to the policy (empty in mock by default).
func (s *snapshotPolicyStore) handleFileSystems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Return an empty list — the mock does not track file system attachments.
	WriteJSONListResponse(w, http.StatusOK, []client.PolicyMember{})
}

// handlePolicyGet handles GET /api/2.22/policies with optional ?names= param.
func (s *snapshotPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.SnapshotPolicy

	if namesFilter != "" {
		policy, ok := s.policies[namesFilter]
		if ok {
			items = append(items, *policy)
		}
	} else {
		for _, policy := range s.policies {
			items = append(items, *policy)
		}
	}

	if items == nil {
		items = []client.SnapshotPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePolicyPost handles POST /api/2.22/policies?names={name}.
// Accepts optional inline rules in the body for creation-time rule setup.
func (s *snapshotPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.SnapshotPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("snapshot policy %q already exists", name))
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	policy := &client.SnapshotPolicy{
		ID:         uuid.New().String(),
		Name:       name,
		Enabled:    enabled,
		IsLocal:    true,
		PolicyType: "snapshot",
	}

	// Handle inline rules provided at creation time.
	for _, rulePost := range body.Rules {
		ruleName := "snap-rule-" + uuid.New().String()[:8]
		policy.Rules = append(policy.Rules, client.SnapshotPolicyRuleInPolicy{
			Name:       ruleName,
			AtTime:     rulePost.AtTime,
			Every:      rulePost.Every,
			KeepFor:    rulePost.KeepFor,
			Suffix:     rulePost.Suffix,
			ClientName: rulePost.ClientName,
		})
	}

	s.policies[name] = policy

	WriteJSONListResponse(w, http.StatusOK, []client.SnapshotPolicy{*policy})
}

// handlePolicyPatch handles PATCH /api/2.22/policies?names={name}.
// Supports: enabled update, add_rules (appends rules), remove_rules (removes by name).
// Name is read-only for snapshot policies and is silently ignored if provided.
func (s *snapshotPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("snapshot policy %q not found", name))
		return
	}

	// Use a raw map for true PATCH semantics.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Update enabled field if present.
	if v, ok := rawPatch["enabled"]; ok {
		var enabled bool
		if err := json.Unmarshal(v, &enabled); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid enabled field")
			return
		}
		policy.Enabled = enabled
	}

	// Process remove_rules: remove matching rules by name.
	if v, ok := rawPatch["remove_rules"]; ok {
		var removeRules []client.SnapshotPolicyRuleRemove
		if err := json.Unmarshal(v, &removeRules); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid remove_rules field")
			return
		}
		removeSet := make(map[string]bool)
		for _, r := range removeRules {
			removeSet[r.Name] = true
		}
		var remaining []client.SnapshotPolicyRuleInPolicy
		for _, r := range policy.Rules {
			if !removeSet[r.Name] {
				remaining = append(remaining, r)
			}
		}
		policy.Rules = remaining
	}

	// Process add_rules: append new rules with auto-generated names.
	if v, ok := rawPatch["add_rules"]; ok {
		var addRules []client.SnapshotPolicyRulePost
		if err := json.Unmarshal(v, &addRules); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid add_rules field")
			return
		}
		for _, rulePost := range addRules {
			ruleName := "snap-rule-" + uuid.New().String()[:8]
			policy.Rules = append(policy.Rules, client.SnapshotPolicyRuleInPolicy{
				Name:       ruleName,
				AtTime:     rulePost.AtTime,
				Every:      rulePost.Every,
				KeepFor:    rulePost.KeepFor,
				Suffix:     rulePost.Suffix,
				ClientName: rulePost.ClientName,
			})
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.SnapshotPolicy{*policy})
}

// handlePolicyDelete handles DELETE /api/2.22/policies?names={name}.
func (s *snapshotPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("snapshot policy %q not found", name))
		return
	}

	delete(s.policies, name)

	w.WriteHeader(http.StatusOK)
}

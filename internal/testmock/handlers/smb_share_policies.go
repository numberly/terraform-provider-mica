package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// smbSharePolicyStore is the thread-safe in-memory state for SMB share policy handlers.
type smbSharePolicyStore struct {
	mu       sync.Mutex
	policies map[string]*client.SmbSharePolicy              // policyName -> policy
	rules    map[string]map[string]*client.SmbSharePolicyRule // policyName -> ruleName -> rule
}

// RegisterSmbSharePolicyHandlers registers CRUD handlers for SMB share policies and rules.
// Returns the store pointer so resource tests can cross-reference state.
func RegisterSmbSharePolicyHandlers(mux *http.ServeMux) *smbSharePolicyStore {
	store := &smbSharePolicyStore{
		policies: make(map[string]*client.SmbSharePolicy),
		rules:    make(map[string]map[string]*client.SmbSharePolicyRule),
	}
	mux.HandleFunc("/api/2.22/smb-share-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/smb-share-policies/rules", store.handleRules)
	return store
}

// handlePolicy dispatches SMB share policy requests by HTTP method.
func (s *smbSharePolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
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

// handleRules dispatches SMB share policy rule requests by HTTP method.
func (s *smbSharePolicyStore) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleRulesGet(w, r)
	case http.MethodPost:
		s.handleRulesPost(w, r)
	case http.MethodPatch:
		s.handleRulesPatch(w, r)
	case http.MethodDelete:
		s.handleRulesDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePolicyGet handles GET /api/2.22/smb-share-policies with optional ?names= param.
func (s *smbSharePolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.SmbSharePolicy

	if namesFilter != "" {
		policy, ok := s.policies[namesFilter]
		if ok {
			p := s.policyWithRules(policy)
			items = append(items, p)
		}
	} else {
		for _, policy := range s.policies {
			p := s.policyWithRules(policy)
			items = append(items, p)
		}
	}

	if items == nil {
		items = []client.SmbSharePolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// policyWithRules returns a copy of the policy with its rules populated from the rules store.
// Caller must hold s.mu.
func (s *smbSharePolicyStore) policyWithRules(policy *client.SmbSharePolicy) client.SmbSharePolicy {
	p := *policy

	policyRules := s.rules[policy.Name]
	p.Rules = nil
	for _, r := range policyRules {
		p.Rules = append(p.Rules, client.SmbSharePolicyRuleInPolicy{
			Name:        r.Name,
			Principal:   r.Principal,
			Change:      r.Change,
			FullControl: r.FullControl,
			Read:        r.Read,
		})
	}
	return p
}

// handlePolicyPost handles POST /api/2.22/smb-share-policies?names={name}.
func (s *smbSharePolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.SmbSharePolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("SMB share policy %q already exists", name))
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	policy := &client.SmbSharePolicy{
		ID:         uuid.New().String(),
		Name:       name,
		Enabled:    enabled,
		IsLocal:    true,
		PolicyType: "smb",
	}

	s.policies[name] = policy
	s.rules[name] = make(map[string]*client.SmbSharePolicyRule)

	WriteJSONListResponse(w, http.StatusOK, []client.SmbSharePolicy{s.policyWithRules(policy)})
}

// handlePolicyPatch handles PATCH /api/2.22/smb-share-policies?names={name}.
func (s *smbSharePolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB share policy %q not found", name))
		return
	}

	// Use a raw map for true PATCH semantics.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
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

	if v, ok := rawPatch["name"]; ok {
		var newName string
		if err := json.Unmarshal(v, &newName); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid name field")
			return
		}
		if newName != name {
			delete(s.policies, name)
			policy.Name = newName
			s.policies[newName] = policy

			// Move rules to new policy name key.
			if ruleMap, exists := s.rules[name]; exists {
				s.rules[newName] = ruleMap
				delete(s.rules, name)
				// Update policy reference in each rule.
				for _, rule := range ruleMap {
					rule.Policy = client.NamedReference{Name: newName}
				}
			}
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.SmbSharePolicy{s.policyWithRules(policy)})
}

// handlePolicyDelete handles DELETE /api/2.22/smb-share-policies?names={name}.
func (s *smbSharePolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB share policy %q not found", name))
		return
	}

	delete(s.policies, name)
	delete(s.rules, name)

	w.WriteHeader(http.StatusOK)
}

// handleRulesGet handles GET /api/2.22/smb-share-policies/rules.
// Filters by ?policy_names= and optionally ?names=.
func (s *smbSharePolicyStore) handleRulesGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	policyName := r.URL.Query().Get("policy_names")
	ruleName := r.URL.Query().Get("names")

	var items []client.SmbSharePolicyRule

	if policyName != "" {
		policyRules, ok := s.rules[policyName]
		if ok {
			if ruleName != "" {
				if rule, exists := policyRules[ruleName]; exists {
					items = append(items, *rule)
				}
			} else {
				for _, rule := range policyRules {
					items = append(items, *rule)
				}
			}
		}
	} else {
		for _, policyRules := range s.rules {
			for _, rule := range policyRules {
				items = append(items, *rule)
			}
		}
	}

	if items == nil {
		items = []client.SmbSharePolicyRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleRulesPost handles POST /api/2.22/smb-share-policies/rules?policy_names={name}.
func (s *smbSharePolicyStore) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	policyName := r.URL.Query().Get("policy_names")
	if policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names query parameter is required for POST")
		return
	}

	var body client.SmbSharePolicyRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[policyName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB share policy %q not found", policyName))
		return
	}

	id := uuid.New().String()
	ruleName := "smb-rule-" + id[:8]

	rule := &client.SmbSharePolicyRule{
		ID:          id,
		Name:        ruleName,
		Policy:      client.NamedReference{Name: policyName},
		Principal:   body.Principal,
		Change:      body.Change,
		FullControl: body.FullControl,
		Read:        body.Read,
	}

	if s.rules[policyName] == nil {
		s.rules[policyName] = make(map[string]*client.SmbSharePolicyRule)
	}
	s.rules[policyName][ruleName] = rule

	WriteJSONListResponse(w, http.StatusOK, []client.SmbSharePolicyRule{*rule})
}

// handleRulesPatch handles PATCH /api/2.22/smb-share-policies/rules?names={name}&policy_names={policy}.
func (s *smbSharePolicyStore) handleRulesPatch(w http.ResponseWriter, r *http.Request) {
	ruleName := r.URL.Query().Get("names")
	policyName := r.URL.Query().Get("policy_names")
	if ruleName == "" || policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "names and policy_names query parameters are required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policyRules, ok := s.rules[policyName]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB share policy %q not found", policyName))
		return
	}

	rule, ok := policyRules[ruleName]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB share policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	// Use a raw map for true PATCH semantics.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["principal"]; ok {
		var principal string
		if err := json.Unmarshal(v, &principal); err == nil {
			rule.Principal = principal
		}
	}
	if v, ok := rawPatch["change"]; ok {
		var change string
		if err := json.Unmarshal(v, &change); err == nil {
			rule.Change = change
		}
	}
	if v, ok := rawPatch["full_control"]; ok {
		var fullControl string
		if err := json.Unmarshal(v, &fullControl); err == nil {
			rule.FullControl = fullControl
		}
	}
	if v, ok := rawPatch["read"]; ok {
		var read string
		if err := json.Unmarshal(v, &read); err == nil {
			rule.Read = read
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.SmbSharePolicyRule{*rule})
}

// handleRulesDelete handles DELETE /api/2.22/smb-share-policies/rules?names={name}&policy_names={policy}.
func (s *smbSharePolicyStore) handleRulesDelete(w http.ResponseWriter, r *http.Request) {
	ruleName := r.URL.Query().Get("names")
	policyName := r.URL.Query().Get("policy_names")
	if ruleName == "" || policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "names and policy_names query parameters are required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policyRules, ok := s.rules[policyName]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB share policy %q not found", policyName))
		return
	}

	if _, ok := policyRules[ruleName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB share policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	delete(policyRules, ruleName)

	w.WriteHeader(http.StatusOK)
}

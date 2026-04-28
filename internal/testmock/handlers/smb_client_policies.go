package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/terraform-provider-mica/internal/client"
)

// smbClientPolicyStore is the thread-safe in-memory state for SMB client policy handlers.
type smbClientPolicyStore struct {
	mu       sync.Mutex
	policies map[string]*client.SmbClientPolicy                // policyName -> policy
	rules    map[string]map[string]*client.SmbClientPolicyRule // policyName -> ruleName -> rule
}

// RegisterSmbClientPolicyHandlers registers CRUD handlers for SMB client policies and rules.
// Returns the store pointer so resource tests can cross-reference state.
func RegisterSmbClientPolicyHandlers(mux *http.ServeMux) *smbClientPolicyStore {
	store := &smbClientPolicyStore{
		policies: make(map[string]*client.SmbClientPolicy),
		rules:    make(map[string]map[string]*client.SmbClientPolicyRule),
	}
	mux.HandleFunc("/api/2.22/smb-client-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/smb-client-policies/rules", store.handleRules)
	return store
}

// handlePolicy dispatches SMB client policy requests by HTTP method.
func (s *smbClientPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
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

// handleRules dispatches SMB client policy rule requests by HTTP method.
func (s *smbClientPolicyStore) handleRules(w http.ResponseWriter, r *http.Request) {
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

// handlePolicyGet handles GET /api/2.22/smb-client-policies with optional ?names= param.
func (s *smbClientPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.SmbClientPolicy

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
		items = []client.SmbClientPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// policyWithRules returns a copy of the policy with its rules populated from the rules store.
// Caller must hold s.mu.
func (s *smbClientPolicyStore) policyWithRules(policy *client.SmbClientPolicy) client.SmbClientPolicy {
	p := *policy

	policyRules := s.rules[policy.Name]
	p.Rules = nil
	for _, r := range policyRules {
		p.Rules = append(p.Rules, client.SmbClientPolicyRuleInPolicy{
			Name:       r.Name,
			Index:      r.Index,
			Client:     r.Client,
			Encryption: r.Encryption,
			Permission: r.Permission,
		})
	}
	return p
}

// handlePolicyPost handles POST /api/2.22/smb-client-policies?names={name}.
func (s *smbClientPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.SmbClientPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("SMB client policy %q already exists", name))
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	abeEnabled := false
	if body.AccessBasedEnumerationEnabled != nil {
		abeEnabled = *body.AccessBasedEnumerationEnabled
	}

	policy := &client.SmbClientPolicy{
		ID:                            uuid.New().String(),
		Name:                          name,
		Enabled:                       enabled,
		IsLocal:                       true,
		PolicyType:                    "smb",
		Version:                       "1",
		AccessBasedEnumerationEnabled: abeEnabled,
	}

	s.policies[name] = policy
	s.rules[name] = make(map[string]*client.SmbClientPolicyRule)

	WriteJSONListResponse(w, http.StatusOK, []client.SmbClientPolicy{s.policyWithRules(policy)})
}

// handlePolicyPatch handles PATCH /api/2.22/smb-client-policies?names={name}.
func (s *smbClientPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB client policy %q not found", name))
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

	if v, ok := rawPatch["access_based_enumeration_enabled"]; ok {
		var abeEnabled bool
		if err := json.Unmarshal(v, &abeEnabled); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid access_based_enumeration_enabled field")
			return
		}
		policy.AccessBasedEnumerationEnabled = abeEnabled
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

	WriteJSONListResponse(w, http.StatusOK, []client.SmbClientPolicy{s.policyWithRules(policy)})
}

// handlePolicyDelete handles DELETE /api/2.22/smb-client-policies?names={name}.
func (s *smbClientPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB client policy %q not found", name))
		return
	}

	delete(s.policies, name)
	delete(s.rules, name)

	w.WriteHeader(http.StatusOK)
}

// handleRulesGet handles GET /api/2.22/smb-client-policies/rules.
// Filters by ?policy_names= and optionally ?names=.
func (s *smbClientPolicyStore) handleRulesGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	policyName := r.URL.Query().Get("policy_names")
	ruleName := r.URL.Query().Get("names")

	var items []client.SmbClientPolicyRule

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
		items = []client.SmbClientPolicyRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleRulesPost handles POST /api/2.22/smb-client-policies/rules?policy_names={name}.
func (s *smbClientPolicyStore) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	policyName := r.URL.Query().Get("policy_names")
	if policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names query parameter is required for POST")
		return
	}

	var body client.SmbClientPolicyRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[policyName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB client policy %q not found", policyName))
		return
	}

	id := uuid.New().String()
	ruleName := "smb-client-rule-" + id[:8]

	index := len(s.rules[policyName]) + 1
	if body.Index != nil {
		index = *body.Index
	}

	rule := &client.SmbClientPolicyRule{
		ID:         id,
		Name:       ruleName,
		Index:      index,
		Policy:     client.NamedReference{Name: policyName},
		Client:     body.Client,
		Encryption: body.Encryption,
		Permission: body.Permission,
	}

	if s.rules[policyName] == nil {
		s.rules[policyName] = make(map[string]*client.SmbClientPolicyRule)
	}
	s.rules[policyName][ruleName] = rule

	WriteJSONListResponse(w, http.StatusOK, []client.SmbClientPolicyRule{*rule})
}

// handleRulesPatch handles PATCH /api/2.22/smb-client-policies/rules?names={name}&policy_names={policy}.
func (s *smbClientPolicyStore) handleRulesPatch(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB client policy %q not found", policyName))
		return
	}

	rule, ok := policyRules[ruleName]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB client policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	// Use a raw map for true PATCH semantics.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["client"]; ok {
		var c string
		if err := json.Unmarshal(v, &c); err == nil {
			rule.Client = c
		}
	}
	if v, ok := rawPatch["encryption"]; ok {
		var enc string
		if err := json.Unmarshal(v, &enc); err == nil {
			rule.Encryption = enc
		}
	}
	if v, ok := rawPatch["permission"]; ok {
		var perm string
		if err := json.Unmarshal(v, &perm); err == nil {
			rule.Permission = perm
		}
	}
	if v, ok := rawPatch["index"]; ok {
		var idx int
		if err := json.Unmarshal(v, &idx); err == nil {
			rule.Index = idx
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.SmbClientPolicyRule{*rule})
}

// handleRulesDelete handles DELETE /api/2.22/smb-client-policies/rules?names={name}&policy_names={policy}.
func (s *smbClientPolicyStore) handleRulesDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB client policy %q not found", policyName))
		return
	}

	if _, ok := policyRules[ruleName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("SMB client policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	delete(policyRules, ruleName)

	w.WriteHeader(http.StatusOK)
}

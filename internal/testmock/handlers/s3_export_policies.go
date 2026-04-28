package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/numberly/terraform-provider-mica/internal/client"
)

// s3ExportPolicyStore is the thread-safe in-memory state for S3 export policy handlers.
type s3ExportPolicyStore struct {
	mu            sync.Mutex
	policies      map[string]*client.S3ExportPolicy                // policyName -> policy
	rules         map[string]map[string]*client.S3ExportPolicyRule // policyName -> ruleName -> rule
	nextRuleIndex map[string]int                                   // policyName -> next index counter
}

// RegisterS3ExportPolicyHandlers registers CRUD handlers for S3 export policies and rules.
// Returns the store pointer so resource tests can cross-reference state.
func RegisterS3ExportPolicyHandlers(mux *http.ServeMux) *s3ExportPolicyStore {
	store := &s3ExportPolicyStore{
		policies:      make(map[string]*client.S3ExportPolicy),
		rules:         make(map[string]map[string]*client.S3ExportPolicyRule),
		nextRuleIndex: make(map[string]int),
	}
	mux.HandleFunc("/api/2.22/s3-export-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/s3-export-policies/rules", store.handleRules)
	return store
}

// handlePolicy dispatches S3 export policy requests by HTTP method.
func (s *s3ExportPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
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

// handleRules dispatches S3 export policy rule requests by HTTP method.
func (s *s3ExportPolicyStore) handleRules(w http.ResponseWriter, r *http.Request) {
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

// handlePolicyGet handles GET /api/2.22/s3-export-policies with optional ?names= param.
func (s *s3ExportPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.S3ExportPolicy

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
		items = []client.S3ExportPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePolicyPost handles POST /api/2.22/s3-export-policies?names={name}.
func (s *s3ExportPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.S3ExportPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("S3 export policy %q already exists", name))
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	policy := &client.S3ExportPolicy{
		ID:         uuid.New().String(),
		Name:       name,
		Enabled:    enabled,
		IsLocal:    true,
		PolicyType: "s3-export",
		Version:    fmt.Sprintf("%d", time.Now().UnixMilli()),
	}

	s.policies[name] = policy
	s.rules[name] = make(map[string]*client.S3ExportPolicyRule)
	s.nextRuleIndex[name] = 1

	WriteJSONListResponse(w, http.StatusOK, []client.S3ExportPolicy{*policy})
}

// handlePolicyPatch handles PATCH /api/2.22/s3-export-policies?names={name}.
func (s *s3ExportPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("S3 export policy %q not found", name))
		return
	}

	// Use a raw map for true PATCH semantics — only provided fields are updated.
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
			// Rename: update map keys for policy and rules.
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
			if idx, exists := s.nextRuleIndex[name]; exists {
				s.nextRuleIndex[newName] = idx
				delete(s.nextRuleIndex, name)
			}
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.S3ExportPolicy{*policy})
}

// handlePolicyDelete handles DELETE /api/2.22/s3-export-policies?names={name}.
func (s *s3ExportPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("S3 export policy %q not found", name))
		return
	}

	delete(s.policies, name)
	delete(s.rules, name)
	delete(s.nextRuleIndex, name)

	w.WriteHeader(http.StatusOK)
}

// handleRulesGet handles GET /api/2.22/s3-export-policies/rules.
// Filters by ?policy_names= and optionally ?names=.
func (s *s3ExportPolicyStore) handleRulesGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	policyName := r.URL.Query().Get("policy_names")
	ruleName := r.URL.Query().Get("names")

	var items []client.S3ExportPolicyRule

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

	// Sort by index for deterministic output.
	sort.Slice(items, func(i, j int) bool {
		return items[i].Index < items[j].Index
	})

	if items == nil {
		items = []client.S3ExportPolicyRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleRulesPost handles POST /api/2.22/s3-export-policies/rules?policy_names={name}.
func (s *s3ExportPolicyStore) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	policyName := r.URL.Query().Get("policy_names")
	if policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names query parameter is required for POST")
		return
	}

	var body client.S3ExportPolicyRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[policyName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("S3 export policy %q not found", policyName))
		return
	}

	// Auto-assign name and index.
	id := uuid.New().String()
	ruleName := "rule-" + id[:8]
	index := s.nextRuleIndex[policyName]
	s.nextRuleIndex[policyName] = index + 1

	rule := &client.S3ExportPolicyRule{
		ID:        id,
		Name:      ruleName,
		Index:     index,
		Policy:    client.NamedReference{Name: policyName},
		Effect:    body.Effect,
		Actions:   body.Actions,
		Resources: body.Resources,
	}

	if s.rules[policyName] == nil {
		s.rules[policyName] = make(map[string]*client.S3ExportPolicyRule)
	}
	s.rules[policyName][ruleName] = rule

	WriteJSONListResponse(w, http.StatusOK, []client.S3ExportPolicyRule{*rule})
}

// handleRulesPatch handles PATCH /api/2.22/s3-export-policies/rules?names={name}&policy_names={policy}.
func (s *s3ExportPolicyStore) handleRulesPatch(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("S3 export policy %q not found", policyName))
		return
	}

	rule, ok := policyRules[ruleName]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("S3 export policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	// Use a raw map for true PATCH semantics.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["effect"]; ok {
		var effect string
		if err := json.Unmarshal(v, &effect); err == nil {
			rule.Effect = effect
		}
	}
	if v, ok := rawPatch["actions"]; ok {
		var actions []string
		if err := json.Unmarshal(v, &actions); err == nil {
			rule.Actions = actions
		}
	}
	if v, ok := rawPatch["resources"]; ok {
		var resources []string
		if err := json.Unmarshal(v, &resources); err == nil {
			rule.Resources = resources
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.S3ExportPolicyRule{*rule})
}

// handleRulesDelete handles DELETE /api/2.22/s3-export-policies/rules?names={name}&policy_names={policy}.
func (s *s3ExportPolicyStore) handleRulesDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("S3 export policy %q not found", policyName))
		return
	}

	if _, ok := policyRules[ruleName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("S3 export policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	delete(policyRules, ruleName)

	w.WriteHeader(http.StatusOK)
}

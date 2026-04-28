package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/terraform-provider-mica/internal/client"
)

// objectStoreAccessPolicyStore is the thread-safe in-memory state for OAP handlers.
type objectStoreAccessPolicyStore struct {
	mu       sync.Mutex
	policies map[string]*client.ObjectStoreAccessPolicy           // policyName -> policy
	rules    map[string]map[string]*client.ObjectStoreAccessPolicyRule // policyName/ruleName -> rule
}

// RegisterObjectStoreAccessPolicyHandlers registers CRUD handlers for object store access policies and rules.
func RegisterObjectStoreAccessPolicyHandlers(mux *http.ServeMux) *objectStoreAccessPolicyStore {
	store := &objectStoreAccessPolicyStore{
		policies: make(map[string]*client.ObjectStoreAccessPolicy),
		rules:    make(map[string]map[string]*client.ObjectStoreAccessPolicyRule),
	}
	mux.HandleFunc("/api/2.22/object-store-access-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/object-store-access-policies/rules", store.handleRules)
	// Stub for policy-user membership checks (delete guard). Always returns empty list.
	mux.HandleFunc("/api/2.22/object-store-access-policies/object-store-users", func(w http.ResponseWriter, r *http.Request) {
		WriteJSONListResponse(w, http.StatusOK, []any{})
	})
	return store
}

// handlePolicy dispatches OAP requests by HTTP method.
func (s *objectStoreAccessPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
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

// handleRules dispatches OAP rule requests by HTTP method.
func (s *objectStoreAccessPolicyStore) handleRules(w http.ResponseWriter, r *http.Request) {
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

func (s *objectStoreAccessPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")
	var items []client.ObjectStoreAccessPolicy

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
		items = []client.ObjectStoreAccessPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// policyWithRules returns a copy of the policy with its rules populated.
// Caller must hold s.mu.
func (s *objectStoreAccessPolicyStore) policyWithRules(policy *client.ObjectStoreAccessPolicy) client.ObjectStoreAccessPolicy {
	p := *policy
	policyRules := s.rules[policy.Name]
	p.Rules = nil
	for _, rule := range policyRules {
		p.Rules = append(p.Rules, *rule)
	}
	return p
}

func (s *objectStoreAccessPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.ObjectStoreAccessPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("object store access policy %q already exists", name))
		return
	}

	policy := &client.ObjectStoreAccessPolicy{
		ID:          uuid.New().String(),
		Name:        name,
		Description: body.Description,
		Enabled:     true,
		IsLocal:     true,
		PolicyType:  "object-store-access",
	}

	s.policies[name] = policy
	s.rules[name] = make(map[string]*client.ObjectStoreAccessPolicyRule)

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccessPolicy{s.policyWithRules(policy)})
}

func (s *objectStoreAccessPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store access policy %q not found", name))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
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

			if ruleMap, exists := s.rules[name]; exists {
				s.rules[newName] = ruleMap
				delete(s.rules, name)
				for _, rule := range ruleMap {
					rule.Policy = &client.NamedReference{Name: newName}
				}
			}
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccessPolicy{s.policyWithRules(policy)})
}

func (s *objectStoreAccessPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store access policy %q not found", name))
		return
	}

	delete(s.policies, name)
	delete(s.rules, name)

	w.WriteHeader(http.StatusOK)
}

func (s *objectStoreAccessPolicyStore) handleRulesGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	policyName := r.URL.Query().Get("policy_names")
	ruleName := r.URL.Query().Get("names")

	var items []client.ObjectStoreAccessPolicyRule

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
		items = []client.ObjectStoreAccessPolicyRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *objectStoreAccessPolicyStore) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	policyName := r.URL.Query().Get("policy_names")
	ruleName := r.URL.Query().Get("names")
	if policyName == "" || ruleName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names and names query parameters are required for POST")
		return
	}

	var body client.ObjectStoreAccessPolicyRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[policyName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store access policy %q not found", policyName))
		return
	}

	if s.rules[policyName] == nil {
		s.rules[policyName] = make(map[string]*client.ObjectStoreAccessPolicyRule)
	}

	if _, exists := s.rules[policyName][ruleName]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("rule %q already exists in policy %q", ruleName, policyName))
		return
	}

	rule := &client.ObjectStoreAccessPolicyRule{
		Name:       ruleName,
		Effect:     body.Effect,
		Actions:    body.Actions,
		Conditions: body.Conditions,
		Resources:  body.Resources,
		Policy:     &client.NamedReference{Name: policyName},
	}

	s.rules[policyName][ruleName] = rule

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccessPolicyRule{*rule})
}

func (s *objectStoreAccessPolicyStore) handleRulesPatch(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store access policy %q not found", policyName))
		return
	}

	rule, ok := policyRules[ruleName]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("rule %q not found in policy %q", ruleName, policyName))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["actions"]; ok {
		var actions []string
		if err := json.Unmarshal(v, &actions); err == nil {
			rule.Actions = actions
		}
	}
	if v, ok := rawPatch["conditions"]; ok {
		rule.Conditions = v
	}
	if v, ok := rawPatch["resources"]; ok {
		var resources []string
		if err := json.Unmarshal(v, &resources); err == nil {
			rule.Resources = resources
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccessPolicyRule{*rule})
}

func (s *objectStoreAccessPolicyStore) handleRulesDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store access policy %q not found", policyName))
		return
	}

	if _, ok := policyRules[ruleName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("rule %q not found in policy %q", ruleName, policyName))
		return
	}

	delete(policyRules, ruleName)

	w.WriteHeader(http.StatusOK)
}

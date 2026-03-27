package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// networkAccessPolicyStore is the thread-safe in-memory state for NAP handlers.
type networkAccessPolicyStore struct {
	mu            sync.Mutex
	policies      map[string]*client.NetworkAccessPolicy              // policyName -> policy
	rules         map[string]map[string]*client.NetworkAccessPolicyRule // policyName -> ruleName -> rule
	nextRuleIndex map[string]int                                       // policyName -> next index counter
}

// RegisterNetworkAccessPolicyHandlers registers CRUD handlers for network access policies and rules.
// Pre-seeds a default network access policy named "default".
func RegisterNetworkAccessPolicyHandlers(mux *http.ServeMux) *networkAccessPolicyStore {
	store := &networkAccessPolicyStore{
		policies:      make(map[string]*client.NetworkAccessPolicy),
		rules:         make(map[string]map[string]*client.NetworkAccessPolicyRule),
		nextRuleIndex: make(map[string]int),
	}

	// Pre-seed a default policy (network access policies are singletons).
	defaultPolicy := &client.NetworkAccessPolicy{
		ID:         uuid.New().String(),
		Name:       "default",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "network-access",
		Version:    "1",
	}
	store.policies["default"] = defaultPolicy
	store.rules["default"] = make(map[string]*client.NetworkAccessPolicyRule)
	store.nextRuleIndex["default"] = 1

	mux.HandleFunc("/api/2.22/network-access-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/network-access-policies/rules", store.handleRules)
	return store
}

// handlePolicy dispatches NAP requests — GET and PATCH only (singletons have no POST/DELETE).
func (s *networkAccessPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handlePolicyGet(w, r)
	case http.MethodPatch:
		s.handlePolicyPatch(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRules dispatches NAP rule requests by HTTP method.
func (s *networkAccessPolicyStore) handleRules(w http.ResponseWriter, r *http.Request) {
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

func (s *networkAccessPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")
	var items []client.NetworkAccessPolicy

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
		items = []client.NetworkAccessPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *networkAccessPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("network access policy %q not found", name))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["enabled"]; ok {
		var enabled bool
		if err := json.Unmarshal(v, &enabled); err == nil {
			policy.Enabled = enabled
		}
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
			if idx, exists := s.nextRuleIndex[name]; exists {
				s.nextRuleIndex[newName] = idx
				delete(s.nextRuleIndex, name)
			}
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.NetworkAccessPolicy{*policy})
}

func (s *networkAccessPolicyStore) handleRulesGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	policyName := r.URL.Query().Get("policy_names")
	ruleName := r.URL.Query().Get("names")

	var items []client.NetworkAccessPolicyRule

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
		items = []client.NetworkAccessPolicyRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *networkAccessPolicyStore) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	policyName := r.URL.Query().Get("policy_names")
	if policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names query parameter is required for POST")
		return
	}

	var body client.NetworkAccessPolicyRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[policyName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("network access policy %q not found", policyName))
		return
	}

	id := uuid.New().String()
	ruleName := "rule-" + id[:8]
	index := s.nextRuleIndex[policyName]
	if body.Index != 0 {
		index = body.Index
	} else {
		s.nextRuleIndex[policyName] = index + 1
	}

	rule := &client.NetworkAccessPolicyRule{
		ID:         id,
		Name:       ruleName,
		Client:     body.Client,
		Effect:     body.Effect,
		Index:      index,
		Interfaces: body.Interfaces,
		Policy:     &client.NamedReference{Name: policyName},
	}

	if s.rules[policyName] == nil {
		s.rules[policyName] = make(map[string]*client.NetworkAccessPolicyRule)
	}
	s.rules[policyName][ruleName] = rule

	WriteJSONListResponse(w, http.StatusOK, []client.NetworkAccessPolicyRule{*rule})
}

func (s *networkAccessPolicyStore) handleRulesPatch(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("network access policy %q not found", policyName))
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

	if v, ok := rawPatch["client"]; ok {
		var c string
		if err := json.Unmarshal(v, &c); err == nil {
			rule.Client = c
		}
	}
	if v, ok := rawPatch["effect"]; ok {
		var effect string
		if err := json.Unmarshal(v, &effect); err == nil {
			rule.Effect = effect
		}
	}
	if v, ok := rawPatch["index"]; ok {
		var index int
		if err := json.Unmarshal(v, &index); err == nil {
			rule.Index = index
		}
	}
	if v, ok := rawPatch["interfaces"]; ok {
		var interfaces []string
		if err := json.Unmarshal(v, &interfaces); err == nil {
			rule.Interfaces = interfaces
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.NetworkAccessPolicyRule{*rule})
}

func (s *networkAccessPolicyStore) handleRulesDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("network access policy %q not found", policyName))
		return
	}

	if _, ok := policyRules[ruleName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("rule %q not found in policy %q", ruleName, policyName))
		return
	}

	delete(policyRules, ruleName)

	w.WriteHeader(http.StatusOK)
}

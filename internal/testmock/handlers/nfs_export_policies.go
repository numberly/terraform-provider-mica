package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// nfsExportPolicyStore is the thread-safe in-memory state for NFS export policy handlers.
type nfsExportPolicyStore struct {
	mu            sync.Mutex
	policies      map[string]*client.NfsExportPolicy           // policyName -> policy
	rules         map[string]map[string]*client.NfsExportPolicyRule // policyName -> ruleName -> rule
	nextRuleIndex map[string]int                                // policyName -> next index counter
}

// RegisterNfsExportPolicyHandlers registers CRUD handlers for NFS export policies and rules.
// Returns the store pointer so resource tests can cross-reference state.
func RegisterNfsExportPolicyHandlers(mux *http.ServeMux) *nfsExportPolicyStore {
	store := &nfsExportPolicyStore{
		policies:      make(map[string]*client.NfsExportPolicy),
		rules:         make(map[string]map[string]*client.NfsExportPolicyRule),
		nextRuleIndex: make(map[string]int),
	}
	mux.HandleFunc("/api/2.22/nfs-export-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/nfs-export-policies/rules", store.handleRules)
	return store
}

// handlePolicy dispatches NFS export policy requests by HTTP method.
func (s *nfsExportPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
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

// handleRules dispatches NFS export policy rule requests by HTTP method.
func (s *nfsExportPolicyStore) handleRules(w http.ResponseWriter, r *http.Request) {
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

// handlePolicyGet handles GET /api/2.22/nfs-export-policies with optional ?names= param.
func (s *nfsExportPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "ids"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.NfsExportPolicy

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
		items = []client.NfsExportPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// policyWithRules returns a copy of the policy with its rules populated and sorted by index.
// Caller must hold s.mu.
func (s *nfsExportPolicyStore) policyWithRules(policy *client.NfsExportPolicy) client.NfsExportPolicy {
	p := *policy

	policyRules := s.rules[policy.Name]
	var ruleList []*client.NfsExportPolicyRule
	for _, r := range policyRules {
		ruleList = append(ruleList, r)
	}
	sort.Slice(ruleList, func(i, j int) bool {
		return ruleList[i].Index < ruleList[j].Index
	})

	p.Rules = nil
	for _, r := range ruleList {
		p.Rules = append(p.Rules, client.NfsExportPolicyRuleInPolicy{
			Index:                     r.Index,
			Access:                    r.Access,
			Client:                    r.Client,
			Permission:                r.Permission,
			Anonuid:                   r.Anonuid,
			Anongid:                   r.Anongid,
			Atime:                     r.Atime,
			Fileid32bit:               r.Fileid32bit,
			Secure:                    r.Secure,
			Security:                  r.Security,
			RequiredTransportSecurity: r.RequiredTransportSecurity,
		})
	}
	return p
}

// handlePolicyPost handles POST /api/2.22/nfs-export-policies?names={name}.
func (s *nfsExportPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.NfsExportPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("NFS export policy %q already exists", name))
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	policy := &client.NfsExportPolicy{
		ID:         uuid.New().String(),
		Name:       name,
		Enabled:    enabled,
		IsLocal:    true,
		PolicyType: "nfs",
		Version:    fmt.Sprintf("%d", time.Now().UnixMilli()),
	}

	s.policies[name] = policy
	s.rules[name] = make(map[string]*client.NfsExportPolicyRule)
	s.nextRuleIndex[name] = 1

	WriteJSONListResponse(w, http.StatusOK, []client.NfsExportPolicy{s.policyWithRules(policy)})
}

// handlePolicyPatch handles PATCH /api/2.22/nfs-export-policies?names={name}.
func (s *nfsExportPolicyStore) handlePolicyPatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, ok := s.policies[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("NFS export policy %q not found", name))
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

	WriteJSONListResponse(w, http.StatusOK, []client.NfsExportPolicy{s.policyWithRules(policy)})
}

// handlePolicyDelete handles DELETE /api/2.22/nfs-export-policies?names={name}.
func (s *nfsExportPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("NFS export policy %q not found", name))
		return
	}

	delete(s.policies, name)
	delete(s.rules, name)
	delete(s.nextRuleIndex, name)

	w.WriteHeader(http.StatusOK)
}

// handleRulesGet handles GET /api/2.22/nfs-export-policies/rules.
// Filters by ?policy_names= and optionally ?names=.
func (s *nfsExportPolicyStore) handleRulesGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "policy_names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policyName := r.URL.Query().Get("policy_names")
	ruleName := r.URL.Query().Get("names")

	var items []client.NfsExportPolicyRule

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
		items = []client.NfsExportPolicyRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleRulesPost handles POST /api/2.22/nfs-export-policies/rules?policy_names={name}.
func (s *nfsExportPolicyStore) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"policy_names"}) {
		return
	}

	policyName := r.URL.Query().Get("policy_names")
	if policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names query parameter is required for POST")
		return
	}

	var body client.NfsExportPolicyRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.policies[policyName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("NFS export policy %q not found", policyName))
		return
	}

	// Auto-assign name and index.
	id := uuid.New().String()
	ruleName := "rule-" + id[:8]
	index := s.nextRuleIndex[policyName]
	s.nextRuleIndex[policyName] = index + 1

	atime := true
	if body.Atime != nil {
		atime = *body.Atime
	}
	fileid32bit := false
	if body.Fileid32bit != nil {
		fileid32bit = *body.Fileid32bit
	}
	secure := false
	if body.Secure != nil {
		secure = *body.Secure
	}

	rule := &client.NfsExportPolicyRule{
		ID:                        id,
		Name:                      ruleName,
		Index:                     index,
		Policy:                    client.NamedReference{Name: policyName},
		Access:                    body.Access,
		Client:                    body.Client,
		Permission:                body.Permission,
		Anonuid:                   body.Anonuid,
		Anongid:                   body.Anongid,
		Atime:                     atime,
		Fileid32bit:               fileid32bit,
		Secure:                    secure,
		Security:                  body.Security,
		RequiredTransportSecurity: body.RequiredTransportSecurity,
	}

	if s.rules[policyName] == nil {
		s.rules[policyName] = make(map[string]*client.NfsExportPolicyRule)
	}
	s.rules[policyName][ruleName] = rule

	WriteJSONListResponse(w, http.StatusOK, []client.NfsExportPolicyRule{*rule})
}

// handleRulesPatch handles PATCH /api/2.22/nfs-export-policies/rules?names={name}&policy_names={policy}.
func (s *nfsExportPolicyStore) handleRulesPatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "policy_names"}) {
		return
	}

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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("NFS export policy %q not found", policyName))
		return
	}

	rule, ok := policyRules[ruleName]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("NFS export policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	// Use a raw map for true PATCH semantics.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["index"]; ok {
		var index int
		if err := json.Unmarshal(v, &index); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid index field")
			return
		}
		rule.Index = index
	}
	if v, ok := rawPatch["access"]; ok {
		var access string
		if err := json.Unmarshal(v, &access); err == nil {
			rule.Access = access
		}
	}
	if v, ok := rawPatch["client"]; ok {
		var c string
		if err := json.Unmarshal(v, &c); err == nil {
			rule.Client = c
		}
	}
	if v, ok := rawPatch["permission"]; ok {
		var permission string
		if err := json.Unmarshal(v, &permission); err == nil {
			rule.Permission = permission
		}
	}
	if v, ok := rawPatch["atime"]; ok {
		var atime bool
		if err := json.Unmarshal(v, &atime); err == nil {
			rule.Atime = atime
		}
	}
	if v, ok := rawPatch["fileid_32bit"]; ok {
		var fileid32bit bool
		if err := json.Unmarshal(v, &fileid32bit); err == nil {
			rule.Fileid32bit = fileid32bit
		}
	}
	if v, ok := rawPatch["secure"]; ok {
		var secure bool
		if err := json.Unmarshal(v, &secure); err == nil {
			rule.Secure = secure
		}
	}
	if v, ok := rawPatch["security"]; ok {
		var security []string
		if err := json.Unmarshal(v, &security); err == nil {
			rule.Security = security
		}
	}
	if v, ok := rawPatch["required_transport_security"]; ok {
		var rts string
		if err := json.Unmarshal(v, &rts); err == nil {
			rule.RequiredTransportSecurity = rts
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.NfsExportPolicyRule{*rule})
}

// handleRulesDelete handles DELETE /api/2.22/nfs-export-policies/rules?names={name}&policy_names={policy}.
func (s *nfsExportPolicyStore) handleRulesDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "policy_names"}) {
		return
	}

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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("NFS export policy %q not found", policyName))
		return
	}

	if _, ok := policyRules[ruleName]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("NFS export policy rule %q not found in policy %q", ruleName, policyName))
		return
	}

	delete(policyRules, ruleName)

	w.WriteHeader(http.StatusOK)
}

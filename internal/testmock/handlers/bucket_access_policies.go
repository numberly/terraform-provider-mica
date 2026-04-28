package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// bucketAccessPolicyStore is the thread-safe in-memory state for bucket access policy handlers.
type bucketAccessPolicyStore struct {
	mu       sync.Mutex
	policies map[string]*client.BucketAccessPolicy // keyed by bucket name
	nextID   int
	nextRule int
}

// RegisterBucketAccessPolicyHandlers registers CRUD handlers for
// /api/2.22/buckets/bucket-access-policies and /api/2.22/buckets/bucket-access-policies/rules
// against the provided ServeMux. The returned store pointer can be used for test setup.
func RegisterBucketAccessPolicyHandlers(mux *http.ServeMux) *bucketAccessPolicyStore {
	store := &bucketAccessPolicyStore{
		policies: make(map[string]*client.BucketAccessPolicy),
	}
	mux.HandleFunc("/api/2.22/buckets/bucket-access-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.22/buckets/bucket-access-policies/rules", store.handleRule)
	return store
}

// Seed adds a bucket access policy directly to the store for test setup.
func (s *bucketAccessPolicyStore) Seed(policy *client.BucketAccessPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[policy.Bucket.Name] = policy
}

// handlePolicy dispatches bucket access policy requests by HTTP method.
func (s *bucketAccessPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handlePolicyGet(w, r)
	case http.MethodPost:
		s.handlePolicyPost(w, r)
	case http.MethodDelete:
		s.handlePolicyDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePolicyGet handles GET /api/2.22/buckets/bucket-access-policies.
func (s *bucketAccessPolicyStore) handlePolicyGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	bucketNamesFilter := q.Get("bucket_names")

	var items []client.BucketAccessPolicy

	if bucketNamesFilter != "" {
		if policy, ok := s.policies[bucketNamesFilter]; ok {
			items = append(items, *policy)
		}
	} else {
		for _, policy := range s.policies {
			items = append(items, *policy)
		}
	}

	if items == nil {
		items = []client.BucketAccessPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePolicyPost handles POST /api/2.22/buckets/bucket-access-policies.
func (s *bucketAccessPolicyStore) handlePolicyPost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names"}) {
		return
	}

	bucketName, ok := RequireQueryParam(w, r, "bucket_names")
	if !ok {
		return
	}

	var body client.BucketAccessPolicyPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[bucketName]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("bucket access policy for bucket %q already exists", bucketName))
		return
	}

	s.nextID++
	id := fmt.Sprintf("bap-%d", s.nextID)

	// Build rules from body, auto-generating names and setting defaults.
	var rules []client.BucketAccessPolicyRule
	for _, rulePost := range body.Rules {
		s.nextRule++
		rule := client.BucketAccessPolicyRule{
			Name:       fmt.Sprintf("rule-%d", s.nextRule),
			Actions:    rulePost.Actions,
			Effect:     "allow",
			Principals: rulePost.Principals,
			Resources:  rulePost.Resources,
			Policy:     &client.NamedReference{Name: bucketName},
		}
		rules = append(rules, rule)
	}

	policy := &client.BucketAccessPolicy{
		ID:         id,
		Name:       bucketName,
		Bucket:     client.NamedReference{Name: bucketName},
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "s3",
		Rules:      rules,
	}

	s.policies[bucketName] = policy

	WriteJSONListResponse(w, http.StatusOK, []client.BucketAccessPolicy{*policy})
}

// handlePolicyDelete handles DELETE /api/2.22/buckets/bucket-access-policies.
func (s *bucketAccessPolicyStore) handlePolicyDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names"}) {
		return
	}

	bucketName, ok := RequireQueryParam(w, r, "bucket_names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[bucketName]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("bucket access policy for bucket %q not found", bucketName))
		return
	}

	delete(s.policies, bucketName)

	w.WriteHeader(http.StatusOK)
}

// handleRule dispatches bucket access policy rule requests by HTTP method.
func (s *bucketAccessPolicyStore) handleRule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleRuleGet(w, r)
	case http.MethodPost:
		s.handleRulePost(w, r)
	case http.MethodDelete:
		s.handleRuleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRuleGet handles GET /api/2.22/buckets/bucket-access-policies/rules.
func (s *bucketAccessPolicyStore) handleRuleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	bucketNamesFilter := q.Get("bucket_names")
	namesFilter := q.Get("names")

	var items []client.BucketAccessPolicyRule

	if bucketNamesFilter != "" {
		policy, ok := s.policies[bucketNamesFilter]
		if ok {
			if namesFilter != "" {
				// Filter rules by name.
				for _, rule := range policy.Rules {
					if rule.Name == namesFilter {
						items = append(items, rule)
					}
				}
			} else {
				items = append(items, policy.Rules...)
			}
		}
	} else {
		// Return all rules from all policies.
		for _, policy := range s.policies {
			items = append(items, policy.Rules...)
		}
	}

	if items == nil {
		items = []client.BucketAccessPolicyRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handleRulePost handles POST /api/2.22/buckets/bucket-access-policies/rules.
func (s *bucketAccessPolicyStore) handleRulePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names"}) {
		return
	}

	// Accept either ?names= or ?bucket_names= on POST.
	bucketName := r.URL.Query().Get("names")
	if bucketName == "" {
		bucketName = r.URL.Query().Get("bucket_names")
	}
	if bucketName == "" {
		WriteJSONError(w, http.StatusBadRequest, "names or bucket_names query parameter is required")
		return
	}

	var body client.BucketAccessPolicyRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[bucketName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("bucket access policy for bucket %q not found", bucketName))
		return
	}

	s.nextRule++
	rule := client.BucketAccessPolicyRule{
		Name:       fmt.Sprintf("rule-%d", s.nextRule),
		Actions:    body.Actions,
		Effect:     "allow",
		Principals: body.Principals,
		Resources:  body.Resources,
		Policy:     &client.NamedReference{Name: bucketName},
	}

	policy.Rules = append(policy.Rules, rule)

	WriteJSONListResponse(w, http.StatusOK, []client.BucketAccessPolicyRule{rule})
}

// handleRuleDelete handles DELETE /api/2.22/buckets/bucket-access-policies/rules.
func (s *bucketAccessPolicyStore) handleRuleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names"}) {
		return
	}

	bucketName, ok := RequireQueryParam(w, r, "bucket_names")
	if !ok {
		return
	}

	q := r.URL.Query()
	ruleName := q.Get("names")
	if ruleName == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[bucketName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("bucket access policy for bucket %q not found", bucketName))
		return
	}

	found := false
	for i, rule := range policy.Rules {
		if rule.Name == ruleName {
			policy.Rules = append(policy.Rules[:i], policy.Rules[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("rule %q not found on bucket %q", ruleName, bucketName))
		return
	}

	w.WriteHeader(http.StatusOK)
}

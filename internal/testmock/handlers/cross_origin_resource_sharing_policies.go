package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// corsPolicyStore is the thread-safe in-memory state for CORS policy handlers.
// Rules are stored embedded in each policy's Rules slice, keyed by rule name.
type corsPolicyStore struct {
	mu       sync.Mutex
	policies map[string]*client.CrossOriginResourceSharingPolicy // keyed by bucket name
	nextID   int
}

// RegisterCorsHandlers registers handlers for the CORS policy and its /rules
// sub-endpoint against the provided ServeMux. The returned store pointer can be
// used for test setup via Seed.
func RegisterCorsHandlers(mux *http.ServeMux) *corsPolicyStore {
	store := &corsPolicyStore{
		policies: make(map[string]*client.CrossOriginResourceSharingPolicy),
	}
	mux.HandleFunc("/api/2.23/buckets/cross-origin-resource-sharing-policies", store.handlePolicy)
	mux.HandleFunc("/api/2.23/buckets/cross-origin-resource-sharing-policies/rules", store.handleRules)
	return store
}

// Seed adds a CORS policy directly to the store for test setup.
func (s *corsPolicyStore) Seed(policy *client.CrossOriginResourceSharingPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[policy.Bucket.Name] = policy
}

// ---------- policy endpoint --------------------------------------------------

// handlePolicy dispatches CORS policy requests by HTTP method. There is no PATCH.
func (s *corsPolicyStore) handlePolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET. A missing match returns an empty list with HTTP 200
// (never 404), matching the real array so getOneByName can detect not-found.
func (s *corsPolicyStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bucketNamesFilter := r.URL.Query().Get("bucket_names")

	var items []client.CrossOriginResourceSharingPolicy
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
		items = []client.CrossOriginResourceSharingPolicy{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST with an EMPTY body — creates the CORS policy for the
// bucket (auto-named). If a policy already exists, it returns an "already exists"
// style error (HTTP 400 with "exist" in the message) so EnsureCorsPolicy's
// tolerance path is exercised.
func (s *corsPolicyStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names"}) {
		return
	}

	bucketName, ok := RequireQueryParam(w, r, "bucket_names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[bucketName]; exists {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("CORS policy for bucket %q already exists", bucketName))
		return
	}

	s.nextID++
	policy := &client.CrossOriginResourceSharingPolicy{
		ID:         fmt.Sprintf("cors-%d", s.nextID),
		Name:       fmt.Sprintf("%s-cors", bucketName),
		Bucket:     client.NamedReference{Name: bucketName},
		IsLocal:    true,
		PolicyType: "cross-origin-resource-sharing",
		Rules:      []client.CorsRule{},
	}
	s.policies[bucketName] = policy

	WriteJSONListResponse(w, http.StatusOK, []client.CrossOriginResourceSharingPolicy{*policy})
}

// handleDelete handles DELETE — removes the CORS policy (and its rules) for the bucket.
func (s *corsPolicyStore) handleDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("CORS policy for bucket %q does not exist", bucketName))
		return
	}

	delete(s.policies, bucketName)
	w.WriteHeader(http.StatusOK)
}

// ---------- rules sub-endpoint -----------------------------------------------

// handleRules dispatches CORS rule requests by HTTP method. Only POST and DELETE
// are supported (no PATCH — rules are changed via delete + recreate).
func (s *corsPolicyStore) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleRulePost(w, r)
	case http.MethodDelete:
		s.handleRuleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRulePost adds or replaces a named rule on an existing policy.
func (s *corsPolicyStore) handleRulePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names", "ids"}) {
		return
	}

	bucketName, ok := RequireQueryParam(w, r, "bucket_names")
	if !ok {
		return
	}
	ruleName, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.CorsRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[bucketName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("CORS policy for bucket %q does not exist", bucketName))
		return
	}

	rule := client.CorsRule{
		Name:           ruleName,
		AllowedHeaders: body.AllowedHeaders,
		AllowedMethods: body.AllowedMethods,
		AllowedOrigins: body.AllowedOrigins,
	}

	replaced := false
	for i := range policy.Rules {
		if policy.Rules[i].Name == ruleName {
			policy.Rules[i] = rule
			replaced = true
			break
		}
	}
	if !replaced {
		policy.Rules = append(policy.Rules, rule)
	}

	WriteJSONListResponse(w, http.StatusOK, []client.CorsRule{rule})
}

// handleRuleDelete removes a named rule from a policy.
func (s *corsPolicyStore) handleRuleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names", "ids"}) {
		return
	}

	bucketName, ok := RequireQueryParam(w, r, "bucket_names")
	if !ok {
		return
	}
	ruleName, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[bucketName]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("CORS policy for bucket %q does not exist", bucketName))
		return
	}

	idx := -1
	for i := range policy.Rules {
		if policy.Rules[i].Name == ruleName {
			idx = i
			break
		}
	}
	if idx == -1 {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("CORS rule %q does not exist", ruleName))
		return
	}

	policy.Rules = append(policy.Rules[:idx], policy.Rules[idx+1:]...)
	w.WriteHeader(http.StatusOK)
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// lifecycleRuleStore is the thread-safe in-memory state for lifecycle rule handlers.
type lifecycleRuleStore struct {
	mu     sync.Mutex
	rules  map[string]*client.LifecycleRule // keyed by composite "bucketName/ruleID"
	nextID int
}

// RegisterLifecycleRuleHandlers registers CRUD handlers for /api/2.22/lifecycle-rules
// against the provided ServeMux. The returned store pointer can be used for test setup.
func RegisterLifecycleRuleHandlers(mux *http.ServeMux) *lifecycleRuleStore {
	store := &lifecycleRuleStore{
		rules: make(map[string]*client.LifecycleRule),
	}
	mux.HandleFunc("/api/2.22/lifecycle-rules", store.handle)
	return store
}

// Seed adds a lifecycle rule directly to the store for test setup.
func (s *lifecycleRuleStore) Seed(rule *client.LifecycleRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := rule.Bucket.Name + "/" + rule.RuleID
	s.rules[key] = rule
}

func (s *lifecycleRuleStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/lifecycle-rules with optional query parameters:
// ?bucket_ids=, ?bucket_names=, ?names=, ?ids=.
func (s *lifecycleRuleStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names", "ids"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	namesFilter := q.Get("names")
	bucketNamesFilter := q.Get("bucket_names")
	idsFilter := q.Get("ids")

	var items []client.LifecycleRule

	if namesFilter != "" {
		// Lookup by composite key directly.
		if rule, ok := s.rules[namesFilter]; ok {
			items = append(items, *rule)
		}
	} else if idsFilter != "" {
		// Iterate and match by ID.
		for _, rule := range s.rules {
			if rule.ID == idsFilter {
				items = append(items, *rule)
			}
		}
	} else if bucketNamesFilter != "" {
		// Return all rules where bucket name matches.
		for _, rule := range s.rules {
			if rule.Bucket.Name == bucketNamesFilter {
				items = append(items, *rule)
			}
		}
	} else {
		// Return all rules.
		for _, rule := range s.rules {
			items = append(items, *rule)
		}
	}

	if items == nil {
		items = []client.LifecycleRule{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/lifecycle-rules.
// Query params: confirm_date (optional).
// Body: LifecycleRulePost.
func (s *lifecycleRuleStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"confirm_date"}) {
		return
	}

	var body client.LifecycleRulePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if body.Bucket.Name == "" {
		WriteJSONError(w, http.StatusBadRequest, "bucket name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := body.Bucket.Name + "/" + body.RuleID
	if _, exists := s.rules[key]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("lifecycle rule %q already exists", key))
		return
	}

	s.nextID++
	id := fmt.Sprintf("lcr-%d", s.nextID)

	rule := &client.LifecycleRule{
		ID:                                   id,
		Name:                                 key,
		Bucket:                               client.NamedReference{Name: body.Bucket.Name},
		RuleID:                               body.RuleID,
		Prefix:                               body.Prefix,
		Enabled:                              true,
		AbortIncompleteMultipartUploadsAfter: body.AbortIncompleteMultipartUploadsAfter,
		KeepCurrentVersionFor:                body.KeepCurrentVersionFor,
		KeepCurrentVersionUntil:              body.KeepCurrentVersionUntil,
		KeepPreviousVersionFor:               body.KeepPreviousVersionFor,
	}

	s.rules[key] = rule

	WriteJSONListResponse(w, http.StatusOK, []client.LifecycleRule{*rule})
}

// handlePatch handles PATCH /api/2.22/lifecycle-rules.
// Identification by ?names= (composite key "bucketName/ruleID").
// Uses raw JSON decode for partial update semantics.
func (s *lifecycleRuleStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_names", "names", "confirm_date", "bucket_ids"}) {
		return
	}

	q := r.URL.Query()
	namesFilter := q.Get("names")

	s.mu.Lock()
	defer s.mu.Unlock()

	rule, ok := s.rules[namesFilter]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, "lifecycle rule not found")
		return
	}

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
		rule.Enabled = enabled
	}
	if v, ok := rawPatch["prefix"]; ok {
		var prefix string
		if err := json.Unmarshal(v, &prefix); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid prefix field")
			return
		}
		rule.Prefix = prefix
	}
	if v, ok := rawPatch["abort_incomplete_multipart_uploads_after"]; ok {
		var val int64
		if err := json.Unmarshal(v, &val); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid abort_incomplete_multipart_uploads_after field")
			return
		}
		rule.AbortIncompleteMultipartUploadsAfter = &val
	}
	if v, ok := rawPatch["keep_current_version_for"]; ok {
		var val int64
		if err := json.Unmarshal(v, &val); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid keep_current_version_for field")
			return
		}
		rule.KeepCurrentVersionFor = &val
	}
	if v, ok := rawPatch["keep_current_version_until"]; ok {
		var val int64
		if err := json.Unmarshal(v, &val); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid keep_current_version_until field")
			return
		}
		rule.KeepCurrentVersionUntil = &val
	}
	if v, ok := rawPatch["keep_previous_version_for"]; ok {
		var val int64
		if err := json.Unmarshal(v, &val); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid keep_previous_version_for field")
			return
		}
		rule.KeepPreviousVersionFor = &val
	}

	WriteJSONListResponse(w, http.StatusOK, []client.LifecycleRule{*rule})
}

// handleDelete handles DELETE /api/2.22/lifecycle-rules.
// Identification by ?names= (composite key "bucketName/ruleID").
func (s *lifecycleRuleStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_names", "names", "bucket_ids"}) {
		return
	}

	q := r.URL.Query()
	namesFilter := q.Get("names")

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[namesFilter]; !ok {
		WriteJSONError(w, http.StatusNotFound, "lifecycle rule not found")
		return
	}

	delete(s.rules, namesFilter)

	w.WriteHeader(http.StatusOK)
}

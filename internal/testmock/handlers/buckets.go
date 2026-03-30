package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// BucketStore is the thread-safe in-memory state for bucket handlers.
type BucketStore struct {
	mu       sync.Mutex
	byName   map[string]*client.Bucket
	byID     map[string]*client.Bucket
	accounts *objectStoreAccountStore
}

// RegisterBucketHandlers registers CRUD handlers for /api/2.22/buckets against the provided
// ServeMux. The accounts store is used to validate account references on POST.
// The returned store pointer can be used by other handlers that need bucket cross-reference.
func RegisterBucketHandlers(mux *http.ServeMux, accounts *objectStoreAccountStore) *BucketStore {
	store := &BucketStore{
		byName:   make(map[string]*client.Bucket),
		byID:     make(map[string]*client.Bucket),
		accounts: accounts,
	}
	mux.HandleFunc("/api/2.22/buckets", store.handle)
	return store
}

// handle dispatches bucket requests by HTTP method.
func (s *BucketStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/buckets with optional ?names=, ?ids=, ?destroyed=,
// and ?account_names= query parameters.
func (s *BucketStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "ids", "destroyed", "account_names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	namesFilter := q.Get("names")
	idsFilter := q.Get("ids")
	destroyedFilter := q.Get("destroyed")
	accountNamesFilter := q.Get("account_names")

	var items []client.Bucket

	if namesFilter != "" {
		// Filter by name.
		b, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *b)
		}
	} else if idsFilter != "" {
		// Filter by ID.
		b, ok := s.byID[idsFilter]
		if ok {
			items = append(items, *b)
		}
	} else {
		// Return all buckets (from byID to avoid duplicates).
		for _, b := range s.byID {
			items = append(items, *b)
		}
	}

	// Apply ?account_names= filter.
	if accountNamesFilter != "" {
		accountNames := strings.Split(accountNamesFilter, ",")
		accountSet := make(map[string]bool, len(accountNames))
		for _, an := range accountNames {
			accountSet[strings.TrimSpace(an)] = true
		}
		filtered := items[:0]
		for _, b := range items {
			if accountSet[b.Account.Name] {
				filtered = append(filtered, b)
			}
		}
		items = filtered
	}

	// Apply ?destroyed= filter.
	if destroyedFilter != "" {
		wantDestroyed := destroyedFilter == "true"
		filtered := items[:0]
		for _, b := range items {
			if b.Destroyed == wantDestroyed {
				filtered = append(filtered, b)
			}
		}
		items = filtered
	}

	if items == nil {
		items = []client.Bucket{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/buckets?names={name}.
// Validates the account reference against the account store before creating the bucket.
func (s *BucketStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.BucketPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if body.Account.Name == "" {
		WriteJSONError(w, http.StatusBadRequest, "account.name is required in request body")
		return
	}

	// Validate account exists — cross-reference with account store.
	s.accounts.mu.Lock()
	acct, accountExists := s.accounts.byName[body.Account.Name]
	var accountID string
	if accountExists {
		accountID = acct.ID
	}
	s.accounts.mu.Unlock()

	if !accountExists {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("object store account %q not found", body.Account.Name))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("bucket %q already exists", name))
		return
	}

	b := &client.Bucket{
		ID:   uuid.New().String(),
		Name: name,
		Account: client.NamedReference{
			Name: body.Account.Name,
			ID:   accountID,
		},
		Created:          time.Now().UnixMilli(),
		Destroyed:        false,
		QuotaLimit:       parseQuotaLimit(body.QuotaLimit),
		HardLimitEnabled: body.HardLimitEnabled,
		RetentionLock:    body.RetentionLock,
		Space:            client.Space{},
		EradicationConfig: client.EradicationConfig{
			EradicationDelay:  86400000, // 24h default in ms
			EradicationMode:   "retention-based",
			ManualEradication: "disabled",
		},
		ObjectLockConfig:   client.ObjectLockConfig{},
		PublicAccessConfig:  client.PublicAccessConfig{},
		PublicStatus:        "not-public",
	}

	// Apply config overrides from POST body if provided.
	if body.EradicationConfig != nil {
		b.EradicationConfig = *body.EradicationConfig
	}
	if body.ObjectLockConfig != nil {
		b.ObjectLockConfig = *body.ObjectLockConfig
	}

	s.byName[b.Name] = b
	s.byID[b.ID] = b

	WriteJSONListResponse(w, http.StatusOK, []client.Bucket{*b})
}

// handlePatch handles PATCH /api/2.22/buckets?ids={id}.
// Uses raw map for true PATCH semantics — only provided fields are updated.
func (s *BucketStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids"}) {
		return
	}

	id := r.URL.Query().Get("ids")
	if id == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.byID[id]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("bucket with id %q not found", id))
		return
	}

	// Use a raw map to decode only provided fields (true PATCH semantics).
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["destroyed"]; ok {
		var destroyed bool
		if err := json.Unmarshal(v, &destroyed); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid destroyed field")
			return
		}
		b.Destroyed = destroyed
		if destroyed {
			// Set a non-zero time_remaining to simulate pending eradication.
			b.TimeRemaining = int64(24 * 60 * 60 * 1000) // 24 hours in ms
		} else {
			b.TimeRemaining = 0
		}
	}

	if v, ok := rawPatch["versioning"]; ok {
		var versioning string
		if err := json.Unmarshal(v, &versioning); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid versioning field")
			return
		}
		b.Versioning = versioning
	}

	if v, ok := rawPatch["quota_limit"]; ok {
		var quotaStr string
		if err := json.Unmarshal(v, &quotaStr); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid quota_limit field")
			return
		}
		b.QuotaLimit = parseQuotaLimit(quotaStr)
	}

	if v, ok := rawPatch["hard_limit_enabled"]; ok {
		var hardLimitEnabled bool
		if err := json.Unmarshal(v, &hardLimitEnabled); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid hard_limit_enabled field")
			return
		}
		b.HardLimitEnabled = hardLimitEnabled
	}

	if v, ok := rawPatch["retention_lock"]; ok {
		var retentionLock string
		if err := json.Unmarshal(v, &retentionLock); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid retention_lock field")
			return
		}
		b.RetentionLock = retentionLock
	}

	if v, ok := rawPatch["eradication_config"]; ok {
		var cfg client.EradicationConfig
		if err := json.Unmarshal(v, &cfg); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid eradication_config field")
			return
		}
		b.EradicationConfig = cfg
	}

	if v, ok := rawPatch["object_lock_config"]; ok {
		var cfg client.ObjectLockConfig
		if err := json.Unmarshal(v, &cfg); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid object_lock_config field")
			return
		}
		b.ObjectLockConfig = cfg
	}

	if v, ok := rawPatch["public_access_config"]; ok {
		var cfg client.PublicAccessConfig
		if err := json.Unmarshal(v, &cfg); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid public_access_config field")
			return
		}
		b.PublicAccessConfig = cfg
		if cfg.BlockPublicAccess {
			b.PublicStatus = "not-public"
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.Bucket{*b})
}

// handleDelete handles DELETE /api/2.22/buckets?ids={id}.
// The bucket must already be soft-deleted (destroyed=true) before eradication.
func (s *BucketStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids"}) {
		return
	}

	id := r.URL.Query().Get("ids")
	if id == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.byID[id]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("bucket with id %q not found", id))
		return
	}

	if !b.Destroyed {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("bucket %q must be soft-deleted before eradication", b.Name))
		return
	}

	delete(s.byName, b.Name)
	delete(s.byID, b.ID)

	w.WriteHeader(http.StatusOK)
}

// SetObjectCount sets the ObjectCount field on a bucket identified by name.
// This is a test helper for simulating non-empty buckets.
func (s *BucketStore) SetObjectCount(name string, count int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b, ok := s.byName[name]; ok {
		b.ObjectCount = count
	}
}

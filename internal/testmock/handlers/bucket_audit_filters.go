package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// bucketAuditFilterStore is the thread-safe in-memory state for bucket audit filter handlers.
type bucketAuditFilterStore struct {
	mu      sync.Mutex
	filters map[string]*client.BucketAuditFilter // keyed by bucket name
}

// RegisterBucketAuditFilterHandlers registers CRUD handlers for
// /api/2.22/buckets/audit-filters against the provided ServeMux.
// The returned store pointer can be used for test setup.
func RegisterBucketAuditFilterHandlers(mux *http.ServeMux) *bucketAuditFilterStore {
	store := &bucketAuditFilterStore{
		filters: make(map[string]*client.BucketAuditFilter),
	}
	mux.HandleFunc("/api/2.22/buckets/audit-filters", store.handle)
	return store
}

// Seed adds a bucket audit filter directly to the store for test setup.
func (s *bucketAuditFilterStore) Seed(filter *client.BucketAuditFilter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filters[filter.Bucket.Name] = filter
}

// handle dispatches bucket audit filter requests by HTTP method.
func (s *bucketAuditFilterStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/buckets/audit-filters.
func (s *bucketAuditFilterStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	namesFilter := q.Get("names")
	bucketNamesFilter := q.Get("bucket_names")

	var items []client.BucketAuditFilter

	if namesFilter != "" {
		// Lookup by filter name.
		for _, filter := range s.filters {
			if filter.Name == namesFilter {
				items = append(items, *filter)
			}
		}
	} else if bucketNamesFilter != "" {
		if filter, ok := s.filters[bucketNamesFilter]; ok {
			items = append(items, *filter)
		}
	} else {
		for _, filter := range s.filters {
			items = append(items, *filter)
		}
	}

	if items == nil {
		items = []client.BucketAuditFilter{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/buckets/audit-filters.
func (s *bucketAuditFilterStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names"}) {
		return
	}

	// POST requires both ?names=<filter_name> and ?bucket_names=<bucket_name>.
	filterName := r.URL.Query().Get("names")
	bucketName := r.URL.Query().Get("bucket_names")
	if filterName == "" || bucketName == "" {
		WriteJSONError(w, http.StatusBadRequest, "both names and bucket_names query parameters are required for POST")
		return
	}

	var body client.BucketAuditFilterPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.filters[bucketName]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("bucket audit filter for bucket %q already exists", bucketName))
		return
	}

	filter := &client.BucketAuditFilter{
		Actions:    body.Actions,
		Bucket:     client.NamedReference{Name: bucketName},
		Name:       filterName,
		S3Prefixes: body.S3Prefixes,
	}

	s.filters[bucketName] = filter

	WriteJSONListResponse(w, http.StatusOK, []client.BucketAuditFilter{*filter})
}

// handlePatch handles PATCH /api/2.22/buckets/audit-filters.
func (s *bucketAuditFilterStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names"}) {
		return
	}

	// Accept ?names= (filter name) for lookup.
	filterName := r.URL.Query().Get("names")

	var body client.BucketAuditFilterPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var filter *client.BucketAuditFilter
	for _, f := range s.filters {
		if f.Name == filterName {
			filter = f
			break
		}
	}
	if filter == nil {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("bucket audit filter %q not found", filterName))
		return
	}

	if body.Actions != nil {
		filter.Actions = *body.Actions
	}
	if body.S3Prefixes != nil {
		filter.S3Prefixes = *body.S3Prefixes
	}

	WriteJSONListResponse(w, http.StatusOK, []client.BucketAuditFilter{*filter})
}

// handleDelete handles DELETE /api/2.22/buckets/audit-filters.
func (s *bucketAuditFilterStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"bucket_ids", "bucket_names", "names"}) {
		return
	}

	filterName := r.URL.Query().Get("names")
	if filterName == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find by filter name and delete.
	var bucketKey string
	for k, f := range s.filters {
		if f.Name == filterName {
			bucketKey = k
			break
		}
	}
	if bucketKey == "" {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("bucket audit filter %q not found", filterName))
		return
	}

	delete(s.filters, bucketKey)

	w.WriteHeader(http.StatusOK)
}

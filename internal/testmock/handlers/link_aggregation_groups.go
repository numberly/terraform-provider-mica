// Package handlers provides in-memory mock handlers for FlashBlade API endpoints.
// The LAG mock is read-only (GET only) — LAGs are hardware-managed and cannot be
// created, updated, or deleted via the API. Tests must pre-seed data via Seed().
package handlers

import (
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// lagStore is the thread-safe in-memory state for link aggregation group handlers.
type lagStore struct {
	mu   sync.Mutex
	lags map[string]*client.LinkAggregationGroup
}

// RegisterLinkAggregationGroupHandlers registers a GET-only handler for
// /api/2.22/link-aggregation-groups against the provided ServeMux.
// Non-GET methods return 405 Method Not Allowed.
func RegisterLinkAggregationGroupHandlers(mux *http.ServeMux) *lagStore {
	store := &lagStore{
		lags: make(map[string]*client.LinkAggregationGroup),
	}
	mux.HandleFunc("/api/2.22/link-aggregation-groups", store.handle)
	return store
}

// Seed inserts a LAG directly into the store for test setup.
// Because the LAG endpoint is read-only, there is no POST handler;
// test data must be injected via this method before any GET requests.
func (s *lagStore) Seed(lag *client.LinkAggregationGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lags[lag.Name] = lag
}

// handle dispatches LAG requests by HTTP method.
// Only GET is supported; all other methods return 405.
func (s *lagStore) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleGet(w, r)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// handleGet handles GET /api/2.22/link-aggregation-groups with optional ?names= param.
// If names is provided, returns the matching LAG or an empty list.
// If names is absent, returns all LAGs.
func (s *lagStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.LinkAggregationGroup

	if namesFilter != "" {
		lag, ok := s.lags[namesFilter]
		if ok {
			items = append(items, *lag)
		}
	} else {
		for _, lag := range s.lags {
			items = append(items, *lag)
		}
	}

	if items == nil {
		items = []client.LinkAggregationGroup{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// Package handlers provides in-memory mock handlers for FlashBlade API endpoints.
// The resiliency group mock is read-only (GET only) — resiliency groups are
// hardware-managed and cannot be created, updated, or deleted via the API.
// Tests must pre-seed data via Seed().
package handlers

import (
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// resiliencyGroupStore is the thread-safe in-memory state for resiliency group handlers.
type resiliencyGroupStore struct {
	mu     sync.Mutex
	groups map[string]*client.ResiliencyGroup
}

// RegisterResiliencyGroupHandlers registers a GET-only handler for
// /api/<APIVersion>/resiliency-groups against the provided ServeMux.
// Non-GET methods return 405 Method Not Allowed.
func RegisterResiliencyGroupHandlers(mux *http.ServeMux) *resiliencyGroupStore {
	store := &resiliencyGroupStore{
		groups: make(map[string]*client.ResiliencyGroup),
	}
	mux.HandleFunc(APIPrefix+"/resiliency-groups", store.handle)
	return store
}

// Seed inserts a resiliency group directly into the store for test setup.
// Because the endpoint is read-only, there is no POST handler;
// test data must be injected via this method before any GET requests.
func (s *resiliencyGroupStore) Seed(rg *client.ResiliencyGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups[rg.Name] = rg
}

// Only GET is supported; all other methods return 405.
func (s *resiliencyGroupStore) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleGet(w, r)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// handleGet handles GET /api/<APIVersion>/resiliency-groups with optional ?names= param.
// If names is provided, returns the matching group or an empty list (HTTP 200,
// never 404 — matches real API and lets getOneByName[T] detect not-found).
// If names is absent, returns all groups.
func (s *resiliencyGroupStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.ResiliencyGroup

	if namesFilter != "" {
		rg, ok := s.groups[namesFilter]
		if ok {
			items = append(items, *rg)
		}
	} else {
		for _, rg := range s.groups {
			items = append(items, *rg)
		}
	}

	if items == nil {
		items = []client.ResiliencyGroup{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

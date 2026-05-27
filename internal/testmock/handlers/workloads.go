package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// workloadStore is the thread-safe in-memory state for workload handlers.
type workloadStore struct {
	mu     sync.Mutex
	byName map[string]*client.Workload
	nextID int
}

// RegisterWorkloadHandlers registers CRUD handlers for /api/2.23/workloads
// against the provided ServeMux. The store pointer is returned for test setup.
func RegisterWorkloadHandlers(mux *http.ServeMux) *workloadStore {
	store := &workloadStore{
		byName: make(map[string]*client.Workload),
		nextID: 1,
	}
	mux.HandleFunc(APIPrefix+"/workloads", store.handle)
	return store
}

// Seed adds a workload directly to the store for test setup.
func (s *workloadStore) Seed(w *client.Workload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[w.Name] = w
}

func (s *workloadStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.23/workloads with optional ?names= and ?destroyed= params.
// Returns an empty list (HTTP 200) when no match is found — never 404.
func (s *workloadStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "destroyed"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")
	destroyedFilter := r.URL.Query().Get("destroyed")

	var items []client.Workload

	if namesFilter != "" {
		wl, ok := s.byName[namesFilter]
		if ok {
			// Apply destroyed filter if specified.
			if destroyedFilter == "true" {
				if wl.Destroyed {
					items = append(items, *wl)
				}
			} else {
				if !wl.Destroyed {
					items = append(items, *wl)
				}
			}
		}
	} else {
		for _, wl := range s.byName {
			if destroyedFilter == "true" {
				if wl.Destroyed {
					items = append(items, *wl)
				}
			} else {
				if !wl.Destroyed {
					items = append(items, *wl)
				}
			}
		}
	}

	if items == nil {
		items = []client.Workload{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.23/workloads?names={name}&preset_names={preset}.
// Returns 409 if name already exists.
func (s *workloadStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "preset_names", "preset_ids"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	presetName := r.URL.Query().Get("preset_names")

	var body client.WorkloadPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("workload %q already exists", name))
		return
	}

	id := fmt.Sprintf("wl-%d", s.nextID)
	s.nextID++

	var preset *client.WorkloadPreset
	if presetName != "" {
		preset = &client.WorkloadPreset{
			ID:   fmt.Sprintf("preset-%d", s.nextID),
			Name: presetName,
		}
	}

	// Parameters are consumed on POST; the API does not echo them back in GET.
	_ = body.Parameters

	wl := &client.Workload{
		ID:            id,
		Name:          name,
		Preset:        preset,
		Status:        "ready",
		StatusDetails: []string{},
	}

	s.byName[name] = wl

	WriteJSONListResponse(w, http.StatusOK, []client.Workload{*wl})
}

// handlePatch handles PATCH /api/2.23/workloads?names={name}.
// Applies non-nil pointer fields. Returns 404 if not found.
func (s *workloadStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.WorkloadPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	wl, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("workload %q not found", name))
		return
	}

	if body.Destroyed != nil {
		wl.Destroyed = *body.Destroyed
		if wl.Destroyed {
			wl.Status = "destroying"
			wl.TimeRemaining = 86400000 // 24 hours in ms
		} else {
			wl.Status = "ready"
			wl.TimeRemaining = 0
		}
	}
	if body.Name != nil {
		delete(s.byName, name)
		wl.Name = *body.Name
		s.byName[wl.Name] = wl
	}

	WriteJSONListResponse(w, http.StatusOK, []client.Workload{*wl})
}

// handleDelete handles DELETE /api/2.23/workloads?names={name}.
func (s *workloadStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("workload %q not found", name))
		return
	}

	delete(s.byName, name)

	w.WriteHeader(http.StatusOK)
}

package handlers

import (
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// arrayConnectionStore is the thread-safe in-memory state for array connection handlers.
type arrayConnectionStore struct {
	mu   sync.Mutex
	byID map[string]*client.ArrayConnection
}

// RegisterArrayConnectionHandlers registers GET handler for /api/2.22/array-connections
// against the provided ServeMux. The handler shares in-memory state and is thread-safe.
func RegisterArrayConnectionHandlers(mux *http.ServeMux) *arrayConnectionStore {
	store := &arrayConnectionStore{
		byID: make(map[string]*client.ArrayConnection),
	}
	mux.HandleFunc("/api/2.22/array-connections", store.handle)
	return store
}

// Seed inserts an array connection directly into the store (used by tests to pre-populate data).
func (s *arrayConnectionStore) Seed(conn *client.ArrayConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[conn.ID] = conn
}

// handle dispatches array connection requests by HTTP method.
func (s *arrayConnectionStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET /api/2.22/array-connections with optional ?remote_names= filter.
func (s *arrayConnectionStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	remoteNamesFilter := r.URL.Query().Get("remote_names")

	var items []client.ArrayConnection

	if remoteNamesFilter != "" {
		for _, conn := range s.byID {
			if conn.Remote.Name == remoteNamesFilter {
				items = append(items, *conn)
			}
		}
	} else {
		for _, conn := range s.byID {
			items = append(items, *conn)
		}
	}

	if items == nil {
		items = []client.ArrayConnection{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

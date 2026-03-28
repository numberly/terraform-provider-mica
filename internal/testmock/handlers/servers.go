package handlers

import (
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// serverStore is the thread-safe in-memory state for server handlers.
type serverStore struct {
	mu     sync.Mutex
	byName map[string]*client.Server
	byID   map[string]*client.Server
}

// RegisterServerHandlers registers a GET handler for /api/2.22/servers
// against the provided ServeMux. The handlers share in-memory state and are thread-safe.
func RegisterServerHandlers(mux *http.ServeMux) *serverStore {
	store := &serverStore{
		byName: make(map[string]*client.Server),
		byID:   make(map[string]*client.Server),
	}
	mux.HandleFunc("/api/2.22/servers", store.handle)
	return store
}

// AddServer inserts a server directly into the store (used by tests to seed data).
func (s *serverStore) AddServer(name string) *client.Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := &client.Server{
		ID:   uuid.New().String(),
		Name: name,
	}
	s.byName[srv.Name] = srv
	s.byID[srv.ID] = srv
	return srv
}

// handle dispatches server requests by HTTP method.
func (s *serverStore) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET /api/2.22/servers with optional ?names= param.
func (s *serverStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.Server

	if namesFilter != "" {
		srv, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *srv)
		}
	} else {
		for _, srv := range s.byID {
			items = append(items, *srv)
		}
	}

	if items == nil {
		items = []client.Server{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

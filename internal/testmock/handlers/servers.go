package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// serverStore is the thread-safe in-memory state for server handlers.
type serverStore struct {
	mu     sync.Mutex
	byName map[string]*client.Server
	byID   map[string]*client.Server
}

// RegisterServerHandlers registers CRUD handlers for /api/2.22/servers
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
		ID:      uuid.New().String(),
		Name:    name,
		Created: time.Now().UnixMilli(),
		DNS: []client.ServerDNS{
			{
				Domain:      "test.local",
				Nameservers: []string{"10.0.0.1"},
			},
		},
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

// handlePost handles POST /api/2.22/servers?create_ds={name}.
// The server name comes from the ?create_ds= query parameter.
func (s *serverStore) handlePost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("create_ds")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "create_ds query parameter is required for POST")
		return
	}

	var body client.ServerPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("server %q already exists", name))
		return
	}

	srv := &client.Server{
		ID:      uuid.New().String(),
		Name:    name,
		Created: time.Now().UnixMilli(),
		DNS:     body.DNS,
	}

	s.byName[srv.Name] = srv
	s.byID[srv.ID] = srv

	WriteJSONListResponse(w, http.StatusOK, []client.Server{*srv})
}

// handlePatch handles PATCH /api/2.22/servers?names={name}.
func (s *serverStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	srv, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("server %q not found", name))
		return
	}

	// Use a raw map to decode only provided fields (true PATCH semantics).
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["dns"]; ok {
		var dns []client.ServerDNS
		if err := json.Unmarshal(v, &dns); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid dns field")
			return
		}
		srv.DNS = dns
	}

	WriteJSONListResponse(w, http.StatusOK, []client.Server{*srv})
}

// handleDelete handles DELETE /api/2.22/servers?names={name}.
// Accepts an optional ?cascade_delete= parameter but does not validate it.
func (s *serverStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	srv, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("server %q not found", name))
		return
	}

	delete(s.byName, srv.Name)
	delete(s.byID, srv.ID)

	w.WriteHeader(http.StatusOK)
}

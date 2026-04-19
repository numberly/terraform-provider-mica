package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// syslogServerStore is the thread-safe in-memory state for syslog server handlers.
type syslogServerStore struct {
	mu      sync.Mutex
	servers map[string]*client.SyslogServer
}

// RegisterSyslogServerHandlers registers CRUD handlers for syslog servers.
func RegisterSyslogServerHandlers(mux *http.ServeMux) *syslogServerStore {
	store := &syslogServerStore{
		servers: make(map[string]*client.SyslogServer),
	}
	mux.HandleFunc("/api/2.22/syslog-servers", store.handle)
	return store
}

func (s *syslogServerStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/syslog-servers with optional ?names= param.
func (s *syslogServerStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.SyslogServer

	if namesFilter != "" {
		srv, ok := s.servers[namesFilter]
		if ok {
			items = append(items, *srv)
		}
	} else {
		for _, srv := range s.servers {
			items = append(items, *srv)
		}
	}

	if items == nil {
		items = []client.SyslogServer{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/syslog-servers?names={name}.
func (s *syslogServerStore) handlePost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.SyslogServerPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.servers[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("syslog server %q already exists", name))
		return
	}

	services := body.Services
	if services == nil {
		services = []string{}
	}

	sources := body.Sources
	if sources == nil {
		sources = []string{}
	}

	srv := &client.SyslogServer{
		ID:       uuid.New().String(),
		Name:     name,
		URI:      body.URI,
		Services: services,
		Sources:  sources,
	}

	s.servers[name] = srv

	WriteJSONListResponse(w, http.StatusOK, []client.SyslogServer{*srv})
}

// handlePatch handles PATCH /api/2.22/syslog-servers?names={name}.
func (s *syslogServerStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	srv, ok := s.servers[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("syslog server %q not found", name))
		return
	}

	// Use a raw map for true PATCH semantics.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["uri"]; ok {
		var uri string
		if err := json.Unmarshal(v, &uri); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid uri field")
			return
		}
		srv.URI = uri
	}

	if v, ok := rawPatch["services"]; ok {
		var services []string
		if err := json.Unmarshal(v, &services); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid services field")
			return
		}
		srv.Services = services
	}

	if v, ok := rawPatch["sources"]; ok {
		var sources []string
		if err := json.Unmarshal(v, &sources); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid sources field")
			return
		}
		srv.Sources = sources
	}

	WriteJSONListResponse(w, http.StatusOK, []client.SyslogServer{*srv})
}

// handleDelete handles DELETE /api/2.22/syslog-servers?names={name}.
func (s *syslogServerStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.servers[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("syslog server %q not found", name))
		return
	}

	delete(s.servers, name)

	w.WriteHeader(http.StatusOK)
}

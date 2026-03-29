package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// objectStoreVirtualHostStore is the thread-safe in-memory state for virtual host handlers.
type objectStoreVirtualHostStore struct {
	mu    sync.Mutex
	hosts map[string]*client.ObjectStoreVirtualHost // name -> host
}

// RegisterObjectStoreVirtualHostHandlers registers CRUD handlers for object store virtual hosts.
// Returns the store pointer so resource tests can cross-reference state.
func RegisterObjectStoreVirtualHostHandlers(mux *http.ServeMux) *objectStoreVirtualHostStore {
	store := &objectStoreVirtualHostStore{
		hosts: make(map[string]*client.ObjectStoreVirtualHost),
	}
	mux.HandleFunc("/api/2.22/object-store-virtual-hosts", store.handleVirtualHost)
	return store
}

// Seed adds a virtual host directly to the store for test setup.
func (s *objectStoreVirtualHostStore) Seed(vh *client.ObjectStoreVirtualHost) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hosts[vh.Name] = vh
}

// handleVirtualHost dispatches virtual host requests by HTTP method.
func (s *objectStoreVirtualHostStore) handleVirtualHost(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/object-store-virtual-hosts with optional ?names= param.
func (s *objectStoreVirtualHostStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.ObjectStoreVirtualHost

	if namesFilter != "" {
		host, ok := s.hosts[namesFilter]
		if ok {
			items = append(items, *host)
		}
	} else {
		for _, host := range s.hosts {
			items = append(items, *host)
		}
	}

	if items == nil {
		items = []client.ObjectStoreVirtualHost{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/object-store-virtual-hosts?names={hostname}.
func (s *objectStoreVirtualHostStore) handlePost(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("names")
	if hostname == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.ObjectStoreVirtualHostPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Use hostname as the server-assigned name for mock testing.
	name := hostname
	if _, exists := s.hosts[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("object store virtual host %q already exists", name))
		return
	}

	host := &client.ObjectStoreVirtualHost{
		ID:              uuid.New().String(),
		Name:            name,
		Hostname:        hostname,
		AttachedServers: body.AttachedServers,
	}

	s.hosts[name] = host

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreVirtualHost{*host})
}

// handlePatch handles PATCH /api/2.22/object-store-virtual-hosts?names={name}.
func (s *objectStoreVirtualHostStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	host, ok := s.hosts[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store virtual host %q not found", name))
		return
	}

	// Use a raw map for true PATCH semantics — only provided fields are updated.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["hostname"]; ok {
		var hostname string
		if err := json.Unmarshal(v, &hostname); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid hostname field")
			return
		}
		host.Hostname = hostname
	}

	if v, ok := rawPatch["attached_servers"]; ok {
		var servers []client.NamedReference
		if err := json.Unmarshal(v, &servers); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid attached_servers field")
			return
		}
		host.AttachedServers = servers
	}

	if v, ok := rawPatch["name"]; ok {
		var newName string
		if err := json.Unmarshal(v, &newName); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid name field")
			return
		}
		if newName != name {
			// Rename: update map key.
			delete(s.hosts, name)
			host.Name = newName
			s.hosts[newName] = host
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreVirtualHost{*host})
}

// handleDelete handles DELETE /api/2.22/object-store-virtual-hosts?names={name}.
func (s *objectStoreVirtualHostStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.hosts[name]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store virtual host %q not found", name))
		return
	}

	delete(s.hosts, name)

	w.WriteHeader(http.StatusOK)
}

// Package handlers provides in-memory mock handlers for FlashBlade API endpoints.
// The network-interfaces mock supports full CRUD with the ?names= query parameter at POST,
// consistent with the FlashBlade pattern for user-provided resource names.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// networkInterfaceStore is the thread-safe in-memory state for network interface handlers.
type networkInterfaceStore struct {
	mu     sync.Mutex
	byName map[string]*client.NetworkInterface
	byID   map[string]*client.NetworkInterface
}

// RegisterNetworkInterfaceHandlers registers CRUD handlers for /api/2.22/network-interfaces
// against the provided ServeMux. The handlers share in-memory state and are thread-safe.
func RegisterNetworkInterfaceHandlers(mux *http.ServeMux) *networkInterfaceStore {
	store := &networkInterfaceStore{
		byName: make(map[string]*client.NetworkInterface),
		byID:   make(map[string]*client.NetworkInterface),
	}
	mux.HandleFunc("/api/2.22/network-interfaces", store.handle)
	return store
}

// AddNetworkInterface inserts a network interface directly into the store (used by tests to seed data).
// Defaults: Enabled=true, Gateway="10.21.200.1", MTU=1500, Netmask="255.255.255.0", VLAN=0.
func (s *networkInterfaceStore) AddNetworkInterface(name, address, subnetName, niType, service string) *client.NetworkInterface {
	s.mu.Lock()
	defer s.mu.Unlock()

	ni := &client.NetworkInterface{
		ID:              uuid.New().String(),
		Name:            name,
		Address:         address,
		Enabled:         true,
		Gateway:         "10.21.200.1",
		MTU:             1500,
		Netmask:         "255.255.255.0",
		VLAN:            0,
		Type:            niType,
		Services:        []string{service},
		AttachedServers: nil,
	}
	if subnetName != "" {
		ni.Subnet = &client.NamedReference{Name: subnetName}
	}

	s.byName[ni.Name] = ni
	s.byID[ni.ID] = ni
	return ni
}

func (s *networkInterfaceStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/network-interfaces with optional ?names= param.
// If names is provided, returns the matching interface or an empty list.
// If names is absent, returns all network interfaces.
func (s *networkInterfaceStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.NetworkInterface

	if namesFilter != "" {
		ni, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *ni)
		}
	} else {
		for _, ni := range s.byID {
			items = append(items, *ni)
		}
	}

	if items == nil {
		items = []client.NetworkInterface{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/network-interfaces?names={name}&subnet_names={subnet}.
// The network interface name comes from ?names= and subnet from ?subnet_names= query parameters.
func (s *networkInterfaceStore) handlePost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}
	subnetName := r.URL.Query().Get("subnet_names")

	var body client.NetworkInterfacePost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("network-interface %q already exists", name))
		return
	}

	var subnet *client.NamedReference
	if subnetName != "" {
		subnet = &client.NamedReference{Name: subnetName}
	}

	ni := &client.NetworkInterface{
		ID:              uuid.New().String(),
		Name:            name,
		Address:         body.Address,
		Enabled:         true,
		Gateway:         "10.21.200.1",
		MTU:             1500,
		Netmask:         "255.255.255.0",
		VLAN:            0,
		Type:            body.Type,
		Services:        body.Services,
		Subnet:          subnet,
		AttachedServers: body.AttachedServers,
	}

	s.byName[ni.Name] = ni
	s.byID[ni.ID] = ni

	WriteJSONListResponse(w, http.StatusOK, []client.NetworkInterface{*ni})
}

// handlePatch handles PATCH /api/2.22/network-interfaces?names={name}.
// Uses raw map decoding for true PATCH semantics on address.
// services and attached_servers are always full-replaced when present.
func (s *networkInterfaceStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ni, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("network-interface %q not found", name))
		return
	}

	// Use a raw map to decode only provided fields (true PATCH semantics).
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["address"]; ok {
		var address string
		if err := json.Unmarshal(v, &address); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid address field")
			return
		}
		ni.Address = address
	}

	if v, ok := rawPatch["services"]; ok {
		var services []string
		if err := json.Unmarshal(v, &services); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid services field")
			return
		}
		// Full-replace: overwrite entire slice (including clearing with [])
		ni.Services = services
	}

	if v, ok := rawPatch["attached_servers"]; ok {
		var servers []client.NamedReference
		if err := json.Unmarshal(v, &servers); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid attached_servers field")
			return
		}
		// Full-replace: overwrite entire slice (including clearing with [])
		ni.AttachedServers = servers
	}

	WriteJSONListResponse(w, http.StatusOK, []client.NetworkInterface{*ni})
}

// handleDelete handles DELETE /api/2.22/network-interfaces?names={name}.
func (s *networkInterfaceStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ni, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("network-interface %q not found", name))
		return
	}

	delete(s.byName, ni.Name)
	delete(s.byID, ni.ID)

	w.WriteHeader(http.StatusOK)
}

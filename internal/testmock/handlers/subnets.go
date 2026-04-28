// Package handlers provides in-memory mock handlers for FlashBlade API endpoints.
// The subnet mock supports full CRUD with the ?names= query parameter at POST,
// consistent with the FlashBlade pattern for user-provided resource names.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// subnetStore is the thread-safe in-memory state for subnet handlers.
type subnetStore struct {
	mu     sync.Mutex
	byName map[string]*client.Subnet
	byID   map[string]*client.Subnet
	nextID int
}

// RegisterSubnetHandlers registers CRUD handlers for /api/2.22/subnets
// against the provided ServeMux. The handlers share in-memory state and are thread-safe.
func RegisterSubnetHandlers(mux *http.ServeMux) *subnetStore {
	store := &subnetStore{
		byName: make(map[string]*client.Subnet),
		byID:   make(map[string]*client.Subnet),
	}
	mux.HandleFunc("/api/2.22/subnets", store.handle)
	return store
}

// AddSubnet inserts a subnet directly into the store (used by tests to seed data).
// If lagName is non-empty, the LinkAggregationGroup reference is set accordingly.
// Defaults: Enabled=true, MTU=1500, VLAN=0.
func (s *subnetStore) AddSubnet(name string, prefix string, lagName string) *client.Subnet {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	subnet := &client.Subnet{
		ID:      fmt.Sprintf("subnet-%d", s.nextID),
		Name:    name,
		Enabled: true,
		Prefix:  prefix,
		MTU:     1500,
		VLAN:    0,
	}
	if lagName != "" {
		subnet.LinkAggregationGroup = &client.NamedReference{Name: lagName}
	}

	s.byName[subnet.Name] = subnet
	s.byID[subnet.ID] = subnet
	return subnet
}

func (s *subnetStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/subnets with optional ?names= param.
// If names is provided, returns the matching subnet or an empty list.
// If names is absent, returns all subnets.
func (s *subnetStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.Subnet

	if namesFilter != "" {
		subnet, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *subnet)
		}
	} else {
		for _, subnet := range s.byID {
			items = append(items, *subnet)
		}
	}

	if items == nil {
		items = []client.Subnet{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/subnets?names={name}.
// The subnet name comes from the ?names= query parameter, not the request body.
func (s *subnetStore) handlePost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for POST")
		return
	}

	var body client.SubnetPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("subnet %q already exists", name))
		return
	}

	s.nextID++
	subnet := &client.Subnet{
		ID:                   fmt.Sprintf("subnet-%d", s.nextID),
		Name:                 name,
		Enabled:              true,
		Gateway:              body.Gateway,
		LinkAggregationGroup: body.LinkAggregationGroup,
		MTU:                  body.MTU,
		Prefix:               body.Prefix,
	}
	if body.VLAN != nil {
		subnet.VLAN = *body.VLAN
	}

	s.byName[subnet.Name] = subnet
	s.byID[subnet.ID] = subnet

	WriteJSONListResponse(w, http.StatusOK, []client.Subnet{*subnet})
}

// handlePatch handles PATCH /api/2.22/subnets?names={name}.
// Uses raw map decoding for true PATCH semantics — only provided fields are updated.
func (s *subnetStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	subnet, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("subnet %q not found", name))
		return
	}

	// Use a raw map to decode only provided fields (true PATCH semantics).
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["gateway"]; ok {
		var gateway string
		if err := json.Unmarshal(v, &gateway); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid gateway field")
			return
		}
		subnet.Gateway = gateway
	}

	if v, ok := rawPatch["link_aggregation_group"]; ok {
		var lag *client.NamedReference
		if err := json.Unmarshal(v, &lag); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid link_aggregation_group field")
			return
		}
		subnet.LinkAggregationGroup = lag
	}

	if v, ok := rawPatch["mtu"]; ok {
		var mtu int64
		if err := json.Unmarshal(v, &mtu); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid mtu field")
			return
		}
		subnet.MTU = mtu
	}

	if v, ok := rawPatch["prefix"]; ok {
		var prefix string
		if err := json.Unmarshal(v, &prefix); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid prefix field")
			return
		}
		subnet.Prefix = prefix
	}

	if v, ok := rawPatch["vlan"]; ok {
		var vlan int64
		if err := json.Unmarshal(v, &vlan); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid vlan field")
			return
		}
		subnet.VLAN = vlan
	}

	WriteJSONListResponse(w, http.StatusOK, []client.Subnet{*subnet})
}

// handleDelete handles DELETE /api/2.22/subnets?names={name}.
func (s *subnetStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("names")
	if name == "" {
		WriteJSONError(w, http.StatusBadRequest, "names query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	subnet, ok := s.byName[name]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("subnet %q not found", name))
		return
	}

	delete(s.byName, subnet.Name)
	delete(s.byID, subnet.ID)

	w.WriteHeader(http.StatusOK)
}

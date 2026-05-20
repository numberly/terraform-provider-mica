package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// snmpManagerStore is the thread-safe in-memory state for SNMP manager handlers.
type snmpManagerStore struct {
	mu     sync.Mutex
	byName map[string]*client.SnmpManager
	nextID int
}

// RegisterSnmpManagerHandlers registers CRUD handlers for /api/2.23/snmp-managers
// against the provided ServeMux. The store pointer is returned for test setup.
func RegisterSnmpManagerHandlers(mux *http.ServeMux) *snmpManagerStore {
	store := &snmpManagerStore{
		byName: make(map[string]*client.SnmpManager),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.23/snmp-managers", store.handle)
	return store
}

// Seed inserts a pre-existing SNMP manager directly into the store.
// Use this to set up test fixtures; sensitive fields will be stripped from
// every GET response to mirror real API behaviour.
func (s *snmpManagerStore) Seed(m *client.SnmpManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m.ID == "" {
		m.ID = fmt.Sprintf("snmpmgr-%d", s.nextID)
		s.nextID++
	}
	s.byName[m.Name] = m
}

// Get returns a snapshot of the stored manager (raw, sensitive fields intact)
// for assertion in tests.
func (s *snmpManagerStore) Get(name string) (*client.SnmpManager, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.byName[name]
	if !ok {
		return nil, false
	}
	cp := *m
	return &cp, true
}

// Mutate runs the provided callback under the store mutex with a pointer to
// the named manager. Useful to inject drift in tests.
func (s *snmpManagerStore) Mutate(name string, fn func(m *client.SnmpManager)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.byName[name]
	if !ok {
		return false
	}
	fn(m)
	return true
}

func (s *snmpManagerStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.23/snmp-managers with optional ?names= param.
// No-match returns HTTP 200 + {"items": []} to mirror real API behaviour.
func (s *snmpManagerStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.SnmpManager
	if namesFilter != "" {
		if m, ok := s.byName[namesFilter]; ok {
			items = append(items, *stripSensitive(m))
		}
	} else {
		for _, m := range s.byName {
			items = append(items, *stripSensitive(m))
		}
	}

	if items == nil {
		items = []client.SnmpManager{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.23/snmp-managers?names={name}.
func (s *snmpManagerStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.SnmpManagerPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("snmp manager %q already exists", name))
		return
	}

	id := fmt.Sprintf("snmpmgr-%d", s.nextID)
	s.nextID++

	m := &client.SnmpManager{
		ID:           id,
		Name:         name,
		Host:         body.Host,
		Notification: body.Notification,
		Version:      body.Version,
	}
	if body.V2c != nil {
		m.V2c = &client.SnmpV2c{Community: body.V2c.Community}
	}
	if body.V3 != nil {
		m.V3 = &client.SnmpV3{
			User:              body.V3.User,
			AuthProtocol:      body.V3.AuthProtocol,
			AuthPassphrase:    body.V3.AuthPassphrase,
			PrivacyProtocol:   body.V3.PrivacyProtocol,
			PrivacyPassphrase: body.V3.PrivacyPassphrase,
		}
	}
	s.byName[name] = m

	WriteJSONListResponse(w, http.StatusOK, []client.SnmpManager{*stripSensitive(m)})
}

// handlePatch handles PATCH /api/2.23/snmp-managers?names={name}.
func (s *snmpManagerStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.SnmpManagerPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	m, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("snmp manager %q not found", name))
		return
	}

	if body.Host != nil {
		m.Host = *body.Host
	}
	if body.Notification != nil {
		m.Notification = *body.Notification
	}
	if body.Version != nil {
		m.Version = *body.Version
	}
	if body.V2c != nil {
		v := *body.V2c
		m.V2c = &v
	}
	if body.V3 != nil {
		v := *body.V3
		m.V3 = &v
	}

	WriteJSONListResponse(w, http.StatusOK, []client.SnmpManager{*stripSensitive(m)})
}

// handleDelete handles DELETE /api/2.23/snmp-managers?names={name}.
func (s *snmpManagerStore) handleDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("snmp manager %q not found", name))
		return
	}

	delete(s.byName, name)
	w.WriteHeader(http.StatusOK)
}

// stripSensitive returns a shallow copy with sensitive fields (community,
// auth_passphrase, privacy_passphrase) cleared, mirroring real API GET
// responses which never echo write-once secrets.
func stripSensitive(in *client.SnmpManager) *client.SnmpManager {
	out := *in
	if in.V2c != nil {
		v := *in.V2c
		v.Community = ""
		out.V2c = &v
	}
	if in.V3 != nil {
		v := *in.V3
		v.AuthPassphrase = ""
		v.PrivacyPassphrase = ""
		out.V3 = &v
	}
	return &out
}

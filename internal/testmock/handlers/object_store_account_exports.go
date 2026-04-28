package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/terraform-provider-mica/internal/client"
)

// objectStoreAccountExportStore is the thread-safe in-memory state for object store account export handlers.
type objectStoreAccountExportStore struct {
	mu     sync.Mutex
	byName map[string]*client.ObjectStoreAccountExport
	byID   map[string]*client.ObjectStoreAccountExport
}

// RegisterObjectStoreAccountExportHandlers registers CRUD handlers for /api/2.22/object-store-account-exports
// against the provided ServeMux. The handlers share in-memory state and are thread-safe.
func RegisterObjectStoreAccountExportHandlers(mux *http.ServeMux) *objectStoreAccountExportStore {
	store := &objectStoreAccountExportStore{
		byName: make(map[string]*client.ObjectStoreAccountExport),
		byID:   make(map[string]*client.ObjectStoreAccountExport),
	}
	mux.HandleFunc("/api/2.22/object-store-account-exports", store.handle)
	return store
}

// AddObjectStoreAccountExport seeds an export into the store for test setup.
func (s *objectStoreAccountExportStore) AddObjectStoreAccountExport(accountName, policyName, serverName string) *client.ObjectStoreAccountExport {
	s.mu.Lock()
	defer s.mu.Unlock()

	combinedName := accountName + "/" + accountName
	export := &client.ObjectStoreAccountExport{
		Name:    combinedName,
		ID:      uuid.New().String(),
		Enabled: true,
		Member:  &client.NamedReference{Name: accountName},
		Server:  &client.NamedReference{Name: serverName},
		Policy:  &client.NamedReference{Name: policyName},
	}

	s.byName[combinedName] = export
	s.byID[export.ID] = export
	return export
}

// AddObjectStoreAccountExportWithName seeds an export with a custom combined name into the store for test setup.
func (s *objectStoreAccountExportStore) AddObjectStoreAccountExportWithName(accountName, exportName, policyName, serverName string) *client.ObjectStoreAccountExport {
	s.mu.Lock()
	defer s.mu.Unlock()

	combinedName := accountName + "/" + exportName
	export := &client.ObjectStoreAccountExport{
		Name:    combinedName,
		ID:      uuid.New().String(),
		Enabled: true,
		Member:  &client.NamedReference{Name: accountName},
		Server:  &client.NamedReference{Name: serverName},
		Policy:  &client.NamedReference{Name: policyName},
	}

	s.byName[combinedName] = export
	s.byID[export.ID] = export
	return export
}

func (s *objectStoreAccountExportStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/object-store-account-exports with optional ?names= param.
func (s *objectStoreAccountExportStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.ObjectStoreAccountExport

	if namesFilter != "" {
		export, ok := s.byName[namesFilter]
		if ok {
			items = append(items, *export)
		}
	} else {
		for _, export := range s.byID {
			items = append(items, *export)
		}
	}

	if items == nil {
		items = []client.ObjectStoreAccountExport{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/object-store-account-exports?member_names={accountName}&policy_names={policyName}.
func (s *objectStoreAccountExportStore) handlePost(w http.ResponseWriter, r *http.Request) {
	memberName := r.URL.Query().Get("member_names")
	if memberName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names query parameter is required for POST")
		return
	}
	policyName := r.URL.Query().Get("policy_names")
	if policyName == "" {
		WriteJSONError(w, http.StatusBadRequest, "policy_names query parameter is required for POST")
		return
	}

	var body client.ObjectStoreAccountExportPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	combinedName := memberName + "/" + memberName

	export := &client.ObjectStoreAccountExport{
		Name:    combinedName,
		ID:      uuid.New().String(),
		Enabled: body.ExportEnabled,
		Member:  &client.NamedReference{Name: memberName},
		Policy:  &client.NamedReference{Name: policyName},
	}

	if body.Server != nil {
		export.Server = body.Server
	}

	s.byName[combinedName] = export
	s.byID[export.ID] = export

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccountExport{*export})
}

// handlePatch handles PATCH /api/2.22/object-store-account-exports?ids={id}.
func (s *objectStoreAccountExportStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("ids")
	if id == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	export, ok := s.byID[id]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("object store account export with id %q not found", id))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["export_enabled"]; ok {
		var enabled bool
		if err := json.Unmarshal(v, &enabled); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid export_enabled field")
			return
		}
		export.Enabled = enabled
	}

	if v, ok := rawPatch["policy"]; ok {
		if string(v) == "null" {
			export.Policy = nil
		} else {
			var ref client.NamedReference
			if err := json.Unmarshal(v, &ref); err != nil {
				WriteJSONError(w, http.StatusBadRequest, "invalid policy field")
				return
			}
			export.Policy = &ref
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreAccountExport{*export})
}

// handleDelete handles DELETE /api/2.22/object-store-account-exports?member_names={accountName}&names={exportName}.
// The real FlashBlade API expects names= to contain the short export name (not the combined "account/export" format).
// This mock enforces strict lookup: memberName + "/" + exportName must match an existing combined key.
func (s *objectStoreAccountExportStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	memberName := r.URL.Query().Get("member_names")
	exportName := r.URL.Query().Get("names")

	if memberName == "" || exportName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names and names query parameters are required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Strict lookup: memberName + "/" + exportName must match the stored combined name.
	combinedKey := memberName + "/" + exportName
	export, ok := s.byName[combinedKey]

	if !ok {
		// FlashBlade returns 200 with empty items for not-found on delete.
		w.WriteHeader(http.StatusOK)
		return
	}

	delete(s.byName, export.Name)
	delete(s.byID, export.ID)

	w.WriteHeader(http.StatusOK)
}

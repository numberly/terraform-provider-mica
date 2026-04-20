package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// fileSystemExportStore is the thread-safe in-memory state for file system export handlers.
type fileSystemExportStore struct {
	mu     sync.Mutex
	byName map[string]*client.FileSystemExport
	byID   map[string]*client.FileSystemExport
}

// RegisterFileSystemExportHandlers registers CRUD handlers for /api/2.22/file-system-exports
// against the provided ServeMux. The handlers share in-memory state and are thread-safe.
func RegisterFileSystemExportHandlers(mux *http.ServeMux) *fileSystemExportStore {
	store := &fileSystemExportStore{
		byName: make(map[string]*client.FileSystemExport),
		byID:   make(map[string]*client.FileSystemExport),
	}
	mux.HandleFunc("/api/2.22/file-system-exports", store.handle)
	return store
}

// AddFileSystemExport seeds an export into the store for test setup.
func (s *fileSystemExportStore) AddFileSystemExport(fsName, policyName, serverName string) *client.FileSystemExport {
	s.mu.Lock()
	defer s.mu.Unlock()

	combinedName := fsName + "/" + fsName
	export := &client.FileSystemExport{
		Name:       combinedName,
		ExportName: fsName,
		ID:         uuid.New().String(),
		Enabled:    true,
		Member:     &client.NamedReference{Name: fsName},
		Server:     &client.NamedReference{Name: serverName},
		Policy:     &client.NamedReference{Name: policyName},
		PolicyType: "nfs",
		Status:     "exported",
	}

	s.byName[combinedName] = export
	s.byID[export.ID] = export
	return export
}

func (s *fileSystemExportStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/file-system-exports with optional ?names= param.
func (s *fileSystemExportStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.FileSystemExport

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
		items = []client.FileSystemExport{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/file-system-exports?member_names={fsName}&policy_names={policyName}.
func (s *fileSystemExportStore) handlePost(w http.ResponseWriter, r *http.Request) {
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

	var body client.FileSystemExportPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	exportName := body.ExportName
	if exportName == "" {
		exportName = memberName
	}
	combinedName := memberName + "/" + exportName

	export := &client.FileSystemExport{
		Name:       combinedName,
		ExportName: exportName,
		ID:         uuid.New().String(),
		Enabled:    true,
		Member:     &client.NamedReference{Name: memberName},
		Policy:     &client.NamedReference{Name: policyName},
		PolicyType: "nfs",
		Status:     "exported",
	}

	if body.Server != nil {
		export.Server = body.Server
	}
	if body.SharePolicy != nil {
		export.SharePolicy = body.SharePolicy
	}

	s.byName[combinedName] = export
	s.byID[export.ID] = export

	WriteJSONListResponse(w, http.StatusOK, []client.FileSystemExport{*export})
}

// handlePatch handles PATCH /api/2.22/file-system-exports?ids={id}.
func (s *fileSystemExportStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("ids")
	if id == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	export, ok := s.byID[id]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("file system export with id %q not found", id))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["export_name"]; ok {
		var exportName string
		if err := json.Unmarshal(v, &exportName); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid export_name field")
			return
		}
		// Update combined name.
		oldName := export.Name
		delete(s.byName, oldName)
		export.ExportName = exportName
		export.Name = export.Member.Name + "/" + exportName
		s.byName[export.Name] = export
	}

	if v, ok := rawPatch["server"]; ok {
		if string(v) == "null" {
			export.Server = nil
		} else {
			var ref client.NamedReference
			if err := json.Unmarshal(v, &ref); err != nil {
				WriteJSONError(w, http.StatusBadRequest, "invalid server field")
				return
			}
			export.Server = &ref
		}
	}

	if v, ok := rawPatch["share_policy"]; ok {
		if string(v) == "null" {
			export.SharePolicy = nil
		} else {
			var ref client.NamedReference
			if err := json.Unmarshal(v, &ref); err != nil {
				WriteJSONError(w, http.StatusBadRequest, "invalid share_policy field")
				return
			}
			export.SharePolicy = &ref
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.FileSystemExport{*export})
}

// handleDelete handles DELETE /api/2.22/file-system-exports?member_names={fsName}&names={exportName}.
func (s *fileSystemExportStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	memberName := r.URL.Query().Get("member_names")
	exportName := r.URL.Query().Get("names")

	if memberName == "" || exportName == "" {
		WriteJSONError(w, http.StatusBadRequest, "member_names and names query parameters are required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	combinedName := memberName + "/" + exportName
	export, ok := s.byName[combinedName]
	if !ok {
		// FlashBlade returns 200 with empty items for not-found on delete.
		w.WriteHeader(http.StatusOK)
		return
	}

	delete(s.byName, combinedName)
	delete(s.byID, export.ID)

	w.WriteHeader(http.StatusOK)
}

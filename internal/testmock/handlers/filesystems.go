// Package handlers provides mock HTTP handlers for FlashBlade API resources.
// Each handler maintains its own in-memory state and can be registered against
// a MockServer's Mux for testing.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// fileSystemStore is the thread-safe in-memory state for file system handlers.
type fileSystemStore struct {
	mu      sync.Mutex
	byName  map[string]*client.FileSystem
	byID    map[string]*client.FileSystem
}

// RegisterFileSystemHandlers registers CRUD handlers for /api/2.22/file-systems
// against the provided ServeMux. The handlers share in-memory state and are
// thread-safe.
func RegisterFileSystemHandlers(mux *http.ServeMux) {
	store := &fileSystemStore{
		byName: make(map[string]*client.FileSystem),
		byID:   make(map[string]*client.FileSystem),
	}
	mux.HandleFunc("/api/2.22/file-systems", store.handle)
}

// handle dispatches file system requests by HTTP method.
func (s *fileSystemStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/file-systems with optional ?names=, ?ids=, ?destroyed= params.
func (s *fileSystemStore) handleGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	namesFilter := q.Get("names")
	idsFilter := q.Get("ids")
	destroyedFilter := q.Get("destroyed")

	var items []client.FileSystem

	if namesFilter != "" {
		fs, ok := s.byName[namesFilter]
		if ok {
			if destroyedFilter == "" || (destroyedFilter == "true" && fs.Destroyed) || (destroyedFilter == "false" && !fs.Destroyed) {
				items = append(items, *fs)
			}
		}
	} else if idsFilter != "" {
		fs, ok := s.byID[idsFilter]
		if ok {
			if destroyedFilter == "" || (destroyedFilter == "true" && fs.Destroyed) || (destroyedFilter == "false" && !fs.Destroyed) {
				items = append(items, *fs)
			}
		}
	} else {
		// Return all file systems, optionally filtered by destroyed status.
		for _, fs := range s.byID {
			if destroyedFilter == "" || (destroyedFilter == "true" && fs.Destroyed) || (destroyedFilter == "false" && !fs.Destroyed) {
				items = append(items, *fs)
			}
		}
	}

	if items == nil {
		items = []client.FileSystem{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/file-systems.
func (s *fileSystemStore) handlePost(w http.ResponseWriter, r *http.Request) {
	var body client.FileSystemPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}
	if body.Name == "" {
		WriteJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[body.Name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("file system %q already exists", body.Name))
		return
	}

	fs := &client.FileSystem{
		ID:          uuid.New().String(),
		Name:        body.Name,
		Provisioned: body.Provisioned,
		Destroyed:   false,
		Created:     time.Now().UnixMilli(),
		Space:       client.Space{},
		NFS:         body.NFS,
		SMB:         body.SMB,
		Writable:    body.Writable,
	}

	s.byName[fs.Name] = fs
	s.byID[fs.ID] = fs

	WriteJSONListResponse(w, http.StatusOK, []client.FileSystem{*fs})
}

// handlePatch handles PATCH /api/2.22/file-systems?ids={id}.
func (s *fileSystemStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("ids")
	if id == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fs, ok := s.byID[id]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("file system with id %q not found", id))
		return
	}

	// Decode only the fields present in the request body (PATCH semantics).
	// We use a raw map to detect which fields were provided.
	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	oldName := fs.Name

	if v, ok := rawPatch["name"]; ok {
		var name string
		if err := json.Unmarshal(v, &name); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid name field")
			return
		}
		if name != fs.Name {
			// Perform rename: update byName index.
			delete(s.byName, oldName)
			fs.Name = name
			s.byName[name] = fs
		}
	}

	if v, ok := rawPatch["provisioned"]; ok {
		var provisioned int64
		if err := json.Unmarshal(v, &provisioned); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid provisioned field")
			return
		}
		fs.Provisioned = provisioned
	}

	if v, ok := rawPatch["destroyed"]; ok {
		var destroyed bool
		if err := json.Unmarshal(v, &destroyed); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid destroyed field")
			return
		}
		fs.Destroyed = destroyed
	}

	if v, ok := rawPatch["writable"]; ok {
		var writable bool
		if err := json.Unmarshal(v, &writable); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid writable field")
			return
		}
		fs.Writable = writable
	}

	if v, ok := rawPatch["nfs"]; ok {
		var nfs client.NFSConfig
		if err := json.Unmarshal(v, &nfs); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid nfs field")
			return
		}
		fs.NFS = nfs
	}

	if v, ok := rawPatch["smb"]; ok {
		var smb client.SMBConfig
		if err := json.Unmarshal(v, &smb); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid smb field")
			return
		}
		fs.SMB = smb
	}

	WriteJSONListResponse(w, http.StatusOK, []client.FileSystem{*fs})
}

// handleDelete handles DELETE /api/2.22/file-systems?ids={id}.
// Only works on file systems that are already soft-deleted (destroyed=true).
func (s *fileSystemStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("ids")
	if id == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fs, ok := s.byID[id]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("file system with id %q not found", id))
		return
	}

	if !fs.Destroyed {
		WriteJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("file system %q must be destroyed before eradication (set destroyed=true first)", fs.Name))
		return
	}

	// Remove from both indexes — simulates eradication.
	delete(s.byName, fs.Name)
	delete(s.byID, id)

	w.WriteHeader(http.StatusOK)
}


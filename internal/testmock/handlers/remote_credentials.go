package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// remoteCredentialsStore is the thread-safe in-memory state for object store remote credentials handlers.
type remoteCredentialsStore struct {
	mu     sync.Mutex
	byName map[string]*client.ObjectStoreRemoteCredentials
	nextID int
}

// RegisterRemoteCredentialsHandlers registers CRUD handlers for /api/2.22/object-store-remote-credentials
// against the provided ServeMux. The store pointer is returned for cross-reference if needed.
func RegisterRemoteCredentialsHandlers(mux *http.ServeMux) *remoteCredentialsStore {
	store := &remoteCredentialsStore{
		byName: make(map[string]*client.ObjectStoreRemoteCredentials),
		nextID: 1,
	}
	mux.HandleFunc("/api/2.22/object-store-remote-credentials", store.handle)
	return store
}

// Seed adds remote credentials directly to the store for test setup.
func (s *remoteCredentialsStore) Seed(cred *client.ObjectStoreRemoteCredentials) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[cred.Name] = cred
}

func (s *remoteCredentialsStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/object-store-remote-credentials with optional ?names= param.
// IMPORTANT: secret_access_key is NOT returned in GET responses — it is set to empty string.
func (s *remoteCredentialsStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	namesFilter := r.URL.Query().Get("names")

	var items []client.ObjectStoreRemoteCredentials

	if namesFilter != "" {
		cred, ok := s.byName[namesFilter]
		if ok {
			redacted := *cred
			redacted.SecretAccessKey = ""
			items = append(items, redacted)
		}
	} else {
		for _, cred := range s.byName {
			redacted := *cred
			redacted.SecretAccessKey = ""
			items = append(items, redacted)
		}
	}

	if items == nil {
		items = []client.ObjectStoreRemoteCredentials{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/object-store-remote-credentials.
// Requires ?names= and exactly one of ?remote_names= or ?target_names=.
// Body contains access_key_id + secret_access_key.
// Response includes secret_access_key — POST only.
func (s *remoteCredentialsStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names", "remote_names", "target_names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	q := r.URL.Query()
	remoteName := q.Get("remote_names")
	targetName := q.Get("target_names")

	if remoteName != "" && targetName != "" {
		WriteJSONError(w, http.StatusBadRequest, "provide remote_names or target_names, not both")
		return
	}
	if remoteName == "" && targetName == "" {
		WriteJSONError(w, http.StatusBadRequest, "remote_names or target_names is required")
		return
	}

	// Use whichever ref param was provided as the Remote.Name on the stored credential.
	refName := remoteName
	if targetName != "" {
		refName = targetName
	}

	var body client.ObjectStoreRemoteCredentialsPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if body.AccessKeyID == "" {
		WriteJSONError(w, http.StatusBadRequest, "access_key_id is required")
		return
	}
	if body.SecretAccessKey == "" {
		WriteJSONError(w, http.StatusBadRequest, "secret_access_key is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[name]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("remote credentials %q already exist", name))
		return
	}

	id := fmt.Sprintf("rc-%d", s.nextID)
	s.nextID++

	cred := &client.ObjectStoreRemoteCredentials{
		ID:              id,
		Name:            name,
		AccessKeyID:     body.AccessKeyID,
		SecretAccessKey: body.SecretAccessKey,
		Remote:          client.NamedReference{Name: refName},
	}

	s.byName[name] = cred

	// POST response includes secret_access_key.
	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreRemoteCredentials{*cred})
}

// handlePatch handles PATCH /api/2.22/object-store-remote-credentials?names={name}.
// Updates access_key_id and/or secret_access_key. Response does NOT include secret_access_key (like GET).
func (s *remoteCredentialsStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"names"}) {
		return
	}

	name, ok := RequireQueryParam(w, r, "names")
	if !ok {
		return
	}

	var body client.ObjectStoreRemoteCredentialsPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cred, exists := s.byName[name]
	if !exists {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("remote credentials %q not found", name))
		return
	}

	if body.AccessKeyID != nil {
		cred.AccessKeyID = *body.AccessKeyID
	}
	if body.SecretAccessKey != nil {
		cred.SecretAccessKey = *body.SecretAccessKey
	}

	// PATCH response strips secret_access_key (like GET).
	redacted := *cred
	redacted.SecretAccessKey = ""
	WriteJSONListResponse(w, http.StatusOK, []client.ObjectStoreRemoteCredentials{redacted})
}

// handleDelete handles DELETE /api/2.22/object-store-remote-credentials?names={name}.
func (s *remoteCredentialsStore) handleDelete(w http.ResponseWriter, r *http.Request) {
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("remote credentials %q not found", name))
		return
	}

	delete(s.byName, name)

	w.WriteHeader(http.StatusOK)
}

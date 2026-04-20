package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// bucketReplicaLinkStore is the thread-safe in-memory state for bucket replica link handlers.
// Primary key is the link ID — multiple links can exist for the same bucket pair
// (e.g., one via array connection and one via S3 target).
type bucketReplicaLinkStore struct {
	mu     sync.Mutex
	byID   map[string]*client.BucketReplicaLink
	nextID int
}

// RegisterBucketReplicaLinkHandlers registers CRUD handlers for /api/2.22/bucket-replica-links
// against the provided ServeMux. The returned store pointer can be used for cross-reference.
func RegisterBucketReplicaLinkHandlers(mux *http.ServeMux) *bucketReplicaLinkStore {
	store := &bucketReplicaLinkStore{
		byID: make(map[string]*client.BucketReplicaLink),
	}
	mux.HandleFunc("/api/2.22/bucket-replica-links", store.handle)
	return store
}

// Seed adds a bucket replica link directly to the store for test setup.
func (s *bucketReplicaLinkStore) Seed(link *client.BucketReplicaLink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[link.ID] = link
}

func (s *bucketReplicaLinkStore) handle(w http.ResponseWriter, r *http.Request) {
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

// handleGet handles GET /api/2.22/bucket-replica-links with optional query parameters:
// ?local_bucket_names=, ?remote_bucket_names=, ?ids=.
func (s *bucketReplicaLinkStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"local_bucket_names", "remote_bucket_names", "ids"}) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q := r.URL.Query()
	localFilter := q.Get("local_bucket_names")
	remoteFilter := q.Get("remote_bucket_names")
	idsFilter := q.Get("ids")

	var items []client.BucketReplicaLink

	if idsFilter != "" {
		if link, ok := s.byID[idsFilter]; ok {
			items = append(items, *link)
		}
	} else {
		for _, link := range s.byID {
			if localFilter != "" && link.LocalBucket.Name != localFilter {
				continue
			}
			if remoteFilter != "" && link.RemoteBucket.Name != remoteFilter {
				continue
			}
			items = append(items, *link)
		}
	}

	if items == nil {
		items = []client.BucketReplicaLink{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

// handlePost handles POST /api/2.22/bucket-replica-links.
// Query params: local_bucket_names, remote_bucket_names, remote_credentials_names (optional).
// Body: paused, cascading_enabled.
func (s *bucketReplicaLinkStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"local_bucket_names", "remote_bucket_names", "remote_credentials_names"}) {
		return
	}

	q := r.URL.Query()
	localBucket := q.Get("local_bucket_names")
	remoteBucket := q.Get("remote_bucket_names")
	remoteCredentials := q.Get("remote_credentials_names")

	if localBucket == "" || remoteBucket == "" {
		WriteJSONError(w, http.StatusBadRequest, "local_bucket_names and remote_bucket_names query parameters are required for POST")
		return
	}

	var body client.BucketReplicaLinkPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate: same bucket pair AND same remote_credentials.
	for _, existing := range s.byID {
		if existing.LocalBucket.Name == localBucket && existing.RemoteBucket.Name == remoteBucket {
			existingRC := ""
			if existing.RemoteCredentials != nil {
				existingRC = existing.RemoteCredentials.Name
			}
			if existingRC == remoteCredentials {
				WriteJSONError(w, http.StatusConflict, fmt.Sprintf("bucket replica link %s/%s with credentials %q already exists", localBucket, remoteBucket, remoteCredentials))
				return
			}
		}
	}

	s.nextID++
	id := fmt.Sprintf("brl-%d", s.nextID)

	link := &client.BucketReplicaLink{
		ID:               id,
		LocalBucket:      client.NamedReference{Name: localBucket},
		RemoteBucket:     client.NamedReference{Name: remoteBucket},
		Remote:           client.NamedReference{Name: "mock-remote"},
		Paused:           body.Paused,
		CascadingEnabled: body.CascadingEnabled,
		Direction:        "outbound",
		Status:           "replicating",
	}

	if remoteCredentials != "" {
		link.RemoteCredentials = &client.NamedReference{Name: remoteCredentials}
	}

	s.byID[id] = link

	WriteJSONListResponse(w, http.StatusOK, []client.BucketReplicaLink{*link})
}

// handlePatch handles PATCH /api/2.22/bucket-replica-links.
// Identification by ?ids= only (unambiguous).
func (s *bucketReplicaLinkStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids"}) {
		return
	}

	idsFilter := r.URL.Query().Get("ids")
	if idsFilter == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	link := s.byID[idsFilter]
	if link == nil {
		WriteJSONError(w, http.StatusNotFound, "bucket replica link not found")
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["paused"]; ok {
		var paused bool
		if err := json.Unmarshal(v, &paused); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid paused field")
			return
		}
		link.Paused = paused
	}

	WriteJSONListResponse(w, http.StatusOK, []client.BucketReplicaLink{*link})
}

// handleDelete handles DELETE /api/2.22/bucket-replica-links.
// Identification by ?ids= only (unambiguous).
func (s *bucketReplicaLinkStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids"}) {
		return
	}

	idsFilter := r.URL.Query().Get("ids")
	if idsFilter == "" {
		WriteJSONError(w, http.StatusBadRequest, "ids query parameter is required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byID[idsFilter]; !ok {
		WriteJSONError(w, http.StatusNotFound, "bucket replica link not found")
		return
	}

	delete(s.byID, idsFilter)

	w.WriteHeader(http.StatusOK)
}

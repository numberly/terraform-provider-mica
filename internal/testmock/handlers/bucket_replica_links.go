package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// bucketReplicaLinkStore is the thread-safe in-memory state for bucket replica link handlers.
type bucketReplicaLinkStore struct {
	mu     sync.Mutex
	links  map[string]*client.BucketReplicaLink // keyed by "localBucket/remoteBucket"
	byID   map[string]*client.BucketReplicaLink
	nextID int
}

// RegisterBucketReplicaLinkHandlers registers CRUD handlers for /api/2.22/bucket-replica-links
// against the provided ServeMux. The returned store pointer can be used for cross-reference.
func RegisterBucketReplicaLinkHandlers(mux *http.ServeMux) *bucketReplicaLinkStore {
	store := &bucketReplicaLinkStore{
		links: make(map[string]*client.BucketReplicaLink),
		byID:  make(map[string]*client.BucketReplicaLink),
	}
	mux.HandleFunc("/api/2.22/bucket-replica-links", store.handle)
	return store
}

// handle dispatches bucket replica link requests by HTTP method.
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

// compositeKey returns the store key for a bucket replica link.
func compositeKey(localBucket, remoteBucket string) string {
	return localBucket + "/" + remoteBucket
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
		// Lookup by ID.
		if link, ok := s.byID[idsFilter]; ok {
			items = append(items, *link)
		}
	} else if localFilter != "" && remoteFilter != "" {
		// Lookup by composite key.
		key := compositeKey(localFilter, remoteFilter)
		if link, ok := s.links[key]; ok {
			items = append(items, *link)
		}
	} else if localFilter != "" {
		// Return all links for a given local bucket.
		for _, link := range s.links {
			if link.LocalBucket.Name == localFilter {
				items = append(items, *link)
			}
		}
	} else {
		// Return all links.
		for _, link := range s.links {
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

	key := compositeKey(localBucket, remoteBucket)
	if _, exists := s.links[key]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("bucket replica link %q already exists", key))
		return
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

	s.links[key] = link
	s.byID[id] = link

	WriteJSONListResponse(w, http.StatusOK, []client.BucketReplicaLink{*link})
}

// handlePatch handles PATCH /api/2.22/bucket-replica-links.
// Identification by ?ids= (primary) or ?local_bucket_names= + ?remote_bucket_names=.
// Only paused is mutable.
func (s *bucketReplicaLinkStore) handlePatch(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "local_bucket_names", "remote_bucket_names"}) {
		return
	}

	q := r.URL.Query()
	idsFilter := q.Get("ids")
	localFilter := q.Get("local_bucket_names")
	remoteFilter := q.Get("remote_bucket_names")

	s.mu.Lock()
	defer s.mu.Unlock()

	var link *client.BucketReplicaLink

	if idsFilter != "" {
		link = s.byID[idsFilter]
	} else if localFilter != "" && remoteFilter != "" {
		link = s.links[compositeKey(localFilter, remoteFilter)]
	}

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
// Identification by ?local_bucket_names= + ?remote_bucket_names= (or ?ids=).
func (s *bucketReplicaLinkStore) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"ids", "local_bucket_names", "remote_bucket_names"}) {
		return
	}

	q := r.URL.Query()
	idsFilter := q.Get("ids")
	localFilter := q.Get("local_bucket_names")
	remoteFilter := q.Get("remote_bucket_names")

	s.mu.Lock()
	defer s.mu.Unlock()

	var link *client.BucketReplicaLink
	var key string

	if idsFilter != "" {
		link = s.byID[idsFilter]
		if link != nil {
			key = compositeKey(link.LocalBucket.Name, link.RemoteBucket.Name)
		}
	} else if localFilter != "" && remoteFilter != "" {
		key = compositeKey(localFilter, remoteFilter)
		link = s.links[key]
	}

	if link == nil {
		WriteJSONError(w, http.StatusNotFound, "bucket replica link not found")
		return
	}

	delete(s.links, key)
	delete(s.byID, link.ID)

	w.WriteHeader(http.StatusOK)
}

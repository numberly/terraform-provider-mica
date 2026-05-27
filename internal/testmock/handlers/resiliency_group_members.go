// Package handlers — resiliency group members mock.
//
// The endpoint GET /api/2.23/resiliency-groups/members is read-only and
// requires filtering by parent (`resiliency_group_names` query param). The
// mock store therefore keys rows by (groupName, memberName).
package handlers

import (
	"net/http"
	"sync"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// memberKey identifies a member row uniquely.
type memberKey struct {
	group  string
	member string
}

// resiliencyGroupMemberStore is the thread-safe in-memory state for
// resiliency-group member handlers.
type resiliencyGroupMemberStore struct {
	mu      sync.Mutex
	members map[memberKey]*client.ResiliencyGroupMember
}

// RegisterResiliencyGroupMemberHandlers registers a GET-only handler for
// /api/2.23/resiliency-groups/members against the provided ServeMux.
// Non-GET methods return 405 Method Not Allowed.
func RegisterResiliencyGroupMemberHandlers(mux *http.ServeMux) *resiliencyGroupMemberStore {
	store := &resiliencyGroupMemberStore{
		members: make(map[memberKey]*client.ResiliencyGroupMember),
	}
	mux.HandleFunc(APIPrefix+"/resiliency-groups/members", store.handle)
	return store
}

// Seed inserts a member row directly into the store for test setup.
// Members are hardware-managed; no POST handler exists, so tests must seed
// state before any GET request.
func (s *resiliencyGroupMemberStore) Seed(m *client.ResiliencyGroupMember) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[memberKey{group: m.Group.Name, member: m.Member.Name}] = m
}

func (s *resiliencyGroupMemberStore) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleGet(w, r)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// handleGet returns members filtered by `resiliency_group_names`. The
// parameter is required: callers must scope the listing to a parent group.
// Empty match returns HTTP 200 with an empty items list (never 404), matching
// the real API contract.
func (s *resiliencyGroupMemberStore) handleGet(w http.ResponseWriter, r *http.Request) {
	if !ValidateQueryParams(w, r, []string{"resiliency_group_ids", "resiliency_group_names", "ids", "names"}) {
		return
	}

	groupName, ok := RequireQueryParam(w, r, "resiliency_group_names")
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items := []client.ResiliencyGroupMember{}
	for k, m := range s.members {
		if k.group == groupName {
			items = append(items, *m)
		}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// quotaStore is the thread-safe in-memory state for quota handlers.
type quotaStore struct {
	mu          sync.Mutex
	userQuotas  map[string]*client.QuotaUser  // "fileSystemName/uid" -> quota
	groupQuotas map[string]*client.QuotaGroup // "fileSystemName/gid" -> quota
}

func quotaUserKey(fileSystemName, uid string) string {
	return fileSystemName + "/" + uid
}

func quotaGroupKey(fileSystemName, gid string) string {
	return fileSystemName + "/" + gid
}

// RegisterQuotaHandlers registers CRUD handlers for user and group quotas.
func RegisterQuotaHandlers(mux *http.ServeMux) *quotaStore {
	store := &quotaStore{
		userQuotas:  make(map[string]*client.QuotaUser),
		groupQuotas: make(map[string]*client.QuotaGroup),
	}
	mux.HandleFunc("/api/2.22/quotas/users", store.handleUsers)
	mux.HandleFunc("/api/2.22/quotas/groups", store.handleGroups)
	return store
}

func (s *quotaStore) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleUsersGet(w, r)
	case http.MethodPost:
		s.handleUsersPost(w, r)
	case http.MethodPatch:
		s.handleUsersPatch(w, r)
	case http.MethodDelete:
		s.handleUsersDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *quotaStore) handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGroupsGet(w, r)
	case http.MethodPost:
		s.handleGroupsPost(w, r)
	case http.MethodPatch:
		s.handleGroupsPatch(w, r)
	case http.MethodDelete:
		s.handleGroupsDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *quotaStore) handleUsersGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fsName := r.URL.Query().Get("file_system_names")
	uid := r.URL.Query().Get("uids")

	var items []client.QuotaUser

	if uid != "" && fsName != "" {
		key := quotaUserKey(fsName, uid)
		if quota, ok := s.userQuotas[key]; ok {
			items = append(items, *quota)
		}
	} else if fsName != "" {
		for key, quota := range s.userQuotas {
			if len(key) > len(fsName)+1 && key[:len(fsName)+1] == fsName+"/" {
				items = append(items, *quota)
			}
		}
	} else {
		for _, quota := range s.userQuotas {
			items = append(items, *quota)
		}
	}

	if items == nil {
		items = []client.QuotaUser{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *quotaStore) handleUsersPost(w http.ResponseWriter, r *http.Request) {
	fsName := r.URL.Query().Get("file_system_names")
	uid := r.URL.Query().Get("uids")
	if fsName == "" || uid == "" {
		WriteJSONError(w, http.StatusBadRequest, "file_system_names and uids query parameters are required for POST")
		return
	}

	var body client.QuotaUserPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaUserKey(fsName, uid)
	if _, exists := s.userQuotas[key]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("user quota for uid %q on file system %q already exists", uid, fsName))
		return
	}

	quota := &client.QuotaUser{
		FileSystem: &client.NamedReference{Name: fsName},
		User:       &client.NamedReference{Name: uid},
		Quota:      body.Quota,
	}

	s.userQuotas[key] = quota

	WriteJSONListResponse(w, http.StatusOK, []client.QuotaUser{*quota})
}

func (s *quotaStore) handleUsersPatch(w http.ResponseWriter, r *http.Request) {
	fsName := r.URL.Query().Get("file_system_names")
	uid := r.URL.Query().Get("uids")
	if fsName == "" || uid == "" {
		WriteJSONError(w, http.StatusBadRequest, "file_system_names and uids query parameters are required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaUserKey(fsName, uid)
	quota, ok := s.userQuotas[key]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("user quota for uid %q on file system %q not found", uid, fsName))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["quota"]; ok {
		var q int64
		if err := json.Unmarshal(v, &q); err == nil {
			quota.Quota = q
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.QuotaUser{*quota})
}

func (s *quotaStore) handleUsersDelete(w http.ResponseWriter, r *http.Request) {
	fsName := r.URL.Query().Get("file_system_names")
	uid := r.URL.Query().Get("uids")
	if fsName == "" || uid == "" {
		WriteJSONError(w, http.StatusBadRequest, "file_system_names and uids query parameters are required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaUserKey(fsName, uid)
	if _, ok := s.userQuotas[key]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("user quota for uid %q on file system %q not found", uid, fsName))
		return
	}

	delete(s.userQuotas, key)

	w.WriteHeader(http.StatusOK)
}

func (s *quotaStore) handleGroupsGet(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fsName := r.URL.Query().Get("file_system_names")
	gid := r.URL.Query().Get("gids")

	var items []client.QuotaGroup

	if gid != "" && fsName != "" {
		key := quotaGroupKey(fsName, gid)
		if quota, ok := s.groupQuotas[key]; ok {
			items = append(items, *quota)
		}
	} else if fsName != "" {
		for key, quota := range s.groupQuotas {
			if len(key) > len(fsName)+1 && key[:len(fsName)+1] == fsName+"/" {
				items = append(items, *quota)
			}
		}
	} else {
		for _, quota := range s.groupQuotas {
			items = append(items, *quota)
		}
	}

	if items == nil {
		items = []client.QuotaGroup{}
	}

	WriteJSONListResponse(w, http.StatusOK, items)
}

func (s *quotaStore) handleGroupsPost(w http.ResponseWriter, r *http.Request) {
	fsName := r.URL.Query().Get("file_system_names")
	gid := r.URL.Query().Get("gids")
	if fsName == "" || gid == "" {
		WriteJSONError(w, http.StatusBadRequest, "file_system_names and gids query parameters are required for POST")
		return
	}

	var body client.QuotaGroupPost
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaGroupKey(fsName, gid)
	if _, exists := s.groupQuotas[key]; exists {
		WriteJSONError(w, http.StatusConflict, fmt.Sprintf("group quota for gid %q on file system %q already exists", gid, fsName))
		return
	}

	quota := &client.QuotaGroup{
		FileSystem: &client.NamedReference{Name: fsName},
		Group:      &client.NamedReference{Name: gid},
		Quota:      body.Quota,
	}

	s.groupQuotas[key] = quota

	WriteJSONListResponse(w, http.StatusOK, []client.QuotaGroup{*quota})
}

func (s *quotaStore) handleGroupsPatch(w http.ResponseWriter, r *http.Request) {
	fsName := r.URL.Query().Get("file_system_names")
	gid := r.URL.Query().Get("gids")
	if fsName == "" || gid == "" {
		WriteJSONError(w, http.StatusBadRequest, "file_system_names and gids query parameters are required for PATCH")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaGroupKey(fsName, gid)
	quota, ok := s.groupQuotas[key]
	if !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("group quota for gid %q on file system %q not found", gid, fsName))
		return
	}

	var rawPatch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPatch); err != nil {
		WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if v, ok := rawPatch["quota"]; ok {
		var q int64
		if err := json.Unmarshal(v, &q); err == nil {
			quota.Quota = q
		}
	}

	WriteJSONListResponse(w, http.StatusOK, []client.QuotaGroup{*quota})
}

func (s *quotaStore) handleGroupsDelete(w http.ResponseWriter, r *http.Request) {
	fsName := r.URL.Query().Get("file_system_names")
	gid := r.URL.Query().Get("gids")
	if fsName == "" || gid == "" {
		WriteJSONError(w, http.StatusBadRequest, "file_system_names and gids query parameters are required for DELETE")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := quotaGroupKey(fsName, gid)
	if _, ok := s.groupQuotas[key]; !ok {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("group quota for gid %q on file system %q not found", gid, fsName))
		return
	}

	delete(s.groupQuotas, key)

	w.WriteHeader(http.StatusOK)
}

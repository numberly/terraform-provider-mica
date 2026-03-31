package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// --- User Quotas ---

func TestUnit_QuotaUser_Get(t *testing.T) {
	uid := int64(1001)
	expected := client.QuotaUser{
		FileSystem: &client.NamedReference{Name: "my-fs", ID: "fs-id-001"},
		User:       &client.NumericIDReference{ID: uid},
		Quota:      10737418240, // 10 GiB
		Usage:      1073741824,  // 1 GiB used
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/quotas/users":
			fsName := r.URL.Query().Get("file_system_names")
			uidParam := r.URL.Query().Get("uids")
			if fsName != "my-fs" || uidParam != "1001" {
				writeJSON(w, http.StatusOK, listResponse([]client.QuotaUser{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaUser{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetQuotaUser(context.Background(), "my-fs", "1001")
	if err != nil {
		t.Fatalf("GetQuotaUser: %v", err)
	}
	if got.Quota != 10737418240 {
		t.Errorf("expected Quota 10737418240, got %d", got.Quota)
	}
	if got.FileSystem == nil || got.FileSystem.Name != "my-fs" {
		t.Errorf("expected FileSystem.Name my-fs")
	}
}

func TestUnit_QuotaUser_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/quotas/users":
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaUser{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetQuotaUser(context.Background(), "my-fs", "9999")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_QuotaUser_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/quotas/users":
			fsName := r.URL.Query().Get("file_system_names")
			uid := r.URL.Query().Get("uids")
			if fsName == "" || uid == "" {
				http.Error(w, "file_system_names and uids required", http.StatusBadRequest)
				return
			}
			var body client.QuotaUserPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.QuotaUser{
				FileSystem: &client.NamedReference{Name: fsName},
				User:       &client.NumericIDReference{},
				Quota:      body.Quota,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaUser{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostQuotaUser(context.Background(), "my-fs", "1001", client.QuotaUserPost{
		Quota: 5368709120, // 5 GiB
	})
	if err != nil {
		t.Fatalf("PostQuotaUser: %v", err)
	}
	if got.Quota != 5368709120 {
		t.Errorf("expected Quota 5368709120, got %d", got.Quota)
	}
	if got.FileSystem == nil || got.FileSystem.Name != "my-fs" {
		t.Errorf("expected FileSystem.Name my-fs")
	}
}

func TestUnit_QuotaUser_Patch(t *testing.T) {
	var gotBody client.QuotaUserPatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/quotas/users":
			fsName := r.URL.Query().Get("file_system_names")
			uid := r.URL.Query().Get("uids")
			if fsName != "my-fs" || uid != "1001" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newQuota := int64(10737418240)
			if gotBody.Quota != nil {
				newQuota = *gotBody.Quota
			}
			result := client.QuotaUser{
				FileSystem: &client.NamedReference{Name: fsName},
				Quota:      newQuota,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaUser{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newQuota := int64(21474836480) // 20 GiB
	got, err := c.PatchQuotaUser(context.Background(), "my-fs", "1001", client.QuotaUserPatch{
		Quota: &newQuota,
	})
	if err != nil {
		t.Fatalf("PatchQuotaUser: %v", err)
	}
	if got.Quota != 21474836480 {
		t.Errorf("expected Quota 21474836480, got %d", got.Quota)
	}
}

func TestUnit_QuotaUser_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/quotas/users":
			fsName := r.URL.Query().Get("file_system_names")
			uid := r.URL.Query().Get("uids")
			if fsName != "my-fs" || uid != "1001" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteQuotaUser(context.Background(), "my-fs", "1001"); err != nil {
		t.Fatalf("DeleteQuotaUser: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

// --- Group Quotas ---

func TestUnit_QuotaGroup_Get(t *testing.T) {
	gid := int64(2001)
	expected := client.QuotaGroup{
		FileSystem: &client.NamedReference{Name: "my-fs", ID: "fs-id-001"},
		Group:      &client.NumericIDReference{ID: gid},
		Quota:      53687091200, // 50 GiB
		Usage:      10737418240, // 10 GiB used
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/quotas/groups":
			fsName := r.URL.Query().Get("file_system_names")
			gidParam := r.URL.Query().Get("gids")
			if fsName != "my-fs" || gidParam != "2001" {
				writeJSON(w, http.StatusOK, listResponse([]client.QuotaGroup{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaGroup{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetQuotaGroup(context.Background(), "my-fs", "2001")
	if err != nil {
		t.Fatalf("GetQuotaGroup: %v", err)
	}
	if got.Quota != 53687091200 {
		t.Errorf("expected Quota 53687091200, got %d", got.Quota)
	}
	if got.FileSystem == nil || got.FileSystem.Name != "my-fs" {
		t.Errorf("expected FileSystem.Name my-fs")
	}
}

func TestUnit_QuotaGroup_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/quotas/groups":
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaGroup{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetQuotaGroup(context.Background(), "my-fs", "9999")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_QuotaGroup_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/quotas/groups":
			fsName := r.URL.Query().Get("file_system_names")
			gid := r.URL.Query().Get("gids")
			if fsName == "" || gid == "" {
				http.Error(w, "file_system_names and gids required", http.StatusBadRequest)
				return
			}
			var body client.QuotaGroupPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.QuotaGroup{
				FileSystem: &client.NamedReference{Name: fsName},
				Group:      &client.NumericIDReference{},
				Quota:      body.Quota,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaGroup{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostQuotaGroup(context.Background(), "my-fs", "2001", client.QuotaGroupPost{
		Quota: 107374182400, // 100 GiB
	})
	if err != nil {
		t.Fatalf("PostQuotaGroup: %v", err)
	}
	if got.Quota != 107374182400 {
		t.Errorf("expected Quota 107374182400, got %d", got.Quota)
	}
	if got.FileSystem == nil || got.FileSystem.Name != "my-fs" {
		t.Errorf("expected FileSystem.Name my-fs")
	}
}

func TestUnit_QuotaGroup_Patch(t *testing.T) {
	var gotBody client.QuotaGroupPatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/quotas/groups":
			fsName := r.URL.Query().Get("file_system_names")
			gid := r.URL.Query().Get("gids")
			if fsName != "my-fs" || gid != "2001" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newQuota := int64(53687091200)
			if gotBody.Quota != nil {
				newQuota = *gotBody.Quota
			}
			result := client.QuotaGroup{
				FileSystem: &client.NamedReference{Name: fsName},
				Quota:      newQuota,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QuotaGroup{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newQuota := int64(214748364800) // 200 GiB
	got, err := c.PatchQuotaGroup(context.Background(), "my-fs", "2001", client.QuotaGroupPatch{
		Quota: &newQuota,
	})
	if err != nil {
		t.Fatalf("PatchQuotaGroup: %v", err)
	}
	if got.Quota != 214748364800 {
		t.Errorf("expected Quota 214748364800, got %d", got.Quota)
	}
}

func TestUnit_QuotaGroup_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/quotas/groups":
			fsName := r.URL.Query().Get("file_system_names")
			gid := r.URL.Query().Get("gids")
			if fsName != "my-fs" || gid != "2001" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteQuotaGroup(context.Background(), "my-fs", "2001"); err != nil {
		t.Fatalf("DeleteQuotaGroup: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

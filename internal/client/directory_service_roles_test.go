package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func newDSRServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/api/2.22/directory-services/roles", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func dsrListResponse(items ...client.DirectoryServiceRole) []byte {
	type listResp struct {
		Items []client.DirectoryServiceRole `json:"items"`
	}
	b, _ := json.Marshal(listResp{Items: items})
	return b
}

func TestUnit_DirectoryServiceRole_Get_Found(t *testing.T) {
	seed := client.DirectoryServiceRole{
		ID:    "dsr-1",
		Name:  "admin-role",
		Group: "cn=admins",
		GroupBase: "ou=corp",
		ManagementAccessPolicies: []client.NamedReference{
			{Name: "pure:policy/array_admin", ID: "map-1"},
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("names"); got != "admin-role" {
			t.Errorf("expected names=admin-role, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsrListResponse(seed))
	})
	srv := newDSRServer(t, handler)
	c := newTestClient(t, srv)

	got, err := c.GetDirectoryServiceRole(context.Background(), "admin-role")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "admin-role" {
		t.Errorf("name: %q", got.Name)
	}
	if got.Group != "cn=admins" {
		t.Errorf("group: %q", got.Group)
	}
	if got.GroupBase != "ou=corp" {
		t.Errorf("group_base: %q", got.GroupBase)
	}
	if len(got.ManagementAccessPolicies) != 1 || got.ManagementAccessPolicies[0].Name != "pure:policy/array_admin" {
		t.Errorf("management_access_policies: %+v", got.ManagementAccessPolicies)
	}
}

func TestUnit_DirectoryServiceRole_Get_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// empty list + 200 matches real API behavior
		_, _ = w.Write(dsrListResponse())
	})
	srv := newDSRServer(t, handler)
	c := newTestClient(t, srv)

	_, err := c.GetDirectoryServiceRole(context.Background(), "no-such-role")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound, got: %v", err)
	}
}

func TestUnit_DirectoryServiceRole_Post(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// POST must carry the name via ?names= query param (per D-01).
		if got := r.URL.Query().Get("names"); got != "my-role" {
			t.Errorf("expected names=my-role, got %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["group"] != "cn=admins" {
			t.Errorf("group: %v", body["group"])
		}
		if body["group_base"] != "ou=corp" {
			t.Errorf("group_base: %v", body["group_base"])
		}
		policies, ok := body["management_access_policies"].([]any)
		if !ok || len(policies) == 0 {
			t.Errorf("management_access_policies missing or empty: %v", body["management_access_policies"])
		} else {
			first, _ := policies[0].(map[string]any)
			if first["name"] != "pure:policy/array_admin" {
				t.Errorf("policies[0].name: %v", first["name"])
			}
		}
		resp := client.DirectoryServiceRole{
			ID:        "dsr-2",
			Name:      "my-role",
			Group:     "cn=admins",
			GroupBase: "ou=corp",
			ManagementAccessPolicies: []client.NamedReference{
				{Name: "pure:policy/array_admin"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsrListResponse(resp))
	})
	srv := newDSRServer(t, handler)
	c := newTestClient(t, srv)

	got, err := c.PostDirectoryServiceRole(context.Background(), "my-role", client.DirectoryServiceRolePost{
		Group:     "cn=admins",
		GroupBase: "ou=corp",
		ManagementAccessPolicies: []client.NamedReference{
			{Name: "pure:policy/array_admin"},
		},
	})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if got.Name != "my-role" {
		t.Errorf("name: %q", got.Name)
	}
}

func TestUnit_DirectoryServiceRole_Patch_Group(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if got := r.URL.Query().Get("names"); got != "admin-role" {
			t.Errorf("expected names=admin-role, got %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["group"] != "cn=new-admins" {
			t.Errorf("group: %v", body["group"])
		}
		// management_access_policies must NOT be in PATCH body
		if _, ok := body["management_access_policies"]; ok {
			t.Errorf("management_access_policies must not be present in PATCH body")
		}
		resp := client.DirectoryServiceRole{
			ID:    "dsr-1",
			Name:  "admin-role",
			Group: "cn=new-admins",
			GroupBase: "ou=corp",
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsrListResponse(resp))
	})
	srv := newDSRServer(t, handler)
	c := newTestClient(t, srv)

	newGroup := "cn=new-admins"
	got, err := c.PatchDirectoryServiceRole(context.Background(), "admin-role", client.DirectoryServiceRolePatch{
		Group: &newGroup,
	})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if got.Group != "cn=new-admins" {
		t.Errorf("group: %q", got.Group)
	}
}

func TestUnit_DirectoryServiceRole_Delete(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if got := r.URL.Query().Get("names"); got != "admin-role" {
			t.Errorf("expected names=admin-role, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := newDSRServer(t, handler)
	c := newTestClient(t, srv)

	err := c.DeleteDirectoryServiceRole(context.Background(), "admin-role")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

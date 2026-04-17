package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func newDSRMServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/api/2.22/management-access-policies/directory-services/roles", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func dsrmListResponse(items ...client.ManagementAccessPolicyDirectoryServiceRoleMembership) []byte {
	type listResp struct {
		Items []client.ManagementAccessPolicyDirectoryServiceRoleMembership `json:"items"`
	}
	b, _ := json.Marshal(listResp{Items: items})
	return b
}

func assertDSRMQueryParams(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.URL.Query().Get("policy_names"); got != "pure:policy/array_admin" {
		t.Errorf("expected policy_names=pure:policy/array_admin, got %q", got)
	}
	if got := r.URL.Query().Get("role_names"); got != "admin-role" {
		t.Errorf("expected role_names=admin-role, got %q", got)
	}
}

func TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_Get_Exists(t *testing.T) {
	seed := client.ManagementAccessPolicyDirectoryServiceRoleMembership{
		Policy: client.NamedReference{Name: "pure:policy/array_admin", ID: "map-1"},
		Role:   client.NamedReference{Name: "admin-role", ID: "dsr-1"},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		assertDSRMQueryParams(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsrmListResponse(seed))
	})
	srv := newDSRMServer(t, handler)
	c := newTestClient(t, srv)

	got, err := c.GetManagementAccessPolicyDirectoryServiceRoleMembership(context.Background(), "pure:policy/array_admin", "admin-role")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Policy.Name != "pure:policy/array_admin" {
		t.Errorf("policy.name: %q", got.Policy.Name)
	}
	if got.Role.Name != "admin-role" {
		t.Errorf("role.name: %q", got.Role.Name)
	}
}

func TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_Get_NotExists(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// empty list + 200 matches real API behavior
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsrmListResponse())
	})
	srv := newDSRMServer(t, handler)
	c := newTestClient(t, srv)

	_, err := c.GetManagementAccessPolicyDirectoryServiceRoleMembership(context.Background(), "pure:policy/array_admin", "admin-role")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound, got: %v", err)
	}
}

func TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_Post(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		assertDSRMQueryParams(t, r)
		resp := client.ManagementAccessPolicyDirectoryServiceRoleMembership{
			Policy: client.NamedReference{Name: "pure:policy/array_admin", ID: "map-1"},
			Role:   client.NamedReference{Name: "admin-role", ID: "dsr-1"},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsrmListResponse(resp))
	})
	srv := newDSRMServer(t, handler)
	c := newTestClient(t, srv)

	got, err := c.PostManagementAccessPolicyDirectoryServiceRoleMembership(context.Background(), "pure:policy/array_admin", "admin-role")
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if got.Policy.Name != "pure:policy/array_admin" {
		t.Errorf("policy.name: %q", got.Policy.Name)
	}
	if got.Role.Name != "admin-role" {
		t.Errorf("role.name: %q", got.Role.Name)
	}
}

func TestUnit_ManagementAccessPolicyDirectoryServiceRoleMembership_Delete(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		assertDSRMQueryParams(t, r)
		w.WriteHeader(http.StatusOK)
	})
	srv := newDSRMServer(t, handler)
	c := newTestClient(t, srv)

	err := c.DeleteManagementAccessPolicyDirectoryServiceRoleMembership(context.Background(), "pure:policy/array_admin", "admin-role")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

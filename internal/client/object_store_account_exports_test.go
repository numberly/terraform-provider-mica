package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_ObjectStoreAccountExport_Get(t *testing.T) {
	expected := client.ObjectStoreAccountExport{
		ID:      "osa-export-id-001",
		Name:    "test-account/_array_server",
		Enabled: true,
		Member:  &client.NamedReference{Name: "test-account"},
		Server:  &client.NamedReference{Name: "array-server-1"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-account-exports":
			name := r.URL.Query().Get("names")
			if name != "test-account/_array_server" {
				writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccountExport{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccountExport{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetObjectStoreAccountExport(context.Background(), "test-account/_array_server")
	if err != nil {
		t.Fatalf("GetObjectStoreAccountExport: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if !got.Enabled {
		t.Errorf("expected Enabled true")
	}
	if got.Member == nil || got.Member.Name != "test-account" {
		t.Errorf("expected Member.Name test-account, got %v", got.Member)
	}
}

func TestUnit_ObjectStoreAccountExport_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-account-exports":
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccountExport{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetObjectStoreAccountExport(context.Background(), "nonexistent/export")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_ObjectStoreAccountExport_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-account-exports":
			// POST uses ?member_names= and ?policy_names=
			memberName := r.URL.Query().Get("member_names")
			policyName := r.URL.Query().Get("policy_names")
			if memberName == "" || policyName == "" {
				http.Error(w, "member_names and policy_names required", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccountExportPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			export := client.ObjectStoreAccountExport{
				ID:      "osa-export-id-002",
				Name:    memberName + "/_array_server",
				Enabled: body.ExportEnabled,
				Member:  &client.NamedReference{Name: memberName},
				Policy:  &client.NamedReference{Name: policyName},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccountExport{export}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostObjectStoreAccountExport(context.Background(), "test-account", "default-policy", client.ObjectStoreAccountExportPost{
		ExportEnabled: true,
		Server:        &client.NamedReference{Name: "array-server-1"},
	})
	if err != nil {
		t.Fatalf("PostObjectStoreAccountExport: %v", err)
	}
	if got.ID != "osa-export-id-002" {
		t.Errorf("expected ID osa-export-id-002, got %q", got.ID)
	}
	if got.Member == nil || got.Member.Name != "test-account" {
		t.Errorf("expected Member.Name test-account, got %v", got.Member)
	}
	if !got.Enabled {
		t.Errorf("expected Enabled true")
	}
}

func TestUnit_ObjectStoreAccountExport_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/object-store-account-exports":
			// PATCH uses ?ids=
			id := r.URL.Query().Get("ids")
			if id != "osa-export-id-003" {
				http.Error(w, "unexpected ids param", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccountExportPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			enabled := false
			if body.ExportEnabled != nil {
				enabled = *body.ExportEnabled
			}
			export := client.ObjectStoreAccountExport{
				ID:      id,
				Name:    "patch-account/_array_server",
				Enabled: enabled,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccountExport{export}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	disabled := false
	got, err := c.PatchObjectStoreAccountExport(context.Background(), "osa-export-id-003", client.ObjectStoreAccountExportPatch{
		ExportEnabled: &disabled,
	})
	if err != nil {
		t.Fatalf("PatchObjectStoreAccountExport: %v", err)
	}
	if got.Enabled {
		t.Errorf("expected Enabled false after patch")
	}
}

func TestUnit_ObjectStoreAccountExport_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/object-store-account-exports":
			// DELETE uses ?member_names= and ?names=
			memberName := r.URL.Query().Get("member_names")
			exportName := r.URL.Query().Get("names")
			if memberName != "test-account" || exportName != "_array_server" {
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
	if err := c.DeleteObjectStoreAccountExport(context.Background(), "test-account", "_array_server"); err != nil {
		t.Fatalf("DeleteObjectStoreAccountExport: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

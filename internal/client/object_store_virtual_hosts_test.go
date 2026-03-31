package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_ObjectStoreVirtualHost_Get(t *testing.T) {
	expected := client.ObjectStoreVirtualHost{
		ID:       "vh-id-001",
		Name:     "s3.example.com",
		Hostname: "s3.example.com",
		AttachedServers: []client.NamedReference{
			{Name: "array-server-1"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-virtual-hosts":
			name := r.URL.Query().Get("names")
			if name != "s3.example.com" {
				writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreVirtualHost{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreVirtualHost{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetObjectStoreVirtualHost(context.Background(), "s3.example.com")
	if err != nil {
		t.Fatalf("GetObjectStoreVirtualHost: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Hostname != expected.Hostname {
		t.Errorf("expected Hostname %q, got %q", expected.Hostname, got.Hostname)
	}
	if len(got.AttachedServers) != 1 || got.AttachedServers[0].Name != "array-server-1" {
		t.Errorf("expected AttachedServers [array-server-1], got %v", got.AttachedServers)
	}
}

func TestUnit_ObjectStoreVirtualHost_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-virtual-hosts":
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreVirtualHost{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetObjectStoreVirtualHost(context.Background(), "nonexistent.example.com")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_ObjectStoreVirtualHost_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-virtual-hosts":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreVirtualHostPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			vh := client.ObjectStoreVirtualHost{
				ID:              "vh-id-002",
				Name:            name,
				Hostname:        body.Hostname,
				AttachedServers: body.AttachedServers,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreVirtualHost{vh}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostObjectStoreVirtualHost(context.Background(), "new.example.com", client.ObjectStoreVirtualHostPost{
		Hostname:        "new.example.com",
		AttachedServers: []client.NamedReference{{Name: "array-server-1"}},
	})
	if err != nil {
		t.Fatalf("PostObjectStoreVirtualHost: %v", err)
	}
	if got.ID != "vh-id-002" {
		t.Errorf("expected ID vh-id-002, got %q", got.ID)
	}
	if got.Hostname != "new.example.com" {
		t.Errorf("expected Hostname new.example.com, got %q", got.Hostname)
	}
	if len(got.AttachedServers) != 1 {
		t.Errorf("expected 1 attached server, got %d", len(got.AttachedServers))
	}
}

func TestUnit_ObjectStoreVirtualHost_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/object-store-virtual-hosts":
			name := r.URL.Query().Get("names")
			if name != "old.example.com" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreVirtualHostPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newHostname := "updated.example.com"
			if body.Hostname != nil {
				newHostname = *body.Hostname
			}
			vh := client.ObjectStoreVirtualHost{
				ID:       "vh-id-003",
				Name:     name,
				Hostname: newHostname,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreVirtualHost{vh}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newHostname := "updated.example.com"
	got, err := c.PatchObjectStoreVirtualHost(context.Background(), "old.example.com", client.ObjectStoreVirtualHostPatch{
		Hostname: &newHostname,
	})
	if err != nil {
		t.Fatalf("PatchObjectStoreVirtualHost: %v", err)
	}
	if got.Hostname != "updated.example.com" {
		t.Errorf("expected Hostname updated.example.com, got %q", got.Hostname)
	}
}

func TestUnit_ObjectStoreVirtualHost_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/object-store-virtual-hosts":
			name := r.URL.Query().Get("names")
			if name != "delete.example.com" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
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
	if err := c.DeleteObjectStoreVirtualHost(context.Background(), "delete.example.com"); err != nil {
		t.Fatalf("DeleteObjectStoreVirtualHost: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

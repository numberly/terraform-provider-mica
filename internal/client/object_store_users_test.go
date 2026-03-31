package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_ObjectStoreUser_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-users":
			name := r.URL.Query().Get("names")
			if name != "test-account/test-user" {
				writeJSON(w, http.StatusOK, listResponse([]map[string]any{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{
				{"name": "test-account/test-user"},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetObjectStoreUser(context.Background(), "test-account/test-user")
	if err != nil {
		t.Fatalf("GetObjectStoreUser: %v", err)
	}
}

func TestUnit_ObjectStoreUser_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-users":
			// Empty items list — triggers IsNotFound in GetObjectStoreUser
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.GetObjectStoreUser(context.Background(), "nonexistent/user")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_ObjectStoreUser_Post(t *testing.T) {
	var postCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-users":
			name := r.URL.Query().Get("names")
			if name != "test-account/new-user" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			postCalled = true
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{
				{"name": name},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.PostObjectStoreUser(context.Background(), "test-account/new-user"); err != nil {
		t.Fatalf("PostObjectStoreUser: %v", err)
	}
	if !postCalled {
		t.Error("expected POST to be called")
	}
}

func TestUnit_ObjectStoreUser_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/object-store-users":
			name := r.URL.Query().Get("names")
			if name != "test-account/delete-user" {
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
	if err := c.DeleteObjectStoreUser(context.Background(), "test-account/delete-user"); err != nil {
		t.Fatalf("DeleteObjectStoreUser: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_ObjectStoreUser_EnsureUser_Creates(t *testing.T) {
	// EnsureObjectStoreUser should POST if user doesn't exist.
	var postCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-users":
			postCalled = true
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{
				{"name": r.URL.Query().Get("names")},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.EnsureObjectStoreUser(context.Background(), "test-account/ensure-user"); err != nil {
		t.Fatalf("EnsureObjectStoreUser: %v", err)
	}
	if !postCalled {
		t.Error("expected POST to be called")
	}
}

func TestUnit_ObjectStoreUser_EnsureUser_AlreadyExists(t *testing.T) {
	// EnsureObjectStoreUser should succeed (not error) when POST returns 409 Conflict.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-users":
			// Simulate "already exists" — 409 Conflict
			writeJSON(w, http.StatusConflict, map[string]any{
				"errors": []map[string]any{
					{"message": "already exists"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.EnsureObjectStoreUser(context.Background(), "test-account/existing-user"); err != nil {
		t.Fatalf("EnsureObjectStoreUser on 409: %v", err)
	}
}

package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// ---------------------------------------------------------------------------
// Task 1 tests: typed GetObjectStoreUser / PostObjectStoreUser
// ---------------------------------------------------------------------------

func TestUnit_GetObjectStoreUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-users":
			name := r.URL.Query().Get("names")
			if name != "acct/user" {
				writeJSON(w, http.StatusOK, listResponse([]map[string]any{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{
				{"name": "acct/user", "id": "abc", "full_access": false},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	user, err := c.GetObjectStoreUser(context.Background(), "acct/user")
	if err != nil {
		t.Fatalf("GetObjectStoreUser: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil ObjectStoreUser, got nil")
	}
	if user.Name != "acct/user" {
		t.Errorf("expected Name %q, got %q", "acct/user", user.Name)
	}
	if user.ID != "abc" {
		t.Errorf("expected ID %q, got %q", "abc", user.ID)
	}
	if user.FullAccess != false {
		t.Errorf("expected FullAccess false, got %v", user.FullAccess)
	}
}

func TestUnit_GetObjectStoreUser_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-users":
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetObjectStoreUser(context.Background(), "acct/nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_PostObjectStoreUser(t *testing.T) {
	var postCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-users":
			name := r.URL.Query().Get("names")
			if name != "acct/new-user" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			postCalled = true
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{
				{"name": name, "id": "def", "full_access": false},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	user, err := c.PostObjectStoreUser(context.Background(), "acct/new-user", client.ObjectStoreUserPost{})
	if err != nil {
		t.Fatalf("PostObjectStoreUser: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil ObjectStoreUser, got nil")
	}
	if user.Name != "acct/new-user" {
		t.Errorf("expected Name %q, got %q", "acct/new-user", user.Name)
	}
	if !postCalled {
		t.Error("expected POST to be called")
	}
}

func TestUnit_PostObjectStoreUser_FullAccess(t *testing.T) {
	var bodyFullAccess *bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-users":
			var body client.ObjectStoreUserPost
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				bodyFullAccess = body.FullAccess
			}
			name := r.URL.Query().Get("names")
			fa := true
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{
				{"name": name, "id": "ghi", "full_access": fa},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	fa := true
	user, err := c.PostObjectStoreUser(context.Background(), "acct/fa-user", client.ObjectStoreUserPost{FullAccess: &fa})
	if err != nil {
		t.Fatalf("PostObjectStoreUser FullAccess: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil ObjectStoreUser, got nil")
	}
	if bodyFullAccess == nil || !*bodyFullAccess {
		t.Errorf("expected body.FullAccess=true, got %v", bodyFullAccess)
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
			if name != "acct/delete-user" {
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
	if err := c.DeleteObjectStoreUser(context.Background(), "acct/delete-user"); err != nil {
		t.Fatalf("DeleteObjectStoreUser: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_ObjectStoreUser_EnsureUser_Creates(t *testing.T) {
	var postCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-users":
			postCalled = true
			writeJSON(w, http.StatusOK, listResponse([]map[string]any{
				{"name": r.URL.Query().Get("names"), "id": "xyz", "full_access": false},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.EnsureObjectStoreUser(context.Background(), "acct/ensure-user"); err != nil {
		t.Fatalf("EnsureObjectStoreUser: %v", err)
	}
	if !postCalled {
		t.Error("expected POST to be called")
	}
}

func TestUnit_ObjectStoreUser_EnsureUser_AlreadyExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-users":
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
	if err := c.EnsureObjectStoreUser(context.Background(), "acct/existing-user"); err != nil {
		t.Fatalf("EnsureObjectStoreUser on 409: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Task 2 tests: user-policy association methods
// ---------------------------------------------------------------------------

func TestUnit_ListObjectStoreUserPolicies(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	// Register object store user handlers (including policy sub-handler)
	store := handlers.RegisterObjectStoreUserHandlers(mux, nil)
	// Pre-populate user and two policies
	store.AddPolicyForTest("acct/listuser", "policy-a")
	store.AddPolicyForTest("acct/listuser", "policy-b")

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv)
	members, err := c.ListObjectStoreUserPolicies(context.Background(), "acct/listuser")
	if err != nil {
		t.Fatalf("ListObjectStoreUserPolicies: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
}

func TestUnit_PostObjectStoreUserPolicy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	handlers.RegisterObjectStoreUserHandlers(mux, nil)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv)

	// First POST — should succeed
	member, err := c.PostObjectStoreUserPolicy(context.Background(), "acct/policyuser", "mypolicy")
	if err != nil {
		t.Fatalf("PostObjectStoreUserPolicy: %v", err)
	}
	if member == nil {
		t.Fatal("expected non-nil member, got nil")
	}
	if member.Member.Name != "acct/policyuser" {
		t.Errorf("expected Member.Name %q, got %q", "acct/policyuser", member.Member.Name)
	}
	if member.Policy.Name != "mypolicy" {
		t.Errorf("expected Policy.Name %q, got %q", "mypolicy", member.Policy.Name)
	}

	// Second POST same pair — should return error (409)
	_, err = c.PostObjectStoreUserPolicy(context.Background(), "acct/policyuser", "mypolicy")
	if err == nil {
		t.Error("expected error on duplicate PostObjectStoreUserPolicy, got nil")
	}
}

func TestUnit_DeleteObjectStoreUserPolicy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterObjectStoreUserHandlers(mux, nil)
	store.AddPolicyForTest("acct/deluser", "del-policy")

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv)

	// Delete the association
	if err := c.DeleteObjectStoreUserPolicy(context.Background(), "acct/deluser", "del-policy"); err != nil {
		t.Fatalf("DeleteObjectStoreUserPolicy: %v", err)
	}

	// Subsequent list should be empty
	members, err := c.ListObjectStoreUserPolicies(context.Background(), "acct/deluser")
	if err != nil {
		t.Fatalf("ListObjectStoreUserPolicies after delete: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("expected 0 members after delete, got %d", len(members))
	}
}

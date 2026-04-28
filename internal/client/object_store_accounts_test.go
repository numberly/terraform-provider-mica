package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_ObjectStoreAccount_Get(t *testing.T) {
	expected := client.ObjectStoreAccount{
		ID:               "acct-id-001",
		Name:             "test-account",
		QuotaLimit:       1073741824,
		HardLimitEnabled: true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-accounts":
			name := r.URL.Query().Get("names")
			if name != "test-account" {
				writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccount{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccount{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetObjectStoreAccount(context.Background(), "test-account")
	if err != nil {
		t.Fatalf("GetObjectStoreAccount: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.QuotaLimit != expected.QuotaLimit {
		t.Errorf("expected QuotaLimit %d, got %d", expected.QuotaLimit, got.QuotaLimit)
	}
	if !got.HardLimitEnabled {
		t.Errorf("expected HardLimitEnabled true")
	}
}

func TestUnit_ObjectStoreAccount_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-accounts":
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccount{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetObjectStoreAccount(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_ObjectStoreAccount_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-accounts":
			// POST uses ?names= query parameter
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccountPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			acct := client.ObjectStoreAccount{
				ID:               "acct-id-002",
				Name:             name,
				HardLimitEnabled: body.HardLimitEnabled,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccount{acct}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostObjectStoreAccount(context.Background(), "new-account", client.ObjectStoreAccountPost{
		HardLimitEnabled: true,
	})
	if err != nil {
		t.Fatalf("PostObjectStoreAccount: %v", err)
	}
	if got.ID != "acct-id-002" {
		t.Errorf("expected ID acct-id-002, got %q", got.ID)
	}
	if got.Name != "new-account" {
		t.Errorf("expected Name new-account, got %q", got.Name)
	}
	if !got.HardLimitEnabled {
		t.Errorf("expected HardLimitEnabled true")
	}
}

func TestUnit_ObjectStoreAccount_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/object-store-accounts":
			name := r.URL.Query().Get("names")
			if name != "patch-account" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccountPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			hardLimit := true
			acct := client.ObjectStoreAccount{
				ID:               "acct-id-003",
				Name:             name,
				HardLimitEnabled: hardLimit,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccount{acct}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	hardLimit := true
	got, err := c.PatchObjectStoreAccount(context.Background(), "patch-account", client.ObjectStoreAccountPatch{
		HardLimitEnabled: &hardLimit,
	})
	if err != nil {
		t.Fatalf("PatchObjectStoreAccount: %v", err)
	}
	if got.Name != "patch-account" {
		t.Errorf("expected Name patch-account, got %q", got.Name)
	}
	if !got.HardLimitEnabled {
		t.Errorf("expected HardLimitEnabled true")
	}
}

func TestUnit_ObjectStoreAccount_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/object-store-accounts":
			name := r.URL.Query().Get("names")
			if name != "delete-account" {
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
	if err := c.DeleteObjectStoreAccount(context.Background(), "delete-account"); err != nil {
		t.Fatalf("DeleteObjectStoreAccount: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_ObjectStoreAccount_List(t *testing.T) {
	accounts := []client.ObjectStoreAccount{
		{ID: "acct-id-010", Name: "list-account-1"},
		{ID: "acct-id-011", Name: "list-account-2"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-accounts":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            accounts,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListObjectStoreAccounts(context.Background(), client.ListObjectStoreAccountsOpts{})
	if err != nil {
		t.Fatalf("ListObjectStoreAccounts: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "list-account-1" {
		t.Errorf("expected first item name list-account-1, got %q", items[0].Name)
	}
}

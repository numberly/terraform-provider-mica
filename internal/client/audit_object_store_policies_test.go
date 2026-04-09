package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

func TestUnit_AuditObjectStorePolicy_Get_Found(t *testing.T) {
	expected := client.AuditObjectStorePolicy{
		ID:         "audit-policy-1",
		Name:       "my-audit-policy",
		Enabled:    true,
		IsLocal:    true,
		PolicyType: "audit",
		LogTargets: []client.NamedReference{{Name: "log-target-1"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/audit-object-store-policies":
			name := r.URL.Query().Get("names")
			if name != "my-audit-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.AuditObjectStorePolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.AuditObjectStorePolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetAuditObjectStorePolicy(context.Background(), "my-audit-policy")
	if err != nil {
		t.Fatalf("GetAuditObjectStorePolicy: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, got.Name)
	}
	if !got.Enabled {
		t.Error("expected Enabled=true")
	}
	if got.PolicyType != "audit" {
		t.Errorf("expected PolicyType %q, got %q", "audit", got.PolicyType)
	}
	if len(got.LogTargets) != 1 || got.LogTargets[0].Name != "log-target-1" {
		t.Errorf("expected LogTargets=[{log-target-1}], got %v", got.LogTargets)
	}
}

func TestUnit_AuditObjectStorePolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/audit-object-store-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.AuditObjectStorePolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetAuditObjectStorePolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_AuditObjectStorePolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/audit-object-store-policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			var body client.AuditObjectStorePolicyPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			enabled := true
			if body.Enabled != nil {
				enabled = *body.Enabled
			}
			p := client.AuditObjectStorePolicy{
				ID:         "audit-policy-42",
				Name:       name,
				Enabled:    enabled,
				PolicyType: "audit",
				LogTargets: body.LogTargets,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.AuditObjectStorePolicy{p}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled := false
	got, err := c.PostAuditObjectStorePolicy(context.Background(), "new-policy", client.AuditObjectStorePolicyPost{
		Enabled:    &enabled,
		LogTargets: []client.NamedReference{{Name: "log-target-1"}},
	})
	if err != nil {
		t.Fatalf("PostAuditObjectStorePolicy: %v", err)
	}
	if got.ID != "audit-policy-42" {
		t.Errorf("expected ID audit-policy-42, got %q", got.ID)
	}
	if got.Name != "new-policy" {
		t.Errorf("expected Name new-policy, got %q", got.Name)
	}
	if got.Enabled {
		t.Error("expected Enabled=false")
	}
	if len(got.LogTargets) != 1 || got.LogTargets[0].Name != "log-target-1" {
		t.Errorf("expected LogTargets=[{log-target-1}], got %v", got.LogTargets)
	}
}

func TestUnit_AuditObjectStorePolicy_Patch_Enabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/audit-object-store-policies":
			name := r.URL.Query().Get("names")
			if name != "existing-policy" {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			var body client.AuditObjectStorePolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			enabled := true
			if body.Enabled != nil {
				enabled = *body.Enabled
			}
			p := client.AuditObjectStorePolicy{
				ID:         "audit-policy-5",
				Name:       name,
				Enabled:    enabled,
				PolicyType: "audit",
			}
			writeJSON(w, http.StatusOK, listResponse([]client.AuditObjectStorePolicy{p}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newEnabled := false
	got, err := c.PatchAuditObjectStorePolicy(context.Background(), "existing-policy", client.AuditObjectStorePolicyPatch{
		Enabled: &newEnabled,
	})
	if err != nil {
		t.Fatalf("PatchAuditObjectStorePolicy: %v", err)
	}
	if got.Enabled {
		t.Error("expected Enabled=false after patch")
	}
	if got.Name != "existing-policy" {
		t.Errorf("expected Name existing-policy, got %q", got.Name)
	}
}

func TestUnit_AuditObjectStorePolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/audit-object-store-policies":
			name := r.URL.Query().Get("names")
			if name != "delete-policy" {
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
	if err := c.DeleteAuditObjectStorePolicy(context.Background(), "delete-policy"); err != nil {
		t.Fatalf("DeleteAuditObjectStorePolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

// TestUnit_AuditObjectStorePolicy_Get_Found_MockServer verifies the mock handler works correctly.
func TestUnit_AuditObjectStorePolicy_Get_Found_MockServer(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)

	store.Seed(&client.AuditObjectStorePolicy{
		ID:         "seeded-policy-1",
		Name:       "seeded-policy",
		Enabled:    true,
		PolicyType: "audit",
		LogTargets: []client.NamedReference{{Name: "target-a"}},
	})

	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint:           ms.URL(),
		APIToken:           "test-token",
		InsecureSkipVerify: true,
		MaxRetries:         1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	got, err := c.GetAuditObjectStorePolicy(context.Background(), "seeded-policy")
	if err != nil {
		t.Fatalf("GetAuditObjectStorePolicy: %v", err)
	}
	if got.Name != "seeded-policy" {
		t.Errorf("expected Name seeded-policy, got %q", got.Name)
	}
	if got.ID != "seeded-policy-1" {
		t.Errorf("expected ID seeded-policy-1, got %q", got.ID)
	}
}

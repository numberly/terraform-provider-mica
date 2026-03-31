package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_NfsExportPolicy_Get(t *testing.T) {
	expected := client.NfsExportPolicy{
		ID:      "nfs-pol-001",
		Name:    "my-nfs-policy",
		Enabled: true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/nfs-export-policies":
			name := r.URL.Query().Get("names")
			if name != "my-nfs-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.GetNfsExportPolicy(context.Background(), "my-nfs-policy")
	if err != nil {
		t.Fatalf("GetNfsExportPolicy: %v", err)
	}
	if pol.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, pol.ID)
	}
	if !pol.Enabled {
		t.Errorf("expected Enabled true")
	}
}

func TestUnit_NfsExportPolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/nfs-export-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetNfsExportPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_NfsExportPolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/nfs-export-policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			enabled := true
			pol := client.NfsExportPolicy{
				ID:      "nfs-pol-002",
				Name:    name,
				Enabled: enabled,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled := true
	pol, err := c.PostNfsExportPolicy(context.Background(), "new-nfs-policy", client.NfsExportPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostNfsExportPolicy: %v", err)
	}
	if pol.ID != "nfs-pol-002" {
		t.Errorf("expected ID nfs-pol-002, got %q", pol.ID)
	}
	if pol.Name != "new-nfs-policy" {
		t.Errorf("expected Name new-nfs-policy, got %q", pol.Name)
	}
}

func TestUnit_NfsExportPolicy_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/nfs-export-policies":
			name := r.URL.Query().Get("names")
			if name != "my-nfs-policy" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.NfsExportPolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newName := "renamed-nfs-policy"
			pol := client.NfsExportPolicy{
				ID:      "nfs-pol-003",
				Name:    newName,
				Enabled: false,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newName := "renamed-nfs-policy"
	pol, err := c.PatchNfsExportPolicy(context.Background(), "my-nfs-policy", client.NfsExportPolicyPatch{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("PatchNfsExportPolicy: %v", err)
	}
	if pol.Name != "renamed-nfs-policy" {
		t.Errorf("expected Name renamed-nfs-policy, got %q", pol.Name)
	}
}

func TestUnit_NfsExportPolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/nfs-export-policies":
			name := r.URL.Query().Get("names")
			if name != "my-nfs-policy" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
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
	if err := c.DeleteNfsExportPolicy(context.Background(), "my-nfs-policy"); err != nil {
		t.Fatalf("DeleteNfsExportPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_NfsExportPolicyRule_Get(t *testing.T) {
	expected := client.NfsExportPolicyRule{
		ID:    "nfs-rule-001",
		Name:  "rule-1",
		Index: 0,
		Policy: client.NamedReference{Name: "my-nfs-policy"},
		Client: "*",
		Access: "root-squash",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/nfs-export-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-nfs-policy" || ruleName != "rule-1" {
				writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicyRule{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicyRule{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetNfsExportPolicyRuleByName(context.Background(), "my-nfs-policy", "rule-1")
	if err != nil {
		t.Fatalf("GetNfsExportPolicyRuleByName: %v", err)
	}
	if rule.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, rule.ID)
	}
	if rule.Client != "*" {
		t.Errorf("expected Client *, got %q", rule.Client)
	}
}

func TestUnit_NfsExportPolicyRule_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/nfs-export-policies/rules":
			writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicyRule{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetNfsExportPolicyRuleByName(context.Background(), "my-nfs-policy", "nonexistent-rule")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_NfsExportPolicyRule_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/nfs-export-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			if policyName != "my-nfs-policy" {
				http.Error(w, "unexpected policy_names", http.StatusBadRequest)
				return
			}
			var body client.NfsExportPolicyRulePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			rule := client.NfsExportPolicyRule{
				ID:     "nfs-rule-002",
				Name:   "rule-2",
				Index:  1,
				Policy: client.NamedReference{Name: policyName},
				Client: body.Client,
				Access: body.Access,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicyRule{rule}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.PostNfsExportPolicyRule(context.Background(), "my-nfs-policy", client.NfsExportPolicyRulePost{
		Client: "10.0.0.0/8",
		Access: "root-squash",
	})
	if err != nil {
		t.Fatalf("PostNfsExportPolicyRule: %v", err)
	}
	if rule.ID != "nfs-rule-002" {
		t.Errorf("expected ID nfs-rule-002, got %q", rule.ID)
	}
	if rule.Client != "10.0.0.0/8" {
		t.Errorf("expected Client 10.0.0.0/8, got %q", rule.Client)
	}
}

func TestUnit_NfsExportPolicyRule_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/nfs-export-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-nfs-policy" || ruleName != "rule-1" {
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
	if err := c.DeleteNfsExportPolicyRule(context.Background(), "my-nfs-policy", "rule-1"); err != nil {
		t.Fatalf("DeleteNfsExportPolicyRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_NfsExportPolicyRule_List(t *testing.T) {
	rules := []client.NfsExportPolicyRule{
		{ID: "nfs-rule-010", Name: "rule-a", Index: 0, Policy: client.NamedReference{Name: "my-nfs-policy"}},
		{ID: "nfs-rule-011", Name: "rule-b", Index: 1, Policy: client.NamedReference{Name: "my-nfs-policy"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/nfs-export-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			if policyName != "my-nfs-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.NfsExportPolicyRule{}))
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            rules,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListNfsExportPolicyRules(context.Background(), "my-nfs-policy")
	if err != nil {
		t.Fatalf("ListNfsExportPolicyRules: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 rules, got %d", len(items))
	}
	if items[0].Name != "rule-a" {
		t.Errorf("expected first rule rule-a, got %q", items[0].Name)
	}
}

func TestUnit_NfsExportPolicyRule_GetByIndex(t *testing.T) {
	rules := []client.NfsExportPolicyRule{
		{ID: "nfs-rule-020", Name: "rule-idx-0", Index: 0, Policy: client.NamedReference{Name: "my-nfs-policy"}},
		{ID: "nfs-rule-021", Name: "rule-idx-1", Index: 1, Policy: client.NamedReference{Name: "my-nfs-policy"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/nfs-export-policies/rules":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            rules,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetNfsExportPolicyRuleByIndex(context.Background(), "my-nfs-policy", 1)
	if err != nil {
		t.Fatalf("GetNfsExportPolicyRuleByIndex: %v", err)
	}
	if rule.Name != "rule-idx-1" {
		t.Errorf("expected Name rule-idx-1, got %q", rule.Name)
	}
}

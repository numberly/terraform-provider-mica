package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_SmbClientPolicy_Get(t *testing.T) {
	expected := client.SmbClientPolicy{
		ID:      "smb-client-pol-001",
		Name:    "my-smb-client-policy",
		Enabled: true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-client-policies":
			name := r.URL.Query().Get("names")
			if name != "my-smb-client-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.GetSmbClientPolicy(context.Background(), "my-smb-client-policy")
	if err != nil {
		t.Fatalf("GetSmbClientPolicy: %v", err)
	}
	if pol.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, pol.ID)
	}
	if !pol.Enabled {
		t.Errorf("expected Enabled true")
	}
}

func TestUnit_SmbClientPolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-client-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetSmbClientPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_SmbClientPolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/smb-client-policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			pol := client.SmbClientPolicy{
				ID:      "smb-client-pol-002",
				Name:    name,
				Enabled: true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled := true
	pol, err := c.PostSmbClientPolicy(context.Background(), "new-smb-client-policy", client.SmbClientPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostSmbClientPolicy: %v", err)
	}
	if pol.ID != "smb-client-pol-002" {
		t.Errorf("expected ID smb-client-pol-002, got %q", pol.ID)
	}
	if pol.Name != "new-smb-client-policy" {
		t.Errorf("expected Name new-smb-client-policy, got %q", pol.Name)
	}
}

func TestUnit_SmbClientPolicy_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/smb-client-policies":
			name := r.URL.Query().Get("names")
			if name != "my-smb-client-policy" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.SmbClientPolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			pol := client.SmbClientPolicy{
				ID:      "smb-client-pol-003",
				Name:    "renamed-smb-client-policy",
				Enabled: false,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newName := "renamed-smb-client-policy"
	pol, err := c.PatchSmbClientPolicy(context.Background(), "my-smb-client-policy", client.SmbClientPolicyPatch{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("PatchSmbClientPolicy: %v", err)
	}
	if pol.Name != "renamed-smb-client-policy" {
		t.Errorf("expected Name renamed-smb-client-policy, got %q", pol.Name)
	}
}

func TestUnit_SmbClientPolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/smb-client-policies":
			name := r.URL.Query().Get("names")
			if name != "my-smb-client-policy" {
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
	if err := c.DeleteSmbClientPolicy(context.Background(), "my-smb-client-policy"); err != nil {
		t.Fatalf("DeleteSmbClientPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_SmbClientPolicyRule_Get(t *testing.T) {
	expected := client.SmbClientPolicyRule{
		ID:         "smb-client-rule-001",
		Name:       "allow-internal",
		Index:      0,
		Policy:     client.NamedReference{Name: "my-smb-client-policy"},
		Client:     "10.0.0.0/8",
		Permission: "read-write",
		Encryption: "optional",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-client-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-smb-client-policy" || ruleName != "allow-internal" {
				writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicyRule{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicyRule{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetSmbClientPolicyRuleByName(context.Background(), "my-smb-client-policy", "allow-internal")
	if err != nil {
		t.Fatalf("GetSmbClientPolicyRuleByName: %v", err)
	}
	if rule.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, rule.ID)
	}
	if rule.Client != "10.0.0.0/8" {
		t.Errorf("expected Client 10.0.0.0/8, got %q", rule.Client)
	}
}

func TestUnit_SmbClientPolicyRule_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-client-policies/rules":
			writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicyRule{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetSmbClientPolicyRuleByName(context.Background(), "my-smb-client-policy", "nonexistent-rule")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_SmbClientPolicyRule_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/smb-client-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			if policyName != "my-smb-client-policy" {
				http.Error(w, "unexpected policy_names", http.StatusBadRequest)
				return
			}
			var body client.SmbClientPolicyRulePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			rule := client.SmbClientPolicyRule{
				ID:         "smb-client-rule-002",
				Name:       "new-rule",
				Policy:     client.NamedReference{Name: policyName},
				Client:     body.Client,
				Permission: body.Permission,
				Encryption: body.Encryption,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicyRule{rule}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.PostSmbClientPolicyRule(context.Background(), "my-smb-client-policy", client.SmbClientPolicyRulePost{
		Client:     "192.168.0.0/16",
		Permission: "read-write",
		Encryption: "required",
	})
	if err != nil {
		t.Fatalf("PostSmbClientPolicyRule: %v", err)
	}
	if rule.ID != "smb-client-rule-002" {
		t.Errorf("expected ID smb-client-rule-002, got %q", rule.ID)
	}
	if rule.Client != "192.168.0.0/16" {
		t.Errorf("expected Client 192.168.0.0/16, got %q", rule.Client)
	}
}

func TestUnit_SmbClientPolicyRule_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/smb-client-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-smb-client-policy" || ruleName != "allow-internal" {
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
	if err := c.DeleteSmbClientPolicyRule(context.Background(), "my-smb-client-policy", "allow-internal"); err != nil {
		t.Fatalf("DeleteSmbClientPolicyRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_SmbClientPolicyRule_List(t *testing.T) {
	rules := []client.SmbClientPolicyRule{
		{ID: "smb-client-rule-010", Name: "rule-a", Index: 0, Policy: client.NamedReference{Name: "my-smb-client-policy"}, Client: "10.0.0.0/8"},
		{ID: "smb-client-rule-011", Name: "rule-b", Index: 1, Policy: client.NamedReference{Name: "my-smb-client-policy"}, Client: "192.168.0.0/16"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-client-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			if policyName != "my-smb-client-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.SmbClientPolicyRule{}))
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
	items, err := c.ListSmbClientPolicyRules(context.Background(), "my-smb-client-policy")
	if err != nil {
		t.Fatalf("ListSmbClientPolicyRules: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 rules, got %d", len(items))
	}
	if items[0].Name != "rule-a" {
		t.Errorf("expected first rule rule-a, got %q", items[0].Name)
	}
}

package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_SmbSharePolicy_Get(t *testing.T) {
	expected := client.SmbSharePolicy{
		ID:      "smb-share-pol-001",
		Name:    "my-smb-share-policy",
		Enabled: true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-share-policies":
			name := r.URL.Query().Get("names")
			if name != "my-smb-share-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.GetSmbSharePolicy(context.Background(), "my-smb-share-policy")
	if err != nil {
		t.Fatalf("GetSmbSharePolicy: %v", err)
	}
	if pol.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, pol.ID)
	}
	if !pol.Enabled {
		t.Errorf("expected Enabled true")
	}
}

func TestUnit_SmbSharePolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-share-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetSmbSharePolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_SmbSharePolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/smb-share-policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			pol := client.SmbSharePolicy{
				ID:      "smb-share-pol-002",
				Name:    name,
				Enabled: true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled := true
	pol, err := c.PostSmbSharePolicy(context.Background(), "new-smb-share-policy", client.SmbSharePolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostSmbSharePolicy: %v", err)
	}
	if pol.ID != "smb-share-pol-002" {
		t.Errorf("expected ID smb-share-pol-002, got %q", pol.ID)
	}
	if pol.Name != "new-smb-share-policy" {
		t.Errorf("expected Name new-smb-share-policy, got %q", pol.Name)
	}
}

func TestUnit_SmbSharePolicy_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/smb-share-policies":
			name := r.URL.Query().Get("names")
			if name != "my-smb-share-policy" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.SmbSharePolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			pol := client.SmbSharePolicy{
				ID:      "smb-share-pol-003",
				Name:    "renamed-smb-share-policy",
				Enabled: false,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newName := "renamed-smb-share-policy"
	pol, err := c.PatchSmbSharePolicy(context.Background(), "my-smb-share-policy", client.SmbSharePolicyPatch{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("PatchSmbSharePolicy: %v", err)
	}
	if pol.Name != "renamed-smb-share-policy" {
		t.Errorf("expected Name renamed-smb-share-policy, got %q", pol.Name)
	}
}

func TestUnit_SmbSharePolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/smb-share-policies":
			name := r.URL.Query().Get("names")
			if name != "my-smb-share-policy" {
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
	if err := c.DeleteSmbSharePolicy(context.Background(), "my-smb-share-policy"); err != nil {
		t.Fatalf("DeleteSmbSharePolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_SmbSharePolicyRule_Get(t *testing.T) {
	expected := client.SmbSharePolicyRule{
		ID:          "smb-share-rule-001",
		Name:        "allow-everyone",
		Policy:      client.NamedReference{Name: "my-smb-share-policy"},
		Principal:   "Everyone",
		FullControl: "allow",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-share-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-smb-share-policy" || ruleName != "allow-everyone" {
				writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicyRule{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicyRule{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetSmbSharePolicyRuleByName(context.Background(), "my-smb-share-policy", "allow-everyone")
	if err != nil {
		t.Fatalf("GetSmbSharePolicyRuleByName: %v", err)
	}
	if rule.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, rule.ID)
	}
	if rule.Principal != "Everyone" {
		t.Errorf("expected Principal Everyone, got %q", rule.Principal)
	}
}

func TestUnit_SmbSharePolicyRule_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-share-policies/rules":
			writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicyRule{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetSmbSharePolicyRuleByName(context.Background(), "my-smb-share-policy", "nonexistent-rule")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_SmbSharePolicyRule_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/smb-share-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			if policyName != "my-smb-share-policy" {
				http.Error(w, "unexpected policy_names", http.StatusBadRequest)
				return
			}
			var body client.SmbSharePolicyRulePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			rule := client.SmbSharePolicyRule{
				ID:          "smb-share-rule-002",
				Name:        "new-rule",
				Policy:      client.NamedReference{Name: policyName},
				Principal:   body.Principal,
				FullControl: body.FullControl,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicyRule{rule}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.PostSmbSharePolicyRule(context.Background(), "my-smb-share-policy", client.SmbSharePolicyRulePost{
		Principal:   "DOMAIN\\user",
		FullControl: "allow",
	})
	if err != nil {
		t.Fatalf("PostSmbSharePolicyRule: %v", err)
	}
	if rule.ID != "smb-share-rule-002" {
		t.Errorf("expected ID smb-share-rule-002, got %q", rule.ID)
	}
	if rule.Principal != "DOMAIN\\user" {
		t.Errorf("expected Principal DOMAIN\\user, got %q", rule.Principal)
	}
}

func TestUnit_SmbSharePolicyRule_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/smb-share-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-smb-share-policy" || ruleName != "allow-everyone" {
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
	if err := c.DeleteSmbSharePolicyRule(context.Background(), "my-smb-share-policy", "allow-everyone"); err != nil {
		t.Fatalf("DeleteSmbSharePolicyRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_SmbSharePolicyRule_List(t *testing.T) {
	rules := []client.SmbSharePolicyRule{
		{ID: "smb-share-rule-010", Name: "rule-a", Policy: client.NamedReference{Name: "my-smb-share-policy"}, Principal: "Everyone"},
		{ID: "smb-share-rule-011", Name: "rule-b", Policy: client.NamedReference{Name: "my-smb-share-policy"}, Principal: "DOMAIN\\admin"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smb-share-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			if policyName != "my-smb-share-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.SmbSharePolicyRule{}))
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
	items, err := c.ListSmbSharePolicyRules(context.Background(), "my-smb-share-policy")
	if err != nil {
		t.Fatalf("ListSmbSharePolicyRules: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 rules, got %d", len(items))
	}
	if items[0].Name != "rule-a" {
		t.Errorf("expected first rule rule-a, got %q", items[0].Name)
	}
}

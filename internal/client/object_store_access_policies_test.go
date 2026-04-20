package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_ObjectStoreAccessPolicy_Get(t *testing.T) {
	expected := client.ObjectStoreAccessPolicy{
		ID:          "oap-001",
		Name:        "my-oap",
		Enabled:     true,
		Description: "Test policy",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-access-policies":
			name := r.URL.Query().Get("names")
			if name != "my-oap" {
				writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.GetObjectStoreAccessPolicy(context.Background(), "my-oap")
	if err != nil {
		t.Fatalf("GetObjectStoreAccessPolicy: %v", err)
	}
	if pol.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, pol.ID)
	}
	if pol.Description != "Test policy" {
		t.Errorf("expected Description 'Test policy', got %q", pol.Description)
	}
}

func TestUnit_ObjectStoreAccessPolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-access-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetObjectStoreAccessPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_ObjectStoreAccessPolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-access-policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccessPolicyPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			pol := client.ObjectStoreAccessPolicy{
				ID:          "oap-002",
				Name:        name,
				Description: body.Description,
				Enabled:     true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.PostObjectStoreAccessPolicy(context.Background(), "new-oap", client.ObjectStoreAccessPolicyPost{
		Description: "New policy for testing",
	})
	if err != nil {
		t.Fatalf("PostObjectStoreAccessPolicy: %v", err)
	}
	if pol.ID != "oap-002" {
		t.Errorf("expected ID oap-002, got %q", pol.ID)
	}
	if pol.Name != "new-oap" {
		t.Errorf("expected Name new-oap, got %q", pol.Name)
	}
	if pol.Description != "New policy for testing" {
		t.Errorf("expected Description 'New policy for testing', got %q", pol.Description)
	}
}

func TestUnit_ObjectStoreAccessPolicy_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/object-store-access-policies":
			name := r.URL.Query().Get("names")
			if name != "my-oap" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccessPolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			pol := client.ObjectStoreAccessPolicy{
				ID:      "oap-003",
				Name:    "renamed-oap",
				Enabled: true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newName := "renamed-oap"
	pol, err := c.PatchObjectStoreAccessPolicy(context.Background(), "my-oap", client.ObjectStoreAccessPolicyPatch{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("PatchObjectStoreAccessPolicy: %v", err)
	}
	if pol.Name != "renamed-oap" {
		t.Errorf("expected Name renamed-oap, got %q", pol.Name)
	}
}

func TestUnit_ObjectStoreAccessPolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/object-store-access-policies":
			name := r.URL.Query().Get("names")
			if name != "my-oap" {
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
	if err := c.DeleteObjectStoreAccessPolicy(context.Background(), "my-oap"); err != nil {
		t.Fatalf("DeleteObjectStoreAccessPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_ObjectStoreAccessPolicy_List(t *testing.T) {
	policies := []client.ObjectStoreAccessPolicy{
		{ID: "oap-010", Name: "policy-a", Enabled: true},
		{ID: "oap-011", Name: "policy-b", Enabled: false},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-access-policies":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            policies,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListObjectStoreAccessPolicies(context.Background())
	if err != nil {
		t.Fatalf("ListObjectStoreAccessPolicies: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "policy-a" {
		t.Errorf("expected first item policy-a, got %q", items[0].Name)
	}
}

func TestUnit_ObjectStoreAccessPolicyRule_Get(t *testing.T) {
	expected := client.ObjectStoreAccessPolicyRule{
		Name:      "allow-get-object",
		Effect:    "Allow",
		Actions:   []string{"s3:GetObject"},
		Resources: []string{"arn:aws:s3:::my-bucket/*"},
		Policy:    &client.NamedReference{Name: "my-oap"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-access-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-oap" || ruleName != "allow-get-object" {
				writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicyRule{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicyRule{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetObjectStoreAccessPolicyRuleByName(context.Background(), "my-oap", "allow-get-object")
	if err != nil {
		t.Fatalf("GetObjectStoreAccessPolicyRuleByName: %v", err)
	}
	if rule.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, rule.Name)
	}
	if rule.Effect != "Allow" {
		t.Errorf("expected Effect Allow, got %q", rule.Effect)
	}
	if len(rule.Actions) != 1 || rule.Actions[0] != "s3:GetObject" {
		t.Errorf("expected Actions [s3:GetObject], got %v", rule.Actions)
	}
}

func TestUnit_ObjectStoreAccessPolicyRule_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-access-policies/rules":
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicyRule{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetObjectStoreAccessPolicyRuleByName(context.Background(), "my-oap", "nonexistent-rule")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_ObjectStoreAccessPolicyRule_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-access-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-oap" || ruleName != "new-rule" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccessPolicyRulePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			rule := client.ObjectStoreAccessPolicyRule{
				Name:      ruleName,
				Effect:    body.Effect,
				Actions:   body.Actions,
				Resources: body.Resources,
				Policy:    &client.NamedReference{Name: policyName},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicyRule{rule}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.PostObjectStoreAccessPolicyRule(context.Background(), "my-oap", "new-rule", client.ObjectStoreAccessPolicyRulePost{
		Effect:    "Allow",
		Actions:   []string{"s3:PutObject", "s3:GetObject"},
		Resources: []string{"arn:aws:s3:::my-bucket/*"},
	})
	if err != nil {
		t.Fatalf("PostObjectStoreAccessPolicyRule: %v", err)
	}
	if rule.Name != "new-rule" {
		t.Errorf("expected Name new-rule, got %q", rule.Name)
	}
	if rule.Effect != "Allow" {
		t.Errorf("expected Effect Allow, got %q", rule.Effect)
	}
	if len(rule.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(rule.Actions))
	}
}

func TestUnit_ObjectStoreAccessPolicyRule_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/object-store-access-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-oap" || ruleName != "allow-get-object" {
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
	if err := c.DeleteObjectStoreAccessPolicyRule(context.Background(), "my-oap", "allow-get-object"); err != nil {
		t.Fatalf("DeleteObjectStoreAccessPolicyRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_ObjectStoreAccessPolicyRule_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/object-store-access-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-oap" || ruleName != "allow-get-object" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreAccessPolicyRulePatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			rule := client.ObjectStoreAccessPolicyRule{
				Name:   ruleName,
				Effect: "Allow",
				Policy: &client.NamedReference{Name: policyName},
			}
			if body.Actions != nil {
				rule.Actions = *body.Actions
			}
			if body.Resources != nil {
				rule.Resources = *body.Resources
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicyRule{rule}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.PatchObjectStoreAccessPolicyRule(context.Background(), "my-oap", "allow-get-object", client.ObjectStoreAccessPolicyRulePatch{
		Actions:   &[]string{"s3:GetObject", "s3:ListBucket"},
		Resources: &[]string{"arn:aws:s3:::my-bucket", "arn:aws:s3:::my-bucket/*"},
	})
	if err != nil {
		t.Fatalf("PatchObjectStoreAccessPolicyRule: %v", err)
	}
	if rule.Name != "allow-get-object" {
		t.Errorf("expected Name allow-get-object, got %q", rule.Name)
	}
	if len(rule.Actions) != 2 {
		t.Errorf("expected 2 actions after patch, got %d", len(rule.Actions))
	}
}

func TestUnit_ObjectStoreAccessPolicyRule_Patch_ClearList(t *testing.T) {
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/object-store-access-policies/rules":
			capturedBody, _ = io.ReadAll(r.Body)
			rule := client.ObjectStoreAccessPolicyRule{
				Name:      r.URL.Query().Get("names"),
				Effect:    "Allow",
				Actions:   []string{},
				Resources: []string{},
				Policy:    &client.NamedReference{Name: r.URL.Query().Get("policy_names")},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreAccessPolicyRule{rule}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.PatchObjectStoreAccessPolicyRule(context.Background(), "my-oap", "allow-get-object", client.ObjectStoreAccessPolicyRulePatch{
		Actions:   &[]string{},
		Resources: &[]string{},
	})
	if err != nil {
		t.Fatalf("PatchObjectStoreAccessPolicyRule: %v", err)
	}
	if !bytes.Contains(capturedBody, []byte(`"actions":[]`)) {
		t.Errorf("expected body to contain \"actions\":[], got %s", capturedBody)
	}
	if !bytes.Contains(capturedBody, []byte(`"resources":[]`)) {
		t.Errorf("expected body to contain \"resources\":[], got %s", capturedBody)
	}
	if len(rule.Actions) != 0 {
		t.Errorf("expected empty Actions, got %v", rule.Actions)
	}
}

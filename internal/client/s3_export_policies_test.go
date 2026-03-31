package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_S3ExportPolicy_Get(t *testing.T) {
	expected := client.S3ExportPolicy{
		ID:      "s3-pol-001",
		Name:    "my-s3-policy",
		Enabled: true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/s3-export-policies":
			name := r.URL.Query().Get("names")
			if name != "my-s3-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.GetS3ExportPolicy(context.Background(), "my-s3-policy")
	if err != nil {
		t.Fatalf("GetS3ExportPolicy: %v", err)
	}
	if pol.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, pol.ID)
	}
	if !pol.Enabled {
		t.Errorf("expected Enabled true")
	}
}

func TestUnit_S3ExportPolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/s3-export-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetS3ExportPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_S3ExportPolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/s3-export-policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			pol := client.S3ExportPolicy{
				ID:      "s3-pol-002",
				Name:    name,
				Enabled: true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled := true
	pol, err := c.PostS3ExportPolicy(context.Background(), "new-s3-policy", client.S3ExportPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostS3ExportPolicy: %v", err)
	}
	if pol.ID != "s3-pol-002" {
		t.Errorf("expected ID s3-pol-002, got %q", pol.ID)
	}
	if pol.Name != "new-s3-policy" {
		t.Errorf("expected Name new-s3-policy, got %q", pol.Name)
	}
}

func TestUnit_S3ExportPolicy_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/s3-export-policies":
			name := r.URL.Query().Get("names")
			if name != "my-s3-policy" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.S3ExportPolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			pol := client.S3ExportPolicy{
				ID:      "s3-pol-003",
				Name:    "renamed-s3-policy",
				Enabled: false,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newName := "renamed-s3-policy"
	pol, err := c.PatchS3ExportPolicy(context.Background(), "my-s3-policy", client.S3ExportPolicyPatch{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("PatchS3ExportPolicy: %v", err)
	}
	if pol.Name != "renamed-s3-policy" {
		t.Errorf("expected Name renamed-s3-policy, got %q", pol.Name)
	}
}

func TestUnit_S3ExportPolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/s3-export-policies":
			name := r.URL.Query().Get("names")
			if name != "my-s3-policy" {
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
	if err := c.DeleteS3ExportPolicy(context.Background(), "my-s3-policy"); err != nil {
		t.Fatalf("DeleteS3ExportPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_S3ExportPolicyRule_Get(t *testing.T) {
	expected := client.S3ExportPolicyRule{
		ID:        "s3-rule-001",
		Name:      "allow-all",
		Index:     0,
		Policy:    client.NamedReference{Name: "my-s3-policy"},
		Effect:    "allow",
		Actions:   []string{"s3:GetObject"},
		Resources: []string{"*"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/s3-export-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-s3-policy" || ruleName != "allow-all" {
				writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicyRule{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicyRule{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetS3ExportPolicyRuleByName(context.Background(), "my-s3-policy", "allow-all")
	if err != nil {
		t.Fatalf("GetS3ExportPolicyRuleByName: %v", err)
	}
	if rule.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, rule.ID)
	}
	if rule.Effect != "allow" {
		t.Errorf("expected Effect allow, got %q", rule.Effect)
	}
}

func TestUnit_S3ExportPolicyRule_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/s3-export-policies/rules":
			writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicyRule{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetS3ExportPolicyRuleByName(context.Background(), "my-s3-policy", "nonexistent-rule")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_S3ExportPolicyRule_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/s3-export-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-s3-policy" || ruleName != "new-rule" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			var body client.S3ExportPolicyRulePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			rule := client.S3ExportPolicyRule{
				ID:        "s3-rule-002",
				Name:      ruleName,
				Index:     0,
				Policy:    client.NamedReference{Name: policyName},
				Effect:    body.Effect,
				Actions:   body.Actions,
				Resources: body.Resources,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.S3ExportPolicyRule{rule}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.PostS3ExportPolicyRule(context.Background(), "my-s3-policy", "new-rule", client.S3ExportPolicyRulePost{
		Effect:    "allow",
		Actions:   []string{"s3:GetObject", "s3:PutObject"},
		Resources: []string{"arn:aws:s3:::my-bucket/*"},
	})
	if err != nil {
		t.Fatalf("PostS3ExportPolicyRule: %v", err)
	}
	if rule.ID != "s3-rule-002" {
		t.Errorf("expected ID s3-rule-002, got %q", rule.ID)
	}
	if rule.Effect != "allow" {
		t.Errorf("expected Effect allow, got %q", rule.Effect)
	}
	if len(rule.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(rule.Actions))
	}
}

func TestUnit_S3ExportPolicyRule_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/s3-export-policies/rules":
			policyName := r.URL.Query().Get("policy_names")
			ruleName := r.URL.Query().Get("names")
			if policyName != "my-s3-policy" || ruleName != "allow-all" {
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
	if err := c.DeleteS3ExportPolicyRule(context.Background(), "my-s3-policy", "allow-all"); err != nil {
		t.Fatalf("DeleteS3ExportPolicyRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_S3ExportPolicyRule_GetByIndex(t *testing.T) {
	rules := []client.S3ExportPolicyRule{
		{ID: "s3-rule-010", Name: "rule-0", Index: 0, Policy: client.NamedReference{Name: "my-s3-policy"}, Effect: "allow"},
		{ID: "s3-rule-011", Name: "rule-1", Index: 1, Policy: client.NamedReference{Name: "my-s3-policy"}, Effect: "deny"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/s3-export-policies/rules":
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
	rule, err := c.GetS3ExportPolicyRuleByIndex(context.Background(), "my-s3-policy", 1)
	if err != nil {
		t.Fatalf("GetS3ExportPolicyRuleByIndex: %v", err)
	}
	if rule.Name != "rule-1" {
		t.Errorf("expected Name rule-1, got %q", rule.Name)
	}
	if rule.Effect != "deny" {
		t.Errorf("expected Effect deny, got %q", rule.Effect)
	}
}

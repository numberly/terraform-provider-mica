package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_SnapshotPolicy_Get(t *testing.T) {
	every := int64(3600000)
	keepFor := int64(86400000)
	expected := client.SnapshotPolicy{
		ID:      "snap-pol-001",
		Name:    "my-snapshot-policy",
		Enabled: true,
		Rules: []client.SnapshotPolicyRuleInPolicy{
			{Every: &every, KeepFor: &keepFor},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/policies":
			name := r.URL.Query().Get("names")
			if name != "my-snapshot-policy" {
				writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.GetSnapshotPolicy(context.Background(), "my-snapshot-policy")
	if err != nil {
		t.Fatalf("GetSnapshotPolicy: %v", err)
	}
	if pol.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, pol.ID)
	}
	if !pol.Enabled {
		t.Errorf("expected Enabled true")
	}
	if len(pol.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(pol.Rules))
	}
}

func TestUnit_SnapshotPolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/policies":
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetSnapshotPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_SnapshotPolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			var body client.SnapshotPolicyPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			pol := client.SnapshotPolicy{
				ID:      "snap-pol-002",
				Name:    name,
				Enabled: true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled := true
	pol, err := c.PostSnapshotPolicy(context.Background(), "new-snapshot-policy", client.SnapshotPolicyPost{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("PostSnapshotPolicy: %v", err)
	}
	if pol.ID != "snap-pol-002" {
		t.Errorf("expected ID snap-pol-002, got %q", pol.ID)
	}
	if pol.Name != "new-snapshot-policy" {
		t.Errorf("expected Name new-snapshot-policy, got %q", pol.Name)
	}
}

func TestUnit_SnapshotPolicy_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/policies":
			name := r.URL.Query().Get("names")
			if name != "my-snapshot-policy" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.SnapshotPolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			pol := client.SnapshotPolicy{
				ID:      "snap-pol-003",
				Name:    "my-snapshot-policy",
				Enabled: false,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	disabled := false
	pol, err := c.PatchSnapshotPolicy(context.Background(), "my-snapshot-policy", client.SnapshotPolicyPatch{
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("PatchSnapshotPolicy: %v", err)
	}
	if pol.Enabled {
		t.Errorf("expected Enabled false after PATCH")
	}
}

func TestUnit_SnapshotPolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/policies":
			name := r.URL.Query().Get("names")
			if name != "my-snapshot-policy" {
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
	if err := c.DeleteSnapshotPolicy(context.Background(), "my-snapshot-policy"); err != nil {
		t.Fatalf("DeleteSnapshotPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_SnapshotPolicy_AddRule(t *testing.T) {
	every := int64(3600000)
	keepFor := int64(86400000)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/policies":
			name := r.URL.Query().Get("names")
			if name != "my-snapshot-policy" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			var body client.SnapshotPolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if len(body.AddRules) == 0 {
				http.Error(w, "expected add_rules in body", http.StatusBadRequest)
				return
			}
			pol := client.SnapshotPolicy{
				ID:      "snap-pol-004",
				Name:    name,
				Enabled: true,
				Rules: []client.SnapshotPolicyRuleInPolicy{
					{Every: &every, KeepFor: &keepFor},
				},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.PostSnapshotPolicyRule(context.Background(), "my-snapshot-policy", client.SnapshotPolicyRulePost{
		Every:   &every,
		KeepFor: &keepFor,
	})
	if err != nil {
		t.Fatalf("PostSnapshotPolicyRule: %v", err)
	}
	if len(pol.Rules) != 1 {
		t.Errorf("expected 1 rule after add, got %d", len(pol.Rules))
	}
}

func TestUnit_SnapshotPolicy_RemoveRule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/policies":
			var body client.SnapshotPolicyPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			if len(body.RemoveRules) == 0 {
				http.Error(w, "expected remove_rules in body", http.StatusBadRequest)
				return
			}
			pol := client.SnapshotPolicy{
				ID:      "snap-pol-005",
				Name:    "my-snapshot-policy",
				Enabled: true,
				Rules:   []client.SnapshotPolicyRuleInPolicy{},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{pol}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	pol, err := c.DeleteSnapshotPolicyRule(context.Background(), "my-snapshot-policy", client.SnapshotPolicyRuleRemove{
		Every:   3600000,
		KeepFor: 86400000,
	})
	if err != nil {
		t.Fatalf("DeleteSnapshotPolicyRule: %v", err)
	}
	if len(pol.Rules) != 0 {
		t.Errorf("expected 0 rules after remove, got %d", len(pol.Rules))
	}
}

func TestUnit_SnapshotPolicy_GetRuleByIndex(t *testing.T) {
	every0 := int64(3600000)
	keepFor0 := int64(86400000)
	every1 := int64(7200000)
	keepFor1 := int64(172800000)

	expected := client.SnapshotPolicy{
		ID:      "snap-pol-006",
		Name:    "my-snapshot-policy",
		Enabled: true,
		Rules: []client.SnapshotPolicyRuleInPolicy{
			{Every: &every0, KeepFor: &keepFor0},
			{Every: &every1, KeepFor: &keepFor1},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/policies":
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetSnapshotPolicyRuleByIndex(context.Background(), "my-snapshot-policy", 1)
	if err != nil {
		t.Fatalf("GetSnapshotPolicyRuleByIndex: %v", err)
	}
	if rule.Every == nil || *rule.Every != every1 {
		t.Errorf("expected Every %d, got %v", every1, rule.Every)
	}
}

func TestUnit_SnapshotPolicy_GetRuleByIndex_NotFound(t *testing.T) {
	every0 := int64(3600000)
	keepFor0 := int64(86400000)

	expected := client.SnapshotPolicy{
		ID:      "snap-pol-007",
		Name:    "my-snapshot-policy",
		Enabled: true,
		Rules: []client.SnapshotPolicyRuleInPolicy{
			{Every: &every0, KeepFor: &keepFor0},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/policies":
			writeJSON(w, http.StatusOK, listResponse([]client.SnapshotPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	// Index 5 is out of range (only 1 rule at index 0)
	_, err := c.GetSnapshotPolicyRuleByIndex(context.Background(), "my-snapshot-policy", 5)
	if err == nil {
		t.Fatal("expected error for out-of-range index, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

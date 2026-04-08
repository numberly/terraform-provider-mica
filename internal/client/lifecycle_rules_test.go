package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func ptrInt64LCR(v int64) *int64 { return &v }

func TestUnit_LifecycleRule_Get(t *testing.T) {
	rules := []client.LifecycleRule{
		{
			ID:                    "lr-id-001",
			Name:                  "my-bucket/expire-old",
			Bucket:                client.NamedReference{Name: "my-bucket"},
			RuleID:                "expire-old",
			Prefix:                "logs/",
			Enabled:               true,
			KeepCurrentVersionFor: ptrInt64LCR(604800000), // 7 days in ms
		},
		{
			ID:     "lr-id-002",
			Name:   "my-bucket/another-rule",
			Bucket: client.NamedReference{Name: "my-bucket"},
			RuleID: "another-rule",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/lifecycle-rules":
			// Get uses bucket_names filter
			bucketName := r.URL.Query().Get("bucket_names")
			if bucketName != "my-bucket" {
				writeJSON(w, http.StatusOK, listResponse([]client.LifecycleRule{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse(rules))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetLifecycleRule(context.Background(), "my-bucket", "expire-old")
	if err != nil {
		t.Fatalf("GetLifecycleRule: %v", err)
	}
	if got.ID != "lr-id-001" {
		t.Errorf("expected ID lr-id-001, got %q", got.ID)
	}
	if got.Prefix != "logs/" {
		t.Errorf("expected Prefix logs/, got %q", got.Prefix)
	}
	if got.RuleID != "expire-old" {
		t.Errorf("expected RuleID expire-old, got %q", got.RuleID)
	}
}

func TestUnit_LifecycleRule_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/lifecycle-rules":
			// Return rules but not the one being searched
			writeJSON(w, http.StatusOK, listResponse([]client.LifecycleRule{
				{ID: "lr-id-999", RuleID: "other-rule", Bucket: client.NamedReference{Name: "my-bucket"}},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetLifecycleRule(context.Background(), "my-bucket", "nonexistent-rule")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_LifecycleRule_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/lifecycle-rules":
			var body client.LifecycleRulePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.LifecycleRule{
				ID:                    "lr-id-003",
				Name:                  body.Bucket.Name + "/" + body.RuleID,
				Bucket:                body.Bucket,
				RuleID:                body.RuleID,
				Prefix:                body.Prefix,
				KeepCurrentVersionFor: body.KeepCurrentVersionFor, // *int64 from POST body
				Enabled:               true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.LifecycleRule{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostLifecycleRule(context.Background(), client.LifecycleRulePost{
		Bucket:                client.NamedReference{Name: "my-bucket"},
		RuleID:                "new-rule",
		Prefix:                "data/",
		KeepCurrentVersionFor: ptrInt64LCR(86400000),
	}, false)
	if err != nil {
		t.Fatalf("PostLifecycleRule: %v", err)
	}
	if got.ID != "lr-id-003" {
		t.Errorf("expected ID lr-id-003, got %q", got.ID)
	}
	if got.RuleID != "new-rule" {
		t.Errorf("expected RuleID new-rule, got %q", got.RuleID)
	}
}

func TestUnit_LifecycleRule_Post_WithConfirmDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/lifecycle-rules":
			if r.URL.Query().Get("confirm_date") != "true" {
				http.Error(w, "expected confirm_date=true", http.StatusBadRequest)
				return
			}
			result := client.LifecycleRule{
				ID:     "lr-id-004",
				RuleID: "date-rule",
				Bucket: client.NamedReference{Name: "my-bucket"},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.LifecycleRule{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostLifecycleRule(context.Background(), client.LifecycleRulePost{
		Bucket: client.NamedReference{Name: "my-bucket"},
		RuleID: "date-rule",
	}, true) // confirmDate = true
	if err != nil {
		t.Fatalf("PostLifecycleRule with confirm_date: %v", err)
	}
	if got.RuleID != "date-rule" {
		t.Errorf("expected RuleID date-rule, got %q", got.RuleID)
	}
}

func TestUnit_LifecycleRule_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/lifecycle-rules":
			// Patch uses composite names: bucketName/ruleID
			name := r.URL.Query().Get("names")
			if name != "my-bucket/expire-old" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			var body client.LifecycleRulePatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			enabled := true
			if body.Enabled != nil {
				enabled = *body.Enabled
			}
			result := client.LifecycleRule{
				ID:      "lr-id-001",
				RuleID:  "expire-old",
				Bucket:  client.NamedReference{Name: "my-bucket"},
				Enabled: enabled,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.LifecycleRule{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	disabled := false
	got, err := c.PatchLifecycleRule(context.Background(), "my-bucket", "expire-old", client.LifecycleRulePatch{
		Enabled: &disabled,
	}, false)
	if err != nil {
		t.Fatalf("PatchLifecycleRule: %v", err)
	}
	// Handler echoes body.Enabled; we sent false so Enabled should be false
	if got.Enabled {
		t.Errorf("expected Enabled false, got true")
	}
}

func TestUnit_LifecycleRule_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/lifecycle-rules":
			// Delete uses composite names: bucketName/ruleID
			name := r.URL.Query().Get("names")
			if name != "my-bucket/expire-old" {
				http.Error(w, "unexpected names param — expected composite bucketName/ruleID", http.StatusBadRequest)
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
	if err := c.DeleteLifecycleRule(context.Background(), "my-bucket", "expire-old"); err != nil {
		t.Fatalf("DeleteLifecycleRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

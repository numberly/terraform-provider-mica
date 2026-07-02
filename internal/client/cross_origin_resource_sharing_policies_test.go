package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

const corsPolicyPath = "/api/2.23/buckets/cross-origin-resource-sharing-policies"
const corsRulesPath = "/api/2.23/buckets/cross-origin-resource-sharing-policies/rules"

func TestUnit_BucketCorsPolicy_GetPolicy_Found(t *testing.T) {
	expected := client.CrossOriginResourceSharingPolicy{
		ID:         "cors-id-001",
		Name:       "my-bucket-cors",
		Bucket:     client.NamedReference{Name: "my-bucket", ID: "bucket-id-001"},
		IsLocal:    true,
		PolicyType: "cross-origin-resource-sharing",
		Rules: []client.CorsRule{
			{
				Name:           "corsrule0",
				AllowedHeaders: []string{"*"},
				AllowedMethods: []string{"GET", "PUT"},
				AllowedOrigins: []string{"https://example.com"},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == corsPolicyPath:
			if r.URL.Query().Get("bucket_names") != "my-bucket" {
				writeJSON(w, http.StatusOK, listResponse([]client.CrossOriginResourceSharingPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.CrossOriginResourceSharingPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetCorsPolicy(context.Background(), "my-bucket")
	if err != nil {
		t.Fatalf("GetCorsPolicy: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if len(got.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got.Rules))
	}
	if got.Rules[0].Name != "corsrule0" {
		t.Errorf("expected rule name corsrule0, got %q", got.Rules[0].Name)
	}
	if len(got.Rules[0].AllowedMethods) != 2 || got.Rules[0].AllowedMethods[0] != "GET" {
		t.Errorf("expected AllowedMethods [GET PUT], got %v", got.Rules[0].AllowedMethods)
	}
}

func TestUnit_BucketCorsPolicy_GetPolicy_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == corsPolicyPath:
			writeJSON(w, http.StatusOK, listResponse([]client.CrossOriginResourceSharingPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetCorsPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_BucketCorsPolicy_EnsurePolicy(t *testing.T) {
	var postCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == corsPolicyPath:
			postCount++
			if r.URL.Query().Get("bucket_names") != "my-bucket" {
				http.Error(w, "bucket_names required", http.StatusBadRequest)
				return
			}
			// Verify the body is empty (empty-body policy create).
			buf := make([]byte, 8)
			n, _ := r.Body.Read(buf)
			if n > 2 { // tolerate "" or "{}"-ish; empty body is expected
				t.Errorf("expected empty POST body, got %q", string(buf[:n]))
			}
			policy := client.CrossOriginResourceSharingPolicy{
				ID:     "cors-id-002",
				Name:   "my-bucket-cors",
				Bucket: client.NamedReference{Name: "my-bucket"},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.CrossOriginResourceSharingPolicy{policy}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.EnsureCorsPolicy(context.Background(), "my-bucket"); err != nil {
		t.Fatalf("EnsureCorsPolicy: %v", err)
	}
	if postCount != 1 {
		t.Errorf("expected 1 POST, got %d", postCount)
	}
}

func TestUnit_BucketCorsPolicy_EnsurePolicy_AlreadyExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == corsPolicyPath:
			// FlashBlade returns HTTP 400 with an "exist" message when the policy is present.
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"errors": []map[string]string{{"message": "CORS policy already exists for bucket"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.EnsureCorsPolicy(context.Background(), "my-bucket"); err != nil {
		t.Fatalf("EnsureCorsPolicy should tolerate already-exists, got: %v", err)
	}
}

func TestUnit_BucketCorsPolicy_PostRule(t *testing.T) {
	var gotBody client.CorsRulePost
	var gotRuleName string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == corsRulesPath:
			gotRuleName = r.URL.Query().Get("names")
			if r.URL.Query().Get("bucket_names") != "my-bucket" {
				http.Error(w, "bucket_names required", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.CorsRule{{Name: gotRuleName}}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.PostCorsRule(context.Background(), "my-bucket", "corsrule0", client.CorsRulePost{
		AllowedMethods: []string{"GET"},
		AllowedOrigins: []string{"*"},
	})
	if err != nil {
		t.Fatalf("PostCorsRule: %v", err)
	}
	if gotRuleName != "corsrule0" {
		t.Errorf("expected rule name corsrule0 in query, got %q", gotRuleName)
	}
	if len(gotBody.AllowedMethods) != 1 || gotBody.AllowedMethods[0] != "GET" {
		t.Errorf("expected body method GET, got %v", gotBody.AllowedMethods)
	}
}

func TestUnit_BucketCorsPolicy_DeleteRule(t *testing.T) {
	var deleteCalled bool
	var gotRuleName string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == corsRulesPath:
			deleteCalled = true
			gotRuleName = r.URL.Query().Get("names")
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteCorsRule(context.Background(), "my-bucket", "corsrule0"); err != nil {
		t.Fatalf("DeleteCorsRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
	if gotRuleName != "corsrule0" {
		t.Errorf("expected rule name corsrule0, got %q", gotRuleName)
	}
}

func TestUnit_BucketCorsPolicy_DeleteRule_MissingTolerated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == corsRulesPath:
			writeJSON(w, http.StatusNotFound, map[string]any{
				"errors": []map[string]string{{"message": "CORS rule does not exist"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteCorsRule(context.Background(), "my-bucket", "ghost"); err != nil {
		t.Fatalf("DeleteCorsRule should tolerate missing rule, got: %v", err)
	}
}

func TestUnit_BucketCorsPolicy_DeletePolicy(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == corsPolicyPath:
			if r.URL.Query().Get("bucket_names") != "my-bucket" {
				http.Error(w, "unexpected bucket_names param", http.StatusBadRequest)
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
	if err := c.DeleteCorsPolicy(context.Background(), "my-bucket"); err != nil {
		t.Fatalf("DeleteCorsPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

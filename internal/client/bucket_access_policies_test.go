package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_BucketAccessPolicy_Get(t *testing.T) {
	expected := client.BucketAccessPolicy{
		ID:         "bap-id-001",
		Name:       "my-bucket",
		Bucket:     client.NamedReference{Name: "my-bucket", ID: "bucket-id-001"},
		Enabled:    true,
		PolicyType: "access",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets/bucket-access-policies":
			bucketName := r.URL.Query().Get("bucket_names")
			if bucketName != "my-bucket" {
				writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetBucketAccessPolicy(context.Background(), "my-bucket")
	if err != nil {
		t.Fatalf("GetBucketAccessPolicy: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if !got.Enabled {
		t.Errorf("expected Enabled true")
	}
}

func TestUnit_BucketAccessPolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets/bucket-access-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetBucketAccessPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_BucketAccessPolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/buckets/bucket-access-policies":
			bucketName := r.URL.Query().Get("bucket_names")
			if bucketName == "" {
				http.Error(w, "bucket_names required", http.StatusBadRequest)
				return
			}
			var body client.BucketAccessPolicyPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.BucketAccessPolicy{
				ID:      "bap-id-002",
				Name:    bucketName,
				Bucket:  client.NamedReference{Name: bucketName},
				Enabled: true,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicy{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostBucketAccessPolicy(context.Background(), "my-bucket", client.BucketAccessPolicyPost{})
	if err != nil {
		t.Fatalf("PostBucketAccessPolicy: %v", err)
	}
	if got.ID != "bap-id-002" {
		t.Errorf("expected ID bap-id-002, got %q", got.ID)
	}
	if got.Bucket.Name != "my-bucket" {
		t.Errorf("expected Bucket.Name my-bucket, got %q", got.Bucket.Name)
	}
}

func TestUnit_BucketAccessPolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/buckets/bucket-access-policies":
			bucketName := r.URL.Query().Get("bucket_names")
			if bucketName != "my-bucket" {
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
	if err := c.DeleteBucketAccessPolicy(context.Background(), "my-bucket"); err != nil {
		t.Fatalf("DeleteBucketAccessPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_BucketAccessPolicyRule_Get(t *testing.T) {
	expected := client.BucketAccessPolicyRule{
		Name:    "allow-public-get",
		Actions: []string{"s3:GetObject"},
		Effect:  "allow",
		Principals: client.BucketAccessPolicyPrincipals{
			All: []string{"*"},
		},
		Resources: []string{"arn:aws:s3:::my-bucket/*"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets/bucket-access-policies/rules":
			bucketName := r.URL.Query().Get("bucket_names")
			ruleName := r.URL.Query().Get("names")
			if bucketName != "my-bucket" || ruleName != "allow-public-get" {
				writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicyRule{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicyRule{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetBucketAccessPolicyRule(context.Background(), "my-bucket", "allow-public-get")
	if err != nil {
		t.Fatalf("GetBucketAccessPolicyRule: %v", err)
	}
	if got.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, got.Name)
	}
	if got.Effect != "allow" {
		t.Errorf("expected Effect allow, got %q", got.Effect)
	}
}

func TestUnit_BucketAccessPolicyRule_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets/bucket-access-policies/rules":
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicyRule{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetBucketAccessPolicyRule(context.Background(), "my-bucket", "nonexistent-rule")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_BucketAccessPolicyRule_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/buckets/bucket-access-policies/rules":
			// POST requires both ?names= (bucket) and ?bucket_names= (bucket)
			names := r.URL.Query().Get("names")
			bucketNames := r.URL.Query().Get("bucket_names")
			if names == "" || bucketNames == "" {
				http.Error(w, "names and bucket_names required", http.StatusBadRequest)
				return
			}
			var body client.BucketAccessPolicyRulePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.BucketAccessPolicyRule{
				Name:       "new-rule",
				Actions:    body.Actions,
				Effect:     "allow",
				Principals: body.Principals,
				Resources:  body.Resources,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAccessPolicyRule{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostBucketAccessPolicyRule(context.Background(), "my-bucket", client.BucketAccessPolicyRulePost{
		Actions: []string{"s3:PutObject"},
		Principals: client.BucketAccessPolicyPrincipals{
			All: []string{"*"},
		},
		Resources: []string{"arn:aws:s3:::my-bucket/*"},
	})
	if err != nil {
		t.Fatalf("PostBucketAccessPolicyRule: %v", err)
	}
	if got.Effect != "allow" {
		t.Errorf("expected Effect allow, got %q", got.Effect)
	}
	if len(got.Actions) != 1 || got.Actions[0] != "s3:PutObject" {
		t.Errorf("expected Actions [s3:PutObject], got %v", got.Actions)
	}
}

func TestUnit_BucketAccessPolicyRule_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/buckets/bucket-access-policies/rules":
			bucketName := r.URL.Query().Get("bucket_names")
			ruleName := r.URL.Query().Get("names")
			if bucketName != "my-bucket" || ruleName != "allow-public-get" {
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
	if err := c.DeleteBucketAccessPolicyRule(context.Background(), "my-bucket", "allow-public-get"); err != nil {
		t.Fatalf("DeleteBucketAccessPolicyRule: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

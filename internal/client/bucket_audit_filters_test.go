package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_BucketAuditFilter_Get(t *testing.T) {
	expected := client.BucketAuditFilter{
		Name:       "audit-filter-01",
		Bucket:     client.NamedReference{Name: "my-bucket", ID: "bucket-id-001"},
		Actions:    []string{"s3:GetObject", "s3:PutObject"},
		S3Prefixes: []string{"logs/", "data/"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets/audit-filters":
			filterName := r.URL.Query().Get("names")
			bucketName := r.URL.Query().Get("bucket_names")
			if filterName != "audit-filter-01" || bucketName != "my-bucket" {
				writeJSON(w, http.StatusOK, listResponse([]client.BucketAuditFilter{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAuditFilter{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetBucketAuditFilter(context.Background(), "audit-filter-01", "my-bucket")
	if err != nil {
		t.Fatalf("GetBucketAuditFilter: %v", err)
	}
	if got.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, got.Name)
	}
	if len(got.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(got.Actions))
	}
	if len(got.S3Prefixes) != 2 {
		t.Errorf("expected 2 S3Prefixes, got %d", len(got.S3Prefixes))
	}
}

func TestUnit_BucketAuditFilter_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/buckets/audit-filters":
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAuditFilter{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetBucketAuditFilter(context.Background(), "nonexistent", "my-bucket")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_BucketAuditFilter_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/buckets/audit-filters":
			filterName := r.URL.Query().Get("names")
			bucketName := r.URL.Query().Get("bucket_names")
			if filterName == "" || bucketName == "" {
				http.Error(w, "names and bucket_names required", http.StatusBadRequest)
				return
			}
			var body client.BucketAuditFilterPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.BucketAuditFilter{
				Name:       filterName,
				Bucket:     client.NamedReference{Name: bucketName},
				Actions:    body.Actions,
				S3Prefixes: body.S3Prefixes,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAuditFilter{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostBucketAuditFilter(
		context.Background(),
		"new-filter",
		"my-bucket",
		client.BucketAuditFilterPost{
			Actions:    []string{"s3:DeleteObject"},
			S3Prefixes: []string{"sensitive/"},
		},
	)
	if err != nil {
		t.Fatalf("PostBucketAuditFilter: %v", err)
	}
	if got.Name != "new-filter" {
		t.Errorf("expected Name new-filter, got %q", got.Name)
	}
	if got.Bucket.Name != "my-bucket" {
		t.Errorf("expected Bucket.Name my-bucket, got %q", got.Bucket.Name)
	}
	if len(got.Actions) != 1 || got.Actions[0] != "s3:DeleteObject" {
		t.Errorf("expected Actions [s3:DeleteObject], got %v", got.Actions)
	}
}

func TestUnit_BucketAuditFilter_Patch(t *testing.T) {
	var gotBody client.BucketAuditFilterPatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/buckets/audit-filters":
			filterName := r.URL.Query().Get("names")
			bucketName := r.URL.Query().Get("bucket_names")
			if filterName != "audit-filter-01" || bucketName != "my-bucket" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			actions := []string{"s3:GetObject"}
			if gotBody.Actions != nil {
				actions = *gotBody.Actions
			}
			result := client.BucketAuditFilter{
				Name:    filterName,
				Bucket:  client.NamedReference{Name: bucketName},
				Actions: actions,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketAuditFilter{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newActions := []string{"s3:GetObject", "s3:HeadObject"}
	got, err := c.PatchBucketAuditFilter(
		context.Background(),
		"audit-filter-01",
		"my-bucket",
		client.BucketAuditFilterPatch{
			Actions: &newActions,
		},
	)
	if err != nil {
		t.Fatalf("PatchBucketAuditFilter: %v", err)
	}
	if len(got.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(got.Actions))
	}
	// PATCH semantics: S3Prefixes should be absent
	if gotBody.S3Prefixes != nil {
		t.Errorf("expected S3Prefixes absent in PATCH body")
	}
}

func TestUnit_BucketAuditFilter_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/buckets/audit-filters":
			filterName := r.URL.Query().Get("names")
			bucketName := r.URL.Query().Get("bucket_names")
			if filterName != "audit-filter-01" || bucketName != "my-bucket" {
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
	if err := c.DeleteBucketAuditFilter(context.Background(), "audit-filter-01", "my-bucket"); err != nil {
		t.Fatalf("DeleteBucketAuditFilter: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

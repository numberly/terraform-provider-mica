package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_LogTargetObjectStore_Get_Found(t *testing.T) {
	expected := client.LogTargetObjectStore{
		ID:   "ltos-id-001",
		Name: "audit-target-1",
		Bucket: client.NamedReference{
			Name: "audit-bucket",
			ID:   "bucket-id-001",
		},
		LogNamePrefix: client.AuditLogNamePrefix{Prefix: "audit/"},
		LogRotate:     client.AuditLogRotate{Duration: 3600000},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/log-targets/object-store":
			name := r.URL.Query().Get("names")
			if name != "audit-target-1" {
				writeJSON(w, http.StatusOK, listResponse([]client.LogTargetObjectStore{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.LogTargetObjectStore{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetLogTargetObjectStore(context.Background(), "audit-target-1")
	if err != nil {
		t.Fatalf("GetLogTargetObjectStore: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Bucket.Name != expected.Bucket.Name {
		t.Errorf("expected Bucket.Name %q, got %q", expected.Bucket.Name, got.Bucket.Name)
	}
	if got.LogNamePrefix.Prefix != expected.LogNamePrefix.Prefix {
		t.Errorf("expected LogNamePrefix.Prefix %q, got %q", expected.LogNamePrefix.Prefix, got.LogNamePrefix.Prefix)
	}
	if got.LogRotate.Duration != expected.LogRotate.Duration {
		t.Errorf("expected LogRotate.Duration %d, got %d", expected.LogRotate.Duration, got.LogRotate.Duration)
	}
}

func TestUnit_LogTargetObjectStore_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/log-targets/object-store":
			writeJSON(w, http.StatusOK, listResponse([]client.LogTargetObjectStore{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetLogTargetObjectStore(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_LogTargetObjectStore_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/log-targets/object-store":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names required", http.StatusBadRequest)
				return
			}
			var body client.LogTargetObjectStorePost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.LogTargetObjectStore{
				ID:            "ltos-id-002",
				Name:          name,
				Bucket:        body.Bucket,
				LogNamePrefix: body.LogNamePrefix,
				LogRotate:     body.LogRotate,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.LogTargetObjectStore{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostLogTargetObjectStore(context.Background(), "new-audit-target", client.LogTargetObjectStorePost{
		Bucket:        client.NamedReference{Name: "audit-bucket"},
		LogNamePrefix: client.AuditLogNamePrefix{Prefix: "logs/"},
		LogRotate:     client.AuditLogRotate{Duration: 7200000},
	})
	if err != nil {
		t.Fatalf("PostLogTargetObjectStore: %v", err)
	}
	if got.Name != "new-audit-target" {
		t.Errorf("expected Name %q, got %q", "new-audit-target", got.Name)
	}
	if got.Bucket.Name != "audit-bucket" {
		t.Errorf("expected Bucket.Name %q, got %q", "audit-bucket", got.Bucket.Name)
	}
	if got.LogNamePrefix.Prefix != "logs/" {
		t.Errorf("expected LogNamePrefix.Prefix %q, got %q", "logs/", got.LogNamePrefix.Prefix)
	}
	if got.LogRotate.Duration != 7200000 {
		t.Errorf("expected LogRotate.Duration 7200000, got %d", got.LogRotate.Duration)
	}
}

func TestUnit_LogTargetObjectStore_Patch_Bucket(t *testing.T) {
	var gotBody client.LogTargetObjectStorePatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/log-targets/object-store":
			name := r.URL.Query().Get("names")
			if name != "audit-target-1" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			bucketName := "old-bucket"
			if gotBody.Bucket != nil {
				bucketName = gotBody.Bucket.Name
			}
			result := client.LogTargetObjectStore{
				ID:     "ltos-id-001",
				Name:   name,
				Bucket: client.NamedReference{Name: bucketName},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.LogTargetObjectStore{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newBucket := client.NamedReference{Name: "new-audit-bucket"}
	got, err := c.PatchLogTargetObjectStore(context.Background(), "audit-target-1", client.LogTargetObjectStorePatch{
		Bucket: &newBucket,
	})
	if err != nil {
		t.Fatalf("PatchLogTargetObjectStore: %v", err)
	}
	if got.Bucket.Name != "new-audit-bucket" {
		t.Errorf("expected Bucket.Name %q, got %q", "new-audit-bucket", got.Bucket.Name)
	}
	// PATCH semantics: LogNamePrefix should not have been sent.
	if gotBody.LogNamePrefix != nil {
		t.Errorf("expected LogNamePrefix absent in PATCH body")
	}
}

func TestUnit_LogTargetObjectStore_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/log-targets/object-store":
			name := r.URL.Query().Get("names")
			if name != "audit-target-1" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
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
	if err := c.DeleteLogTargetObjectStore(context.Background(), "audit-target-1"); err != nil {
		t.Fatalf("DeleteLogTargetObjectStore: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_BucketReplicaLink_Get(t *testing.T) {
	expected := client.BucketReplicaLink{
		ID:           "brl-id-001",
		LocalBucket:  client.NamedReference{Name: "local-bucket", ID: "lb-id-001"},
		RemoteBucket: client.NamedReference{Name: "remote-bucket", ID: "rb-id-001"},
		Remote:       client.NamedReference{Name: "remote-array", ID: "ra-id-001"},
		Paused:       false,
		Direction:    "from-primary",
		Status:       "healthy",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/bucket-replica-links":
			localBucket := r.URL.Query().Get("local_bucket_names")
			remoteBucket := r.URL.Query().Get("remote_bucket_names")
			if localBucket != "local-bucket" || remoteBucket != "remote-bucket" {
				writeJSON(w, http.StatusOK, listResponse([]client.BucketReplicaLink{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketReplicaLink{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetBucketReplicaLink(context.Background(), "local-bucket", "remote-bucket")
	if err != nil {
		t.Fatalf("GetBucketReplicaLink: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Status != "healthy" {
		t.Errorf("expected Status healthy, got %q", got.Status)
	}
}

func TestUnit_BucketReplicaLink_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/bucket-replica-links":
			writeJSON(w, http.StatusOK, listResponse([]client.BucketReplicaLink{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetBucketReplicaLink(context.Background(), "nonexistent-local", "nonexistent-remote")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_BucketReplicaLink_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/bucket-replica-links":
			localBucket := r.URL.Query().Get("local_bucket_names")
			remoteBucket := r.URL.Query().Get("remote_bucket_names")
			remoteCreds := r.URL.Query().Get("remote_credentials_names")
			if localBucket == "" || remoteBucket == "" {
				http.Error(w, "local_bucket_names and remote_bucket_names required", http.StatusBadRequest)
				return
			}
			var body client.BucketReplicaLinkPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.BucketReplicaLink{
				ID:           "brl-id-002",
				LocalBucket:  client.NamedReference{Name: localBucket},
				RemoteBucket: client.NamedReference{Name: remoteBucket},
				Paused:       body.Paused,
			}
			if remoteCreds != "" {
				cred := client.NamedReference{Name: remoteCreds}
				result.RemoteCredentials = &cred
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketReplicaLink{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostBucketReplicaLink(
		context.Background(),
		"local-bucket",
		"remote-bucket",
		"remote-array/creds-001",
		client.BucketReplicaLinkPost{Paused: false},
	)
	if err != nil {
		t.Fatalf("PostBucketReplicaLink: %v", err)
	}
	if got.ID != "brl-id-002" {
		t.Errorf("expected ID brl-id-002, got %q", got.ID)
	}
	if got.LocalBucket.Name != "local-bucket" {
		t.Errorf("expected LocalBucket.Name local-bucket, got %q", got.LocalBucket.Name)
	}
	if got.RemoteCredentials == nil || got.RemoteCredentials.Name != "remote-array/creds-001" {
		t.Errorf("expected RemoteCredentials.Name remote-array/creds-001")
	}
}

func TestUnit_BucketReplicaLink_Post_NoCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/bucket-replica-links":
			remoteCreds := r.URL.Query().Get("remote_credentials_names")
			if remoteCreds != "" {
				http.Error(w, "unexpected remote_credentials_names", http.StatusBadRequest)
				return
			}
			result := client.BucketReplicaLink{
				ID:           "brl-id-003",
				LocalBucket:  client.NamedReference{Name: "local-bucket"},
				RemoteBucket: client.NamedReference{Name: "remote-bucket"},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketReplicaLink{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	// Empty string for remoteCredentialsName should omit the query param (FB-to-FB case)
	got, err := c.PostBucketReplicaLink(
		context.Background(),
		"local-bucket",
		"remote-bucket",
		"",
		client.BucketReplicaLinkPost{},
	)
	if err != nil {
		t.Fatalf("PostBucketReplicaLink without creds: %v", err)
	}
	if got.ID != "brl-id-003" {
		t.Errorf("expected ID brl-id-003, got %q", got.ID)
	}
}

func TestUnit_BucketReplicaLink_Patch(t *testing.T) {
	// DELETE uses ?ids= not ?names=
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/bucket-replica-links":
			id := r.URL.Query().Get("ids")
			if id != "brl-id-001" {
				http.Error(w, "unexpected ids param", http.StatusBadRequest)
				return
			}
			var body client.BucketReplicaLinkPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			paused := false
			if body.Paused != nil {
				paused = *body.Paused
			}
			result := client.BucketReplicaLink{
				ID:     "brl-id-001",
				Paused: paused,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.BucketReplicaLink{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	paused := true
	got, err := c.PatchBucketReplicaLink(context.Background(), "brl-id-001", client.BucketReplicaLinkPatch{
		Paused: &paused,
	})
	if err != nil {
		t.Fatalf("PatchBucketReplicaLink: %v", err)
	}
	// Server echoes back the value we set (we set paused=true but handler echoes body.Paused)
	_ = got
}

func TestUnit_BucketReplicaLink_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/bucket-replica-links":
			// DELETE uses ?ids= not ?names=
			id := r.URL.Query().Get("ids")
			if id != "brl-id-001" {
				http.Error(w, "unexpected ids param — DELETE must use ?ids=", http.StatusBadRequest)
				return
			}
			// Verify that ?names= is NOT used
			if r.URL.Query().Get("names") != "" {
				http.Error(w, "DELETE must not use ?names=", http.StatusBadRequest)
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
	if err := c.DeleteBucketReplicaLink(context.Background(), "brl-id-001"); err != nil {
		t.Fatalf("DeleteBucketReplicaLink: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

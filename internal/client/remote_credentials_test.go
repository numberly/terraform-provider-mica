package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_RemoteCredentials_Get(t *testing.T) {
	expected := client.ObjectStoreRemoteCredentials{
		ID:          "rc-id-001",
		Name:        "remote-array/creds-001",
		AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
		Remote:      client.NamedReference{Name: "remote-array", ID: "remote-id-001"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-remote-credentials":
			name := r.URL.Query().Get("names")
			if name != "remote-array/creds-001" {
				writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreRemoteCredentials{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreRemoteCredentials{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetRemoteCredentials(context.Background(), "remote-array/creds-001")
	if err != nil {
		t.Fatalf("GetRemoteCredentials: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.AccessKeyID != expected.AccessKeyID {
		t.Errorf("expected AccessKeyID %q, got %q", expected.AccessKeyID, got.AccessKeyID)
	}
}

func TestUnit_RemoteCredentials_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/object-store-remote-credentials":
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreRemoteCredentials{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetRemoteCredentials(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_RemoteCredentials_Post_WithRemote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-remote-credentials":
			name := r.URL.Query().Get("names")
			remoteName := r.URL.Query().Get("remote_names")
			targetName := r.URL.Query().Get("target_names")
			if name == "" || (remoteName == "" && targetName == "") {
				http.Error(w, "names and (remote_names or target_names) required", http.StatusBadRequest)
				return
			}
			if remoteName != "" && targetName != "" {
				http.Error(w, "provide remote_names or target_names, not both", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreRemoteCredentialsPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			refName := remoteName
			if targetName != "" {
				refName = targetName
			}
			result := client.ObjectStoreRemoteCredentials{
				ID:          "rc-id-002",
				Name:        name,
				AccessKeyID: body.AccessKeyID,
				Remote:      client.NamedReference{Name: refName},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreRemoteCredentials{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostRemoteCredentials(
		context.Background(),
		"remote-array/new-creds",
		"remote-array",
		"",
		client.ObjectStoreRemoteCredentialsPost{
			AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	)
	if err != nil {
		t.Fatalf("PostRemoteCredentials: %v", err)
	}
	if got.Name != "remote-array/new-creds" {
		t.Errorf("expected Name remote-array/new-creds, got %q", got.Name)
	}
	if got.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("expected AccessKeyID AKIAIOSFODNN7EXAMPLE, got %q", got.AccessKeyID)
	}
	if got.Remote.Name != "remote-array" {
		t.Errorf("expected Remote.Name remote-array, got %q", got.Remote.Name)
	}
}

func TestUnit_RemoteCredentials_Post_WithTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-remote-credentials":
			name := r.URL.Query().Get("names")
			remoteName := r.URL.Query().Get("remote_names")
			targetName := r.URL.Query().Get("target_names")
			if remoteName != "" {
				http.Error(w, "expected target_names not remote_names", http.StatusBadRequest)
				return
			}
			if targetName == "" {
				http.Error(w, "target_names required", http.StatusBadRequest)
				return
			}
			var body client.ObjectStoreRemoteCredentialsPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.ObjectStoreRemoteCredentials{
				ID:          "rc-id-003",
				Name:        name,
				AccessKeyID: body.AccessKeyID,
				Remote:      client.NamedReference{Name: targetName},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreRemoteCredentials{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostRemoteCredentials(
		context.Background(),
		"target-creds",
		"",
		"my-target",
		client.ObjectStoreRemoteCredentialsPost{
			AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	)
	if err != nil {
		t.Fatalf("PostRemoteCredentials with target: %v", err)
	}
	if got.Name != "target-creds" {
		t.Errorf("expected Name target-creds, got %q", got.Name)
	}
	if got.Remote.Name != "my-target" {
		t.Errorf("expected Remote.Name my-target, got %q", got.Remote.Name)
	}
}

func TestUnit_RemoteCredentials_Post_NeitherParam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/object-store-remote-credentials":
			remoteName := r.URL.Query().Get("remote_names")
			targetName := r.URL.Query().Get("target_names")
			if remoteName == "" && targetName == "" {
				http.Error(w, "remote_names or target_names is required", http.StatusBadRequest)
				return
			}
			// Shouldn't reach here in this test.
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	// Pass empty strings for both remoteName and targetName — server should reject.
	_, err := c.PostRemoteCredentials(
		context.Background(),
		"creds-name",
		"",
		"",
		client.ObjectStoreRemoteCredentialsPost{
			AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	)
	if err == nil {
		t.Fatal("expected error when neither remote_names nor target_names provided, got nil")
	}
}

func TestUnit_RemoteCredentials_Patch(t *testing.T) {
	var gotBody client.ObjectStoreRemoteCredentialsPatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/object-store-remote-credentials":
			name := r.URL.Query().Get("names")
			if name != "remote-array/creds-001" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newKeyID := "NEWKEYID"
			if gotBody.AccessKeyID != nil {
				newKeyID = *gotBody.AccessKeyID
			}
			result := client.ObjectStoreRemoteCredentials{
				ID:          "rc-id-001",
				Name:        name,
				AccessKeyID: newKeyID,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ObjectStoreRemoteCredentials{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newKey := "UPDATEDKEYID"
	got, err := c.PatchRemoteCredentials(
		context.Background(),
		"remote-array/creds-001",
		client.ObjectStoreRemoteCredentialsPatch{
			AccessKeyID: &newKey,
		},
	)
	if err != nil {
		t.Fatalf("PatchRemoteCredentials: %v", err)
	}
	if got.AccessKeyID != "UPDATEDKEYID" {
		t.Errorf("expected AccessKeyID UPDATEDKEYID, got %q", got.AccessKeyID)
	}
	// PATCH semantics: SecretAccessKey should be absent (nil)
	if gotBody.SecretAccessKey != nil {
		t.Errorf("expected SecretAccessKey absent in PATCH body")
	}
}

func TestUnit_RemoteCredentials_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/object-store-remote-credentials":
			name := r.URL.Query().Get("names")
			if name != "remote-array/creds-001" {
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
	if err := c.DeleteRemoteCredentials(context.Background(), "remote-array/creds-001"); err != nil {
		t.Fatalf("DeleteRemoteCredentials: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

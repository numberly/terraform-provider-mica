package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_Server_Get(t *testing.T) {
	expected := client.Server{
		ID:   "srv-id-001",
		Name: "test-server",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/servers":
			name := r.URL.Query().Get("names")
			if name != "test-server" {
				writeJSON(w, http.StatusOK, listResponse([]client.Server{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.Server{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetServer(context.Background(), "test-server")
	if err != nil {
		t.Fatalf("GetServer: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, got.Name)
	}
}

func TestUnit_Server_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/servers":
			writeJSON(w, http.StatusOK, listResponse([]client.Server{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetServer(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_Server_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/servers":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			var body client.ServerPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			s := client.Server{
				ID:   "srv-id-002",
				Name: name,
				DNS:  body.DNS,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.Server{s}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostServer(context.Background(), "new-server", client.ServerPost{
		DNS: []client.NamedReference{{Name: "management"}},
	})
	if err != nil {
		t.Fatalf("PostServer: %v", err)
	}
	if got.ID != "srv-id-002" {
		t.Errorf("expected ID srv-id-002, got %q", got.ID)
	}
	if got.Name != "new-server" {
		t.Errorf("expected Name new-server, got %q", got.Name)
	}
	if len(got.DNS) != 1 || got.DNS[0].Name != "management" {
		t.Errorf("expected DNS name management, got %v", got.DNS)
	}
}

func TestUnit_Server_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/servers":
			name := r.URL.Query().Get("names")
			if name != "existing-server" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			var body client.ServerPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			s := client.Server{
				ID:   "srv-id-003",
				Name: name,
				DNS:  body.DNS,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.Server{s}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PatchServer(context.Background(), "existing-server", client.ServerPatch{
		DNS: []client.NamedReference{{Name: "updated-dns"}},
	})
	if err != nil {
		t.Fatalf("PatchServer: %v", err)
	}
	if got.Name != "existing-server" {
		t.Errorf("expected Name existing-server, got %q", got.Name)
	}
	if len(got.DNS) != 1 || got.DNS[0].Name != "updated-dns" {
		t.Errorf("expected updated DNS name, got %v", got.DNS)
	}
}

func TestUnit_Server_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/servers":
			name := r.URL.Query().Get("names")
			if name != "delete-server" {
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
	if err := c.DeleteServer(context.Background(), "delete-server", nil); err != nil {
		t.Fatalf("DeleteServer: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_Server_Delete_CascadeDelete(t *testing.T) {
	var gotCascade string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/servers":
			gotCascade = r.URL.Query().Get("cascade_delete")
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteServer(context.Background(), "cascade-server", []string{"export1", "export2"}); err != nil {
		t.Fatalf("DeleteServer with cascade: %v", err)
	}
	if gotCascade != "export1,export2" {
		t.Errorf("expected cascade_delete=export1,export2, got %q", gotCascade)
	}
}

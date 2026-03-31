package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_SyslogServer_Get(t *testing.T) {
	expected := client.SyslogServer{
		ID:       "syslog-id-001",
		Name:     "syslog-1",
		URI:      "udp://syslog.example.com:514",
		Services: []string{"array_admin"},
		Sources:  []string{"10.0.0.0/24"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/syslog-servers":
			name := r.URL.Query().Get("names")
			if name != "syslog-1" {
				writeJSON(w, http.StatusOK, listResponse([]client.SyslogServer{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SyslogServer{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetSyslogServer(context.Background(), "syslog-1")
	if err != nil {
		t.Fatalf("GetSyslogServer: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.URI != expected.URI {
		t.Errorf("expected URI %q, got %q", expected.URI, got.URI)
	}
	if len(got.Services) != 1 || got.Services[0] != "array_admin" {
		t.Errorf("expected Services [array_admin], got %v", got.Services)
	}
}

func TestUnit_SyslogServer_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/syslog-servers":
			writeJSON(w, http.StatusOK, listResponse([]client.SyslogServer{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetSyslogServer(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_SyslogServer_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/syslog-servers":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param missing", http.StatusBadRequest)
				return
			}
			var body client.SyslogServerPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			s := client.SyslogServer{
				ID:       "syslog-id-002",
				Name:     name,
				URI:      body.URI,
				Services: body.Services,
				Sources:  body.Sources,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SyslogServer{s}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostSyslogServer(context.Background(), "new-syslog", client.SyslogServerPost{
		URI:      "tcp://syslog.example.com:514",
		Services: []string{"data_reduction"},
		Sources:  []string{"192.168.1.0/24"},
	})
	if err != nil {
		t.Fatalf("PostSyslogServer: %v", err)
	}
	if got.ID != "syslog-id-002" {
		t.Errorf("expected ID syslog-id-002, got %q", got.ID)
	}
	if got.Name != "new-syslog" {
		t.Errorf("expected Name new-syslog, got %q", got.Name)
	}
	if got.URI != "tcp://syslog.example.com:514" {
		t.Errorf("expected URI tcp://syslog.example.com:514, got %q", got.URI)
	}
	if len(got.Services) != 1 || got.Services[0] != "data_reduction" {
		t.Errorf("expected Services [data_reduction], got %v", got.Services)
	}
}

func TestUnit_SyslogServer_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/syslog-servers":
			name := r.URL.Query().Get("names")
			if name != "existing-syslog" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			var body client.SyslogServerPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			uri := "udp://old.example.com:514"
			if body.URI != nil {
				uri = *body.URI
			}
			s := client.SyslogServer{
				ID:   "syslog-id-003",
				Name: name,
				URI:  uri,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SyslogServer{s}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newURI := "tcp://new-syslog.example.com:514"
	got, err := c.PatchSyslogServer(context.Background(), "existing-syslog", client.SyslogServerPatch{
		URI: &newURI,
	})
	if err != nil {
		t.Fatalf("PatchSyslogServer: %v", err)
	}
	if got.URI != "tcp://new-syslog.example.com:514" {
		t.Errorf("expected URI tcp://new-syslog.example.com:514, got %q", got.URI)
	}
}

func TestUnit_SyslogServer_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/syslog-servers":
			name := r.URL.Query().Get("names")
			if name != "delete-syslog" {
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
	if err := c.DeleteSyslogServer(context.Background(), "delete-syslog"); err != nil {
		t.Fatalf("DeleteSyslogServer: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_SyslogServer_List(t *testing.T) {
	servers := []client.SyslogServer{
		{ID: "syslog-id-010", Name: "syslog-list-1", URI: "udp://s1.example.com:514"},
		{ID: "syslog-id-011", Name: "syslog-list-2", URI: "tcp://s2.example.com:514"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/syslog-servers":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            servers,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	items, err := c.ListSyslogServers(context.Background(), client.ListSyslogServersOpts{})
	if err != nil {
		t.Fatalf("ListSyslogServers: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "syslog-list-1" {
		t.Errorf("expected first item name syslog-list-1, got %q", items[0].Name)
	}
}

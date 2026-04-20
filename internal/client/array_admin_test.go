package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

func TestUnit_ArrayDns_Get_Found(t *testing.T) {
	expected := client.ArrayDns{
		ID:          "dns-id-001",
		Name:        "dns-config",
		Domain:      "example.com",
		Nameservers: []string{"1.1.1.1", "8.8.8.8"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/dns":
			name := r.URL.Query().Get("names")
			if name != "dns-config" {
				writeJSON(w, http.StatusOK, listResponse([]client.ArrayDns{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ArrayDns{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetArrayDns(context.Background(), "dns-config")
	if err != nil {
		t.Fatalf("GetArrayDns: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, got.Name)
	}
	if got.Domain != expected.Domain {
		t.Errorf("expected Domain %q, got %q", expected.Domain, got.Domain)
	}
	if len(got.Nameservers) != 2 {
		t.Errorf("expected 2 nameservers, got %d", len(got.Nameservers))
	}
}

func TestUnit_ArrayDns_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/dns":
			writeJSON(w, http.StatusOK, listResponse([]client.ArrayDns{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetArrayDns(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound error, got %v", err)
	}
}

func TestUnit_ArrayDns_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/dns":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names required", http.StatusBadRequest)
				return
			}
			var body client.ArrayDnsPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.ArrayDns{
				ID:          "dns-new-001",
				Name:        name,
				Domain:      body.Domain,
				Nameservers: body.Nameservers,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ArrayDns{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostArrayDns(context.Background(), "my-dns", client.ArrayDnsPost{
		Domain:      "test.example.com",
		Nameservers: []string{"9.9.9.9"},
	})
	if err != nil {
		t.Fatalf("PostArrayDns: %v", err)
	}
	if got.Name != "my-dns" {
		t.Errorf("expected Name %q, got %q", "my-dns", got.Name)
	}
	if got.Domain != "test.example.com" {
		t.Errorf("expected Domain %q, got %q", "test.example.com", got.Domain)
	}
}

func TestUnit_ArrayDns_Patch(t *testing.T) {
	var gotBody client.ArrayDnsPatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/dns":
			name := r.URL.Query().Get("names")
			if name != "dns-config" {
				http.Error(w, "unexpected name", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			domain := "updated.example.com"
			result := client.ArrayDns{
				ID:     "dns-id-001",
				Name:   name,
				Domain: domain,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ArrayDns{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newDomain := "updated.example.com"
	got, err := c.PatchArrayDns(context.Background(), "dns-config", client.ArrayDnsPatch{
		Domain: &newDomain,
	})
	if err != nil {
		t.Fatalf("PatchArrayDns: %v", err)
	}
	if got.Domain != newDomain {
		t.Errorf("expected Domain %q, got %q", newDomain, got.Domain)
	}
	// Verify PATCH semantics: Nameservers should be absent
	if gotBody.Nameservers != nil {
		t.Errorf("expected Nameservers absent in PATCH body")
	}
}

func TestUnit_ArrayDns_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/dns":
			name := r.URL.Query().Get("names")
			if name != "dns-config" {
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
	if err := c.DeleteArrayDns(context.Background(), "dns-config"); err != nil {
		t.Fatalf("DeleteArrayDns: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_ArrayInfo_Get(t *testing.T) {
	expected := client.ArrayInfo{
		ID:         "array-id-001",
		Name:       "pure01",
		NtpServers: []string{"pool.ntp.org", "time.cloudflare.com"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/arrays":
			writeJSON(w, http.StatusOK, listResponse([]client.ArrayInfo{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetArrayNtp(context.Background())
	if err != nil {
		t.Fatalf("GetArrayNtp: %v", err)
	}
	if got.Name != expected.Name {
		t.Errorf("expected Name %q, got %q", expected.Name, got.Name)
	}
	if len(got.NtpServers) != 2 {
		t.Errorf("expected 2 NTP servers, got %d", len(got.NtpServers))
	}
}

func TestUnit_ArrayNtp_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/arrays":
			var body client.ArrayNtpPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			servers := []string{"0.pool.ntp.org"}
			result := client.ArrayInfo{
				ID:         "array-id-001",
				NtpServers: servers,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ArrayInfo{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newServers := []string{"0.pool.ntp.org"}
	got, err := c.PatchArrayNtp(context.Background(), client.ArrayNtpPatch{
		NtpServers: &newServers,
	})
	if err != nil {
		t.Fatalf("PatchArrayNtp: %v", err)
	}
	if len(got.NtpServers) != 1 || got.NtpServers[0] != "0.pool.ntp.org" {
		t.Errorf("expected NtpServers [0.pool.ntp.org], got %v", got.NtpServers)
	}
}

func TestUnit_SmtpServer_Get(t *testing.T) {
	expected := client.SmtpServer{
		ID:           "smtp-id-001",
		Name:         "pure01",
		RelayHost:    "smtp.example.com:25",
		SenderDomain: "example.com",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/smtp-servers":
			writeJSON(w, http.StatusOK, listResponse([]client.SmtpServer{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetSmtpServer(context.Background())
	if err != nil {
		t.Fatalf("GetSmtpServer: %v", err)
	}
	if got.RelayHost != expected.RelayHost {
		t.Errorf("expected RelayHost %q, got %q", expected.RelayHost, got.RelayHost)
	}
	if got.SenderDomain != expected.SenderDomain {
		t.Errorf("expected SenderDomain %q, got %q", expected.SenderDomain, got.SenderDomain)
	}
}

func TestUnit_SmtpServer_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/smtp-servers":
			var body client.SmtpServerPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			result := client.SmtpServer{
				ID:        "smtp-id-001",
				RelayHost: "newsmtp.example.com:587",
			}
			writeJSON(w, http.StatusOK, listResponse([]client.SmtpServer{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newRelay := "newsmtp.example.com:587"
	got, err := c.PatchSmtpServer(context.Background(), client.SmtpServerPatch{
		RelayHost: &newRelay,
	})
	if err != nil {
		t.Fatalf("PatchSmtpServer: %v", err)
	}
	if got.RelayHost != newRelay {
		t.Errorf("expected RelayHost %q, got %q", newRelay, got.RelayHost)
	}
}

func TestUnit_AlertWatcher_Get(t *testing.T) {
	watchers := []client.AlertWatcher{
		{ID: "aw-id-001", Name: "ops@example.com", Enabled: true, MinimumNotificationSeverity: "warning"},
		{ID: "aw-id-002", Name: "dev@example.com", Enabled: false},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/alert-watchers":
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            watchers,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.ListAlertWatchers(context.Background())
	if err != nil {
		t.Fatalf("ListAlertWatchers: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 watchers, got %d", len(got))
	}
	if got[0].Name != "ops@example.com" {
		t.Errorf("expected Name ops@example.com, got %q", got[0].Name)
	}
}

func TestUnit_AlertWatcher_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/alert-watchers":
			email := r.URL.Query().Get("names")
			if email == "" {
				http.Error(w, "names query param required", http.StatusBadRequest)
				return
			}
			result := client.AlertWatcher{
				ID:                          "aw-id-003",
				Name:                        email,
				Enabled:                     true,
				MinimumNotificationSeverity: "info",
			}
			writeJSON(w, http.StatusOK, listResponse([]client.AlertWatcher{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostAlertWatcher(context.Background(), "newops@example.com", client.AlertWatcherPost{
		MinimumNotificationSeverity: "info",
	})
	if err != nil {
		t.Fatalf("PostAlertWatcher: %v", err)
	}
	if got.Name != "newops@example.com" {
		t.Errorf("expected Name newops@example.com, got %q", got.Name)
	}
	if got.ID != "aw-id-003" {
		t.Errorf("expected ID aw-id-003, got %q", got.ID)
	}
}

func TestUnit_AlertWatcher_Patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/alert-watchers":
			email := r.URL.Query().Get("names")
			if email != "ops@example.com" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			var body client.AlertWatcherPatch
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			enabled := false
			if body.Enabled != nil {
				enabled = *body.Enabled
			}
			result := client.AlertWatcher{
				ID:      "aw-id-001",
				Name:    email,
				Enabled: enabled,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.AlertWatcher{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	disabled := false
	got, err := c.PatchAlertWatcher(context.Background(), "ops@example.com", client.AlertWatcherPatch{
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("PatchAlertWatcher: %v", err)
	}
	if got.Enabled {
		t.Errorf("expected Enabled false, got true")
	}
}

func TestUnit_AlertWatcher_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/alert-watchers":
			email := r.URL.Query().Get("names")
			if email != "ops@example.com" {
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
	if err := c.DeleteAlertWatcher(context.Background(), "ops@example.com"); err != nil {
		t.Fatalf("DeleteAlertWatcher: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_ArrayConnection_Get(t *testing.T) {
	expected := client.ArrayConnection{
		ID:                "conn-id-001",
		Status:            "connected",
		Remote:            client.NamedReference{Name: "remote-array", ID: "remote-id-001"},
		ManagementAddress: "10.0.0.1",
		Encrypted:         true,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/array-connections":
			remoteName := r.URL.Query().Get("remote_names")
			if remoteName != "remote-array" {
				writeJSON(w, http.StatusOK, listResponse([]client.ArrayConnection{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.ArrayConnection{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetArrayConnection(context.Background(), "remote-array")
	if err != nil {
		t.Fatalf("GetArrayConnection: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.Remote.Name != "remote-array" {
		t.Errorf("expected Remote.Name remote-array, got %q", got.Remote.Name)
	}
	if !got.Encrypted {
		t.Errorf("expected Encrypted true")
	}
}

// TestUnit_ArrayConnection_Get_NotFound moved to array_connections_test.go (uses full mock handler).

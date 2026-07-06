package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// snmpManagerStoreFacade wraps the opaque store so tests can call Seed.
type snmpManagerStoreFacade struct {
	store interface {
		Seed(m *client.SnmpManager)
	}
}

func newSnmpManagerServer(t *testing.T) (*httptest.Server, *snmpManagerStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterSnmpManagerHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &snmpManagerStoreFacade{store: store}
}

func TestUnit_SnmpManager_Get_Found(t *testing.T) {
	srv, facade := newSnmpManagerServer(t)
	facade.store.Seed(&client.SnmpManager{
		ID:           "snmpmgr-seed-1",
		Name:         "mgr1",
		Host:         "snmp.example.com",
		Notification: "trap",
		Version:      "v3",
		V3: &client.SnmpV3{
			User:              "purity_user",
			AuthProtocol:      "SHA",
			AuthPassphrase:    "auth-secret",
			PrivacyProtocol:   "AES",
			PrivacyPassphrase: "priv-secret-min8",
		},
	})

	c := newTestClient(t, srv)
	got, err := c.GetSnmpManager(context.Background(), "mgr1")
	if err != nil {
		t.Fatalf("GetSnmpManager: %v", err)
	}
	if got.Name != "mgr1" {
		t.Errorf("expected Name %q, got %q", "mgr1", got.Name)
	}
	if got.Host != "snmp.example.com" {
		t.Errorf("expected Host %q, got %q", "snmp.example.com", got.Host)
	}
	if got.Version != "v3" {
		t.Errorf("expected Version %q, got %q", "v3", got.Version)
	}
	if got.V3 == nil {
		t.Fatal("expected V3 set, got nil")
	}
	if got.V3.User != "purity_user" {
		t.Errorf("expected V3.User %q, got %q", "purity_user", got.V3.User)
	}
	// Sensitive fields must be stripped on GET (mirrors real API).
	if got.V3.AuthPassphrase != "" {
		t.Errorf("expected V3.AuthPassphrase to be empty after GET, got %q", got.V3.AuthPassphrase)
	}
	if got.V3.PrivacyPassphrase != "" {
		t.Errorf("expected V3.PrivacyPassphrase to be empty after GET, got %q", got.V3.PrivacyPassphrase)
	}
}

func TestUnit_SnmpManager_Get_NotFound(t *testing.T) {
	srv, _ := newSnmpManagerServer(t)
	c := newTestClient(t, srv)

	_, err := c.GetSnmpManager(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for unknown manager, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_SnmpManager_Post(t *testing.T) {
	srv, _ := newSnmpManagerServer(t)
	c := newTestClient(t, srv)

	got, err := c.PostSnmpManager(context.Background(), "mgr1", client.SnmpManagerPost{
		Host:         "snmp.example.com",
		Notification: "trap",
		Version:      "v3",
		V3: &client.SnmpV3Post{
			User:              "purity_user",
			AuthProtocol:      "SHA",
			AuthPassphrase:    "auth-secret",
			PrivacyProtocol:   "AES",
			PrivacyPassphrase: "priv-secret-min8",
		},
	})
	if err != nil {
		t.Fatalf("PostSnmpManager: %v", err)
	}
	if got.Name != "mgr1" {
		t.Errorf("expected Name %q, got %q", "mgr1", got.Name)
	}
	if got.Host != "snmp.example.com" {
		t.Errorf("expected Host %q, got %q", "snmp.example.com", got.Host)
	}
	if got.Version != "v3" {
		t.Errorf("expected Version %q, got %q", "v3", got.Version)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after POST")
	}
	if got.V3 == nil {
		t.Fatal("expected V3 set, got nil")
	}
	if got.V3.User != "purity_user" {
		t.Errorf("expected V3.User %q, got %q", "purity_user", got.V3.User)
	}
}

func TestUnit_SnmpManager_Patch(t *testing.T) {
	srv, facade := newSnmpManagerServer(t)
	facade.store.Seed(&client.SnmpManager{
		ID:           "snmpmgr-patch-1",
		Name:         "mgr1",
		Host:         "old.example.com",
		Notification: "trap",
		Version:      "v3",
	})

	c := newTestClient(t, srv)
	newHost := "new.example.com"
	got, err := c.PatchSnmpManager(context.Background(), "mgr1", client.SnmpManagerPatch{
		Host: &newHost,
	})
	if err != nil {
		t.Fatalf("PatchSnmpManager host: %v", err)
	}
	if got.Host != "new.example.com" {
		t.Errorf("expected Host %q, got %q", "new.example.com", got.Host)
	}
	if got.Notification != "trap" {
		t.Errorf("expected Notification %q (unchanged), got %q", "trap", got.Notification)
	}
}

func TestUnit_SnmpManager_Delete(t *testing.T) {
	srv, facade := newSnmpManagerServer(t)
	facade.store.Seed(&client.SnmpManager{
		ID:      "snmpmgr-del-1",
		Name:    "mgr1",
		Host:    "snmp.example.com",
		Version: "v2c",
	})

	c := newTestClient(t, srv)
	if err := c.DeleteSnmpManager(context.Background(), "mgr1"); err != nil {
		t.Fatalf("DeleteSnmpManager: %v", err)
	}

	_, err := c.GetSnmpManager(context.Background(), "mgr1")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true after delete, got false; err: %v", err)
	}
}

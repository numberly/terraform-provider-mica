package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

func newArrayConnectionServer(t *testing.T) (*httptest.Server, *arrayConnectionStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterArrayConnectionHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &arrayConnectionStoreFacade{store: store}
}

// arrayConnectionStoreFacade wraps the opaque store so tests can call Seed.
type arrayConnectionStoreFacade struct {
	store interface {
		Seed(conn *client.ArrayConnection)
	}
}

func TestUnit_ArrayConnection_Get_Found(t *testing.T) {
	srv, facade := newArrayConnectionServer(t)
	facade.store.Seed(&client.ArrayConnection{
		ID:                "conn-1",
		Remote:            client.NamedReference{Name: "remote-fb", ID: "remote-id-1"},
		ManagementAddress: "10.0.0.1",
		Encrypted:         true,
		Status:            "connected",
		Type:              "async-replication",
		Version:           "4.3.0",
		ReplicationAddresses: []string{"10.0.1.1"},
	})

	c := newTestClient(t, srv)
	got, err := c.GetArrayConnection(context.Background(), "remote-fb")
	if err != nil {
		t.Fatalf("GetArrayConnection: %v", err)
	}
	if got.ID != "conn-1" {
		t.Errorf("expected ID %q, got %q", "conn-1", got.ID)
	}
	if got.Remote.Name != "remote-fb" {
		t.Errorf("expected Remote.Name %q, got %q", "remote-fb", got.Remote.Name)
	}
	if got.Remote.ID != "remote-id-1" {
		t.Errorf("expected Remote.ID %q, got %q", "remote-id-1", got.Remote.ID)
	}
	if got.ManagementAddress != "10.0.0.1" {
		t.Errorf("expected ManagementAddress %q, got %q", "10.0.0.1", got.ManagementAddress)
	}
	if !got.Encrypted {
		t.Error("expected Encrypted=true")
	}
	if got.Status != "connected" {
		t.Errorf("expected Status %q, got %q", "connected", got.Status)
	}
}

func TestUnit_ArrayConnection_Get_NotFound(t *testing.T) {
	srv, _ := newArrayConnectionServer(t)
	c := newTestClient(t, srv)

	_, err := c.GetArrayConnection(context.Background(), "nonexistent-remote")
	if err == nil {
		t.Fatal("expected error for unknown array connection, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_ArrayConnection_Post(t *testing.T) {
	srv, _ := newArrayConnectionServer(t)
	c := newTestClient(t, srv)

	got, err := c.PostArrayConnection(context.Background(), "new-remote", client.ArrayConnectionPost{
		ManagementAddress: "192.168.1.100",
		ConnectionKey:     "secret-key-123",
		Encrypted:         true,
	})
	if err != nil {
		t.Fatalf("PostArrayConnection: %v", err)
	}
	if got.Remote.Name != "new-remote" {
		t.Errorf("expected Remote.Name %q, got %q", "new-remote", got.Remote.Name)
	}
	if got.ManagementAddress != "192.168.1.100" {
		t.Errorf("expected ManagementAddress %q, got %q", "192.168.1.100", got.ManagementAddress)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after POST")
	}
	if got.Status != "connected" {
		t.Errorf("expected Status %q, got %q", "connected", got.Status)
	}
}

func TestUnit_ArrayConnection_Post_Conflict(t *testing.T) {
	srv, facade := newArrayConnectionServer(t)
	facade.store.Seed(&client.ArrayConnection{
		ID:     "conn-existing",
		Remote: client.NamedReference{Name: "existing-remote"},
	})

	c := newTestClient(t, srv)
	_, err := c.PostArrayConnection(context.Background(), "existing-remote", client.ArrayConnectionPost{
		ManagementAddress: "10.1.1.1",
		ConnectionKey:     "key",
	})
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestUnit_ArrayConnection_Patch_ManagementAddress(t *testing.T) {
	srv, facade := newArrayConnectionServer(t)
	facade.store.Seed(&client.ArrayConnection{
		ID:                "conn-patch-1",
		Remote:            client.NamedReference{Name: "patch-remote"},
		ManagementAddress: "10.1.1.1",
		Status:            "connected",
	})

	c := newTestClient(t, srv)
	newAddr := "10.2.2.2"
	got, err := c.PatchArrayConnection(context.Background(), "patch-remote", client.ArrayConnectionPatch{
		ManagementAddress: &newAddr,
	})
	if err != nil {
		t.Fatalf("PatchArrayConnection: %v", err)
	}
	if got.ManagementAddress != "10.2.2.2" {
		t.Errorf("expected ManagementAddress %q, got %q", "10.2.2.2", got.ManagementAddress)
	}
}

func TestUnit_ArrayConnection_Delete(t *testing.T) {
	srv, facade := newArrayConnectionServer(t)
	facade.store.Seed(&client.ArrayConnection{
		ID:     "conn-del-1",
		Remote: client.NamedReference{Name: "del-remote"},
		Status: "connected",
	})

	c := newTestClient(t, srv)
	if err := c.DeleteArrayConnection(context.Background(), "del-remote"); err != nil {
		t.Fatalf("DeleteArrayConnection: %v", err)
	}

	// Subsequent GET must return IsNotFound.
	_, err := c.GetArrayConnection(context.Background(), "del-remote")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true after delete, got false; err: %v", err)
	}
}

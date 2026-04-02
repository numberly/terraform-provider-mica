package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

func newTargetServer(t *testing.T) (*httptest.Server, *targetStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterTargetHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &targetStoreFacade{store: store}
}

// targetStoreFacade wraps the opaque store so tests can call Seed.
type targetStoreFacade struct {
	store interface {
		Seed(t *client.Target)
	}
}

func TestGetTarget_found(t *testing.T) {
	srv, facade := newTargetServer(t)
	facade.store.Seed(&client.Target{
		ID:      "tgt-seed-1",
		Name:    "array-a",
		Address: "10.0.0.1",
		Status:  "connected",
	})

	c := newTestClient(t, srv)
	got, err := c.GetTarget(context.Background(), "array-a")
	if err != nil {
		t.Fatalf("GetTarget: %v", err)
	}
	if got.ID != "tgt-seed-1" {
		t.Errorf("expected ID %q, got %q", "tgt-seed-1", got.ID)
	}
	if got.Name != "array-a" {
		t.Errorf("expected Name %q, got %q", "array-a", got.Name)
	}
	if got.Address != "10.0.0.1" {
		t.Errorf("expected Address %q, got %q", "10.0.0.1", got.Address)
	}
}

func TestGetTarget_notFound(t *testing.T) {
	srv, _ := newTargetServer(t)
	c := newTestClient(t, srv)

	_, err := c.GetTarget(context.Background(), "nonexistent-target")
	if err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestPostTarget(t *testing.T) {
	srv, _ := newTargetServer(t)
	c := newTestClient(t, srv)

	got, err := c.PostTarget(context.Background(), "new-array", client.TargetPost{
		Address: "192.168.1.100",
	})
	if err != nil {
		t.Fatalf("PostTarget: %v", err)
	}
	if got.Name != "new-array" {
		t.Errorf("expected Name %q, got %q", "new-array", got.Name)
	}
	if got.Address != "192.168.1.100" {
		t.Errorf("expected Address %q, got %q", "192.168.1.100", got.Address)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after POST")
	}
	if got.Status != "connected" {
		t.Errorf("expected Status %q, got %q", "connected", got.Status)
	}
}

func TestPatchTarget_address(t *testing.T) {
	srv, facade := newTargetServer(t)
	facade.store.Seed(&client.Target{
		ID:      "tgt-patch-1",
		Name:    "patch-array",
		Address: "10.1.1.1",
		Status:  "connected",
	})

	c := newTestClient(t, srv)
	newAddr := "10.2.2.2"
	got, err := c.PatchTarget(context.Background(), "patch-array", client.TargetPatch{
		Address: &newAddr,
	})
	if err != nil {
		t.Fatalf("PatchTarget address: %v", err)
	}
	if got.Address != "10.2.2.2" {
		t.Errorf("expected Address %q, got %q", "10.2.2.2", got.Address)
	}
}

func TestPatchTarget_caCertGroup(t *testing.T) {
	srv, facade := newTargetServer(t)
	facade.store.Seed(&client.Target{
		ID:      "tgt-patch-2",
		Name:    "cert-array",
		Address: "10.3.3.3",
		Status:  "connected",
	})

	c := newTestClient(t, srv)
	certGroup := &client.NamedReference{Name: "my-ca-group"}
	got, err := c.PatchTarget(context.Background(), "cert-array", client.TargetPatch{
		CACertificateGroup: &certGroup,
	})
	if err != nil {
		t.Fatalf("PatchTarget caCertGroup: %v", err)
	}
	if got.CACertificateGroup == nil {
		t.Fatal("expected CACertificateGroup to be set, got nil")
	}
	if got.CACertificateGroup.Name != "my-ca-group" {
		t.Errorf("expected CACertificateGroup.Name %q, got %q", "my-ca-group", got.CACertificateGroup.Name)
	}
}

func TestDeleteTarget(t *testing.T) {
	srv, facade := newTargetServer(t)
	facade.store.Seed(&client.Target{
		ID:      "tgt-del-1",
		Name:    "del-array",
		Address: "10.4.4.4",
		Status:  "connected",
	})

	c := newTestClient(t, srv)
	if err := c.DeleteTarget(context.Background(), "del-array"); err != nil {
		t.Fatalf("DeleteTarget: %v", err)
	}

	// Subsequent GET must return IsNotFound.
	_, err := c.GetTarget(context.Background(), "del-array")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true after delete, got false; err: %v", err)
	}
}

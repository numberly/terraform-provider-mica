package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

func newCertificateGroupServer(t *testing.T) (*httptest.Server, *handlers.CertificateGroupStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterCertificateGroupHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, handlers.NewCertificateGroupStoreFacade(store)
}

func TestUnit_CertificateGroup_Get_Found(t *testing.T) {
	srv, facade := newCertificateGroupServer(t)
	facade.Seed(&client.CertificateGroup{
		ID:     "certgroup-seed-1",
		Name:   "my-cert-group",
		Realms: []string{"default"},
	})

	c := newTestClient(t, srv)
	got, err := c.GetCertificateGroup(context.Background(), "my-cert-group")
	if err != nil {
		t.Fatalf("GetCertificateGroup: %v", err)
	}
	if got.ID != "certgroup-seed-1" {
		t.Errorf("expected ID %q, got %q", "certgroup-seed-1", got.ID)
	}
	if got.Name != "my-cert-group" {
		t.Errorf("expected Name %q, got %q", "my-cert-group", got.Name)
	}
	if len(got.Realms) != 1 || got.Realms[0] != "default" {
		t.Errorf("expected Realms [\"default\"], got %v", got.Realms)
	}
}

func TestUnit_CertificateGroup_Get_NotFound(t *testing.T) {
	srv, _ := newCertificateGroupServer(t)
	c := newTestClient(t, srv)

	_, err := c.GetCertificateGroup(context.Background(), "nonexistent-group")
	if err == nil {
		t.Fatal("expected error for unknown certificate group, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_CertificateGroup_Post(t *testing.T) {
	srv, _ := newCertificateGroupServer(t)
	c := newTestClient(t, srv)

	got, err := c.PostCertificateGroup(context.Background(), "new-cert-group")
	if err != nil {
		t.Fatalf("PostCertificateGroup: %v", err)
	}
	if got.Name != "new-cert-group" {
		t.Errorf("expected Name %q, got %q", "new-cert-group", got.Name)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after POST")
	}
	if got.Realms == nil {
		t.Error("expected non-nil Realms after POST")
	}
}

func TestUnit_CertificateGroup_Delete(t *testing.T) {
	srv, facade := newCertificateGroupServer(t)
	facade.Seed(&client.CertificateGroup{
		ID:     "certgroup-del-1",
		Name:   "del-cert-group",
		Realms: []string{},
	})

	c := newTestClient(t, srv)
	if err := c.DeleteCertificateGroup(context.Background(), "del-cert-group"); err != nil {
		t.Fatalf("DeleteCertificateGroup: %v", err)
	}

	// Subsequent GET must return IsNotFound.
	_, err := c.GetCertificateGroup(context.Background(), "del-cert-group")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true after delete, got false; err: %v", err)
	}
}

func TestUnit_CertificateGroupMember_Post(t *testing.T) {
	srv, facade := newCertificateGroupServer(t)
	facade.Seed(&client.CertificateGroup{
		ID:     "certgroup-mem-1",
		Name:   "my-group",
		Realms: []string{},
	})

	c := newTestClient(t, srv)
	got, err := c.PostCertificateGroupMember(context.Background(), "my-group", "my-cert")
	if err != nil {
		t.Fatalf("PostCertificateGroupMember: %v", err)
	}
	if got.Group.Name != "my-group" {
		t.Errorf("expected Group.Name %q, got %q", "my-group", got.Group.Name)
	}
	if got.Certificate.Name != "my-cert" {
		t.Errorf("expected Certificate.Name %q, got %q", "my-cert", got.Certificate.Name)
	}
}

func TestUnit_CertificateGroupMember_List(t *testing.T) {
	srv, facade := newCertificateGroupServer(t)
	facade.Seed(&client.CertificateGroup{
		ID:     "certgroup-list-1",
		Name:   "list-group",
		Realms: []string{},
	})
	facade.SeedMember("list-group", client.CertificateGroupMember{
		Certificate: client.NamedReference{Name: "list-cert"},
		Group:       client.NamedReference{Name: "list-group"},
	})

	c := newTestClient(t, srv)
	members, err := c.ListCertificateGroupMembers(context.Background(), "list-group")
	if err != nil {
		t.Fatalf("ListCertificateGroupMembers: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Certificate.Name != "list-cert" {
		t.Errorf("expected Certificate.Name %q, got %q", "list-cert", members[0].Certificate.Name)
	}
}

func TestUnit_CertificateGroupMember_Delete(t *testing.T) {
	srv, facade := newCertificateGroupServer(t)
	facade.Seed(&client.CertificateGroup{
		ID:     "certgroup-mdb-1",
		Name:   "mdb-group",
		Realms: []string{},
	})
	facade.SeedMember("mdb-group", client.CertificateGroupMember{
		Certificate: client.NamedReference{Name: "mdb-cert"},
		Group:       client.NamedReference{Name: "mdb-group"},
	})

	c := newTestClient(t, srv)

	if err := c.DeleteCertificateGroupMember(context.Background(), "mdb-group", "mdb-cert"); err != nil {
		t.Fatalf("DeleteCertificateGroupMember: %v", err)
	}

	// List must now be empty.
	members, err := c.ListCertificateGroupMembers(context.Background(), "mdb-group")
	if err != nil {
		t.Fatalf("ListCertificateGroupMembers after delete: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("expected 0 members after delete, got %d", len(members))
	}
}

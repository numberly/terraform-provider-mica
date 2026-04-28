package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

func newTlsPolicyServer(t *testing.T) (*httptest.Server, *handlers.TlsPolicyStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterTlsPolicyHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, handlers.NewTlsPolicyStoreFacade(store)
}

func TestUnit_TlsPolicy_Get_Found(t *testing.T) {
	srv, facade := newTlsPolicyServer(t)
	facade.Seed(&client.TlsPolicy{
		ID:            "tls-seed-1",
		Name:          "my-tls-policy",
		Enabled:       true,
		IsLocal:       true,
		MinTlsVersion: "TLSv1_2",
		PolicyType:    "tls",
	})

	c := newTestClient(t, srv)
	got, err := c.GetTlsPolicy(context.Background(), "my-tls-policy")
	if err != nil {
		t.Fatalf("GetTlsPolicy: %v", err)
	}
	if got.ID != "tls-seed-1" {
		t.Errorf("expected ID %q, got %q", "tls-seed-1", got.ID)
	}
	if got.Name != "my-tls-policy" {
		t.Errorf("expected Name %q, got %q", "my-tls-policy", got.Name)
	}
	if got.MinTlsVersion != "TLSv1_2" {
		t.Errorf("expected MinTlsVersion %q, got %q", "TLSv1_2", got.MinTlsVersion)
	}
	if !got.Enabled {
		t.Error("expected Enabled true, got false")
	}
}

func TestUnit_TlsPolicy_Get_NotFound(t *testing.T) {
	srv, _ := newTlsPolicyServer(t)
	c := newTestClient(t, srv)

	_, err := c.GetTlsPolicy(context.Background(), "nonexistent-policy")
	if err == nil {
		t.Fatal("expected error for unknown TLS policy, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_TlsPolicy_Post(t *testing.T) {
	srv, _ := newTlsPolicyServer(t)
	c := newTestClient(t, srv)

	appCert := &client.NamedReference{Name: "my-cert"}
	got, err := c.PostTlsPolicy(context.Background(), "new-tls-policy", client.TlsPolicyPost{
		ApplianceCertificate:       appCert,
		ClientCertificatesRequired: true,
		Enabled:                    true,
		MinTlsVersion:              "TLSv1_2",
		VerifyClientCertificateTrust: true,
	})
	if err != nil {
		t.Fatalf("PostTlsPolicy: %v", err)
	}
	if got.Name != "new-tls-policy" {
		t.Errorf("expected Name %q, got %q", "new-tls-policy", got.Name)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after POST")
	}
	if got.PolicyType != "tls" {
		t.Errorf("expected PolicyType %q, got %q", "tls", got.PolicyType)
	}
	if !got.IsLocal {
		t.Error("expected IsLocal true, got false")
	}
	if got.ApplianceCertificate == nil {
		t.Fatal("expected ApplianceCertificate to be set, got nil")
	}
	if got.ApplianceCertificate.Name != "my-cert" {
		t.Errorf("expected ApplianceCertificate.Name %q, got %q", "my-cert", got.ApplianceCertificate.Name)
	}
	if got.MinTlsVersion != "TLSv1_2" {
		t.Errorf("expected MinTlsVersion %q, got %q", "TLSv1_2", got.MinTlsVersion)
	}
}

func TestUnit_TlsPolicy_Patch_MinTlsVersion(t *testing.T) {
	srv, facade := newTlsPolicyServer(t)
	facade.Seed(&client.TlsPolicy{
		ID:            "tls-patch-1",
		Name:          "patch-policy",
		Enabled:       true,
		IsLocal:       true,
		MinTlsVersion: "TLSv1_2",
		PolicyType:    "tls",
	})

	c := newTestClient(t, srv)
	newVer := "TLSv1_3"
	got, err := c.PatchTlsPolicy(context.Background(), "patch-policy", client.TlsPolicyPatch{
		MinTlsVersion: &newVer,
	})
	if err != nil {
		t.Fatalf("PatchTlsPolicy MinTlsVersion: %v", err)
	}
	if got.MinTlsVersion != "TLSv1_3" {
		t.Errorf("expected MinTlsVersion %q, got %q", "TLSv1_3", got.MinTlsVersion)
	}
	// Other fields must be unchanged.
	if !got.Enabled {
		t.Error("expected Enabled true after patch, got false")
	}
}

func TestUnit_TlsPolicy_Delete(t *testing.T) {
	srv, facade := newTlsPolicyServer(t)
	facade.Seed(&client.TlsPolicy{
		ID:      "tls-del-1",
		Name:    "del-policy",
		Enabled: true,
		IsLocal: true,
	})

	c := newTestClient(t, srv)
	if err := c.DeleteTlsPolicy(context.Background(), "del-policy"); err != nil {
		t.Fatalf("DeleteTlsPolicy: %v", err)
	}

	// Subsequent GET must return IsNotFound.
	_, err := c.GetTlsPolicy(context.Background(), "del-policy")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true after delete, got false; err: %v", err)
	}
}

func TestUnit_TlsPolicyMember_Post(t *testing.T) {
	srv, facade := newTlsPolicyServer(t)
	facade.Seed(&client.TlsPolicy{
		ID:      "tls-mem-1",
		Name:    "member-policy",
		Enabled: true,
		IsLocal: true,
	})

	c := newTestClient(t, srv)
	got, err := c.PostTlsPolicyMember(context.Background(), "member-policy", "eth0.data")
	if err != nil {
		t.Fatalf("PostTlsPolicyMember: %v", err)
	}
	if got.Policy.Name != "member-policy" {
		t.Errorf("expected Policy.Name %q, got %q", "member-policy", got.Policy.Name)
	}
	if got.Member.Name != "eth0.data" {
		t.Errorf("expected Member.Name %q, got %q", "eth0.data", got.Member.Name)
	}
}

func TestUnit_TlsPolicyMember_List(t *testing.T) {
	srv, facade := newTlsPolicyServer(t)
	facade.Seed(&client.TlsPolicy{
		ID:      "tls-list-1",
		Name:    "list-policy",
		Enabled: true,
		IsLocal: true,
	})
	facade.SeedMember("list-policy", client.TlsPolicyMember{
		Policy: client.NamedReference{Name: "list-policy"},
		Member: client.NamedReference{Name: "eth1.data"},
	})

	c := newTestClient(t, srv)
	members, err := c.ListTlsPolicyMembers(context.Background(), "list-policy")
	if err != nil {
		t.Fatalf("ListTlsPolicyMembers: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Member.Name != "eth1.data" {
		t.Errorf("expected Member.Name %q, got %q", "eth1.data", members[0].Member.Name)
	}
}

func TestUnit_TlsPolicyMember_Delete(t *testing.T) {
	srv, facade := newTlsPolicyServer(t)
	facade.Seed(&client.TlsPolicy{
		ID:      "tls-mdb-1",
		Name:    "mdb-policy",
		Enabled: true,
		IsLocal: true,
	})
	facade.SeedMember("mdb-policy", client.TlsPolicyMember{
		Policy: client.NamedReference{Name: "mdb-policy"},
		Member: client.NamedReference{Name: "eth2.data"},
	})

	c := newTestClient(t, srv)

	// Delete the member.
	if err := c.DeleteTlsPolicyMember(context.Background(), "mdb-policy", "eth2.data"); err != nil {
		t.Fatalf("DeleteTlsPolicyMember: %v", err)
	}

	// List must now be empty.
	members, err := c.ListTlsPolicyMembers(context.Background(), "mdb-policy")
	if err != nil {
		t.Fatalf("ListTlsPolicyMembers after delete: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("expected 0 members after delete, got %d", len(members))
	}
}

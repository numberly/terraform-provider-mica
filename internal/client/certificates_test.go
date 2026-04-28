package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

func newCertificateServer(t *testing.T) (*httptest.Server, *certificateStoreFacade) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	store := handlers.RegisterCertificateHandlers(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &certificateStoreFacade{store: store}
}

// certificateStoreFacade wraps the opaque store so tests can call Seed.
type certificateStoreFacade struct {
	store interface {
		Seed(item *client.Certificate)
	}
}

func TestUnit_Certificate_Get_Found(t *testing.T) {
	srv, facade := newCertificateServer(t)
	facade.store.Seed(&client.Certificate{
		ID:              "cert-seed-1",
		Name:            "my-cert",
		Certificate:     "-----BEGIN CERTIFICATE-----\nMIItest\n-----END CERTIFICATE-----",
		CertificateType: "array",
		CommonName:      "test-cert",
		IssuedBy:        "CN=Test CA",
		IssuedTo:        "CN=test-cert",
		Status:          "imported",
		ValidFrom:       1700000000000,
		ValidTo:         1731536000000,
		KeyAlgorithm:    "RSA",
		KeySize:         2048,
	})

	c := newTestClient(t, srv)
	got, err := c.GetCertificate(context.Background(), "my-cert")
	if err != nil {
		t.Fatalf("GetCertificate: %v", err)
	}
	if got.ID != "cert-seed-1" {
		t.Errorf("expected ID %q, got %q", "cert-seed-1", got.ID)
	}
	if got.Name != "my-cert" {
		t.Errorf("expected Name %q, got %q", "my-cert", got.Name)
	}
	if got.IssuedBy != "CN=Test CA" {
		t.Errorf("expected IssuedBy %q, got %q", "CN=Test CA", got.IssuedBy)
	}
	if got.IssuedTo != "CN=test-cert" {
		t.Errorf("expected IssuedTo %q, got %q", "CN=test-cert", got.IssuedTo)
	}
	if got.Status != "imported" {
		t.Errorf("expected Status %q, got %q", "imported", got.Status)
	}
	if got.KeyAlgorithm != "RSA" {
		t.Errorf("expected KeyAlgorithm %q, got %q", "RSA", got.KeyAlgorithm)
	}
	if got.KeySize != 2048 {
		t.Errorf("expected KeySize %d, got %d", 2048, got.KeySize)
	}
	if got.ValidFrom != 1700000000000 {
		t.Errorf("expected ValidFrom %d, got %d", int64(1700000000000), got.ValidFrom)
	}
	if got.ValidTo != 1731536000000 {
		t.Errorf("expected ValidTo %d, got %d", int64(1731536000000), got.ValidTo)
	}
}

func TestUnit_Certificate_Get_NotFound(t *testing.T) {
	srv, _ := newCertificateServer(t)
	c := newTestClient(t, srv)

	_, err := c.GetCertificate(context.Background(), "nonexistent-cert")
	if err == nil {
		t.Fatal("expected error for unknown certificate, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_Certificate_Post(t *testing.T) {
	srv, _ := newCertificateServer(t)
	c := newTestClient(t, srv)

	pem := "-----BEGIN CERTIFICATE-----\nMIItest\n-----END CERTIFICATE-----"
	got, err := c.PostCertificate(context.Background(), "new-cert", client.CertificatePost{
		Certificate:     pem,
		CertificateType: "array",
		PrivateKey:      "-----BEGIN PRIVATE KEY-----\nMIIkey\n-----END PRIVATE KEY-----",
	})
	if err != nil {
		t.Fatalf("PostCertificate: %v", err)
	}
	if got.Name != "new-cert" {
		t.Errorf("expected Name %q, got %q", "new-cert", got.Name)
	}
	if got.Certificate != pem {
		t.Errorf("expected Certificate to match PEM body")
	}
	if got.ID == "" {
		t.Error("expected non-empty ID after POST")
	}
	if got.Status != "imported" {
		t.Errorf("expected Status %q, got %q", "imported", got.Status)
	}
	if got.CertificateType != "array" {
		t.Errorf("expected CertificateType %q, got %q", "array", got.CertificateType)
	}
	if got.IssuedBy != "CN=Test CA" {
		t.Errorf("expected IssuedBy %q, got %q", "CN=Test CA", got.IssuedBy)
	}
	if got.IssuedTo != "CN=test-cert" {
		t.Errorf("expected IssuedTo %q, got %q", "CN=test-cert", got.IssuedTo)
	}
	if got.KeyAlgorithm != "RSA" {
		t.Errorf("expected KeyAlgorithm %q, got %q", "RSA", got.KeyAlgorithm)
	}
	if got.KeySize != 2048 {
		t.Errorf("expected KeySize %d, got %d", 2048, got.KeySize)
	}
}

func TestUnit_Certificate_Patch(t *testing.T) {
	srv, facade := newCertificateServer(t)
	facade.store.Seed(&client.Certificate{
		ID:          "cert-patch-1",
		Name:        "patch-cert",
		Certificate: "-----BEGIN CERTIFICATE-----\nold\n-----END CERTIFICATE-----",
		Status:      "imported",
		KeyAlgorithm: "RSA",
		KeySize:     2048,
	})

	c := newTestClient(t, srv)
	newPEM := "-----BEGIN CERTIFICATE-----\nnew\n-----END CERTIFICATE-----"
	got, err := c.PatchCertificate(context.Background(), "patch-cert", client.CertificatePatch{
		Certificate: &newPEM,
	})
	if err != nil {
		t.Fatalf("PatchCertificate: %v", err)
	}
	if got.Certificate != newPEM {
		t.Errorf("expected updated Certificate PEM, got %q", got.Certificate)
	}
	if got.Name != "patch-cert" {
		t.Errorf("expected Name %q, got %q", "patch-cert", got.Name)
	}
}

func TestUnit_Certificate_Delete(t *testing.T) {
	srv, facade := newCertificateServer(t)
	facade.store.Seed(&client.Certificate{
		ID:     "cert-del-1",
		Name:   "del-cert",
		Status: "imported",
	})

	c := newTestClient(t, srv)
	if err := c.DeleteCertificate(context.Background(), "del-cert"); err != nil {
		t.Fatalf("DeleteCertificate: %v", err)
	}

	// Subsequent GET must return IsNotFound.
	_, err := c.GetCertificate(context.Background(), "del-cert")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true after delete, got false; err: %v", err)
	}
}

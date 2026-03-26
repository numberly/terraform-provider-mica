package client_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/soulkyu/terraform-provider-flashblade/internal/client"
)

// generateTestCACert generates a self-signed CA certificate for tests.
// Returns the PEM-encoded certificate and the TLS certificate.
func generateTestCACert(t *testing.T) (pemBytes []byte, tlsCert tls.Certificate) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	pemBytes = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	tlsCert, err = tls.X509KeyPair(pemBytes, keyPEM)
	if err != nil {
		t.Fatalf("x509 key pair: %v", err)
	}
	return pemBytes, tlsCert
}

func TestUnit_NewClient_WithAPIToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.Header().Set("x-auth-token", "test-session-token")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint: srv.URL,
		APIToken: "test-api-token",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestUnit_NewClient_MissingEndpoint(t *testing.T) {
	_, err := client.NewClient(client.Config{
		APIToken: "some-token",
	})
	if err == nil {
		t.Fatal("expected error for missing endpoint, got nil")
	}
}

func TestUnit_CustomCATLS(t *testing.T) {
	caPEM, tlsCert := generateTestCACert(t)

	// Write PEM to temp file.
	f, err := os.CreateTemp(t.TempDir(), "ca-cert-*.pem")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.Write(caPEM); err != nil {
		t.Fatalf("write ca cert: %v", err)
	}
	f.Close()

	// Create TLS test server using the generated CA cert.
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.Header().Set("x-auth-token", "token")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	srv.StartTLS()
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint:   srv.URL,
		APIToken:   "test-token",
		CACertFile: f.Name(),
	})
	if err != nil {
		t.Fatalf("expected no error with custom CA file, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestUnit_CustomCATLS_InlinePEM(t *testing.T) {
	caPEM, tlsCert := generateTestCACert(t)

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.Header().Set("x-auth-token", "token")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	srv.StartTLS()
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint: srv.URL,
		APIToken: "test-token",
		CACert:   string(caPEM),
	})
	if err != nil {
		t.Fatalf("expected no error with inline CA PEM, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestUnit_InsecureSkipVerify(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.Header().Set("x-auth-token", "token")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint:           srv.URL,
		APIToken:           "test-token",
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("expected no error with InsecureSkipVerify, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestUnit_NegotiateVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.Header().Set("x-auth-token", "token")
			w.WriteHeader(http.StatusOK)
		case "/api/api_version":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"versions":["2.12","2.22"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint: srv.URL,
		APIToken: "test-token",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := c.NegotiateVersion(context.Background()); err != nil {
		t.Fatalf("expected version negotiation to succeed, got: %v", err)
	}
}

func TestUnit_NegotiateVersion_Missing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/login":
			w.Header().Set("x-auth-token", "token")
			w.WriteHeader(http.StatusOK)
		case "/api/api_version":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"versions":["2.12","2.15"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := client.NewClient(client.Config{
		Endpoint: srv.URL,
		APIToken: "test-token",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := c.NegotiateVersion(context.Background()); err == nil {
		t.Fatal("expected version negotiation to fail when v2.22 is absent")
	}
}

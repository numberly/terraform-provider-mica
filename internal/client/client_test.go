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
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// generateTestCerts generates a self-signed CA and a server certificate signed
// by that CA (with 127.0.0.1 as SAN). Returns:
//   - caPEM: PEM-encoded CA certificate for client trust configuration
//   - serverTLSCert: TLS certificate to use on the test server
func generateTestCerts(t *testing.T) (caPEM []byte, serverTLSCert tls.Certificate) {
	t.Helper()

	// Generate CA key.
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}

	// Self-signed CA certificate.
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}
	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	// Generate server key.
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate server key: %v", err)
	}

	// Server certificate signed by the CA — with 127.0.0.1 as IP SAN.
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create server cert: %v", err)
	}
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	serverKeyDER, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		t.Fatalf("marshal server key: %v", err)
	}
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyDER})

	serverTLSCert, err = tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatalf("x509 key pair: %v", err)
	}
	return caPEM, serverTLSCert
}

// loginHandler is a shared handler for the /api/login endpoint.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/login" {
		w.Header().Set("x-auth-token", "test-session-token")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func TestUnit_NewClient_WithAPIToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(loginHandler))
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
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
	_, err := client.NewClient(context.Background(), client.Config{
		APIToken: "some-token",
	})
	if err == nil {
		t.Fatal("expected error for missing endpoint, got nil")
	}
}

func TestUnit_CustomCATLS(t *testing.T) {
	caPEM, serverTLSCert := generateTestCerts(t)

	// Write CA PEM to a temp file.
	f, err := os.CreateTemp(t.TempDir(), "ca-cert-*.pem")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.Write(caPEM); err != nil {
		t.Fatalf("write CA cert: %v", err)
	}
	f.Close()

	// TLS test server using the server cert (signed by our CA).
	srv := httptest.NewUnstartedServer(http.HandlerFunc(loginHandler))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{serverTLSCert}}
	srv.StartTLS()
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
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
	caPEM, serverTLSCert := generateTestCerts(t)

	srv := httptest.NewUnstartedServer(http.HandlerFunc(loginHandler))
	srv.TLS = &tls.Config{Certificates: []tls.Certificate{serverTLSCert}}
	srv.StartTLS()
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
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
	srv := httptest.NewTLSServer(http.HandlerFunc(loginHandler))
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
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
			_, _ = w.Write([]byte(`{"versions":["2.12","2.22"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
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
			_, _ = w.Write([]byte(`{"versions":["2.12","2.15"]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
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

func TestUnit_NewClient_HTTPTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(loginHandler))
	defer srv.Close()

	c, err := client.NewClient(context.Background(), client.Config{
		Endpoint: srv.URL,
		APIToken: "test-token",
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if c.HTTPClient().Timeout != 30*time.Second {
		t.Errorf("expected HTTP client timeout of 30s, got %v", c.HTTPClient().Timeout)
	}
}

func TestUnit_NewClient_ContextPropagation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(loginHandler))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := client.NewClient(ctx, client.Config{
		Endpoint: srv.URL,
		APIToken: "test-token",
	})
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

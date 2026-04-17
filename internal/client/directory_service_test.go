package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// dsResponse builds a JSON response body with the supplied DirectoryService items.
func dsResponse(items ...client.DirectoryService) []byte {
	type listResp struct {
		Items []client.DirectoryService `json:"items"`
	}
	b, _ := json.Marshal(listResp{Items: items})
	return b
}

func newDSManagementServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-auth-token", "tok")
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/api/2.22/directory-services", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestUnit_DirectoryServiceManagement_Get_Found(t *testing.T) {
	seed := client.DirectoryService{
		ID:      "ds-1",
		Name:    "management",
		Enabled: true,
		URIs:    []string{"ldaps://ldap.example.com:636"},
		BaseDN:  "dc=example,dc=com",
		BindUser: "cn=binder",
		Management: client.DirectoryServiceManagement{
			UserLoginAttribute:    "sAMAccountName",
			UserObjectClass:       "User",
			SSHPublicKeyAttribute: "sshPublicKey",
		},
		Services: []string{"management"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("names") == "management" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(dsResponse(seed))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsResponse())
	})

	srv := newDSManagementServer(t, handler)
	c := newTestClient(t, srv)

	got, err := c.GetDirectoryServiceManagement(context.Background(), "management")
	if err != nil {
		t.Fatalf("GetDirectoryServiceManagement: %v", err)
	}
	if got.ID != "ds-1" {
		t.Errorf("expected ID %q, got %q", "ds-1", got.ID)
	}
	if got.Name != "management" {
		t.Errorf("expected Name %q, got %q", "management", got.Name)
	}
	if !got.Enabled {
		t.Errorf("expected Enabled=true, got false")
	}
	if len(got.URIs) != 1 {
		t.Errorf("expected 1 URI, got %d", len(got.URIs))
	} else if got.URIs[0] != "ldaps://ldap.example.com:636" {
		t.Errorf("expected URIs[0] %q, got %q", "ldaps://ldap.example.com:636", got.URIs[0])
	}
	if got.BaseDN != "dc=example,dc=com" {
		t.Errorf("expected BaseDN %q, got %q", "dc=example,dc=com", got.BaseDN)
	}
	if got.Management.UserLoginAttribute != "sAMAccountName" {
		t.Errorf("expected Management.UserLoginAttribute %q, got %q", "sAMAccountName", got.Management.UserLoginAttribute)
	}
}

func TestUnit_DirectoryServiceManagement_Get_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(dsResponse())
	})

	srv := newDSManagementServer(t, handler)
	c := newTestClient(t, srv)

	_, err := c.GetDirectoryServiceManagement(context.Background(), "management")
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_DirectoryServiceManagement_Patch_Uris(t *testing.T) {
	var capturedBody []byte

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			b, _ := io.ReadAll(r.Body)
			capturedBody = b
			// Return a minimal valid response
			resp := client.DirectoryService{
				ID:      "ds-1",
				Name:    "management",
				Enabled: false,
				URIs:    []string{"ldaps://new.example.com"},
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(dsResponse(resp))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(dsResponse())
	})

	srv := newDSManagementServer(t, handler)
	c := newTestClient(t, srv)

	body := client.DirectoryServicePatch{
		URIs: &[]string{"ldaps://new.example.com"},
	}
	_, err := c.PatchDirectoryServiceManagement(context.Background(), "management", body)
	if err != nil {
		t.Fatalf("PatchDirectoryServiceManagement: %v", err)
	}

	if !bytes.Contains(capturedBody, []byte(`"uris":["ldaps://new.example.com"]`)) {
		t.Errorf("expected body to contain uris, got: %s", capturedBody)
	}
	if bytes.Contains(capturedBody, []byte(`"enabled"`)) {
		t.Errorf("expected enabled to be omitted, but found in body: %s", capturedBody)
	}
	if bytes.Contains(capturedBody, []byte(`"base_dn"`)) {
		t.Errorf("expected base_dn to be omitted, but found in body: %s", capturedBody)
	}
	if bytes.Contains(capturedBody, []byte(`"bind_user"`)) {
		t.Errorf("expected bind_user to be omitted, but found in body: %s", capturedBody)
	}
	if bytes.Contains(capturedBody, []byte(`"ca_certificate"`)) {
		t.Errorf("expected ca_certificate to be omitted, but found in body: %s", capturedBody)
	}
	if bytes.Contains(capturedBody, []byte(`"management":{`)) {
		t.Errorf("expected management to be omitted, but found in body: %s", capturedBody)
	}
}

func TestUnit_DirectoryServiceManagement_Patch_CACertificateGroup(t *testing.T) {
	makeHandler := func(capturedBody *[]byte) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPatch {
				b, _ := io.ReadAll(r.Body)
				*capturedBody = b
				resp := client.DirectoryService{ID: "ds-1", Name: "management"}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(dsResponse(resp))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(dsResponse())
		})
	}

	t.Run("set", func(t *testing.T) {
		var capturedBody []byte
		srv := newDSManagementServer(t, makeHandler(&capturedBody))
		c := newTestClient(t, srv)

		ref := &client.NamedReference{Name: "corp-ca"}
		body := client.DirectoryServicePatch{CACertificateGroup: &ref}
		_, err := c.PatchDirectoryServiceManagement(context.Background(), "management", body)
		if err != nil {
			t.Fatalf("PatchDirectoryServiceManagement set: %v", err)
		}
		if !bytes.Contains(capturedBody, []byte(`"ca_certificate_group":{"name":"corp-ca"}`)) {
			t.Errorf("expected ca_certificate_group set to corp-ca, got: %s", capturedBody)
		}
	})

	t.Run("clear", func(t *testing.T) {
		var capturedBody []byte
		srv := newDSManagementServer(t, makeHandler(&capturedBody))
		c := newTestClient(t, srv)

		var nilRef *client.NamedReference
		body := client.DirectoryServicePatch{CACertificateGroup: &nilRef}
		_, err := c.PatchDirectoryServiceManagement(context.Background(), "management", body)
		if err != nil {
			t.Fatalf("PatchDirectoryServiceManagement clear: %v", err)
		}
		if !bytes.Contains(capturedBody, []byte(`"ca_certificate_group":null`)) {
			t.Errorf("expected ca_certificate_group null (clear), got: %s", capturedBody)
		}
	})
}

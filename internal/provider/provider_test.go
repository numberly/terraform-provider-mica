package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// oauth2Type is the tftypes representation of the oauth2 sub-block.
var oauth2Type = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"client_id": tftypes.String,
		"key_id":    tftypes.String,
		"issuer":    tftypes.String,
	},
}

// authType is the tftypes representation of the auth block.
var authType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"api_token": tftypes.String,
		"oauth2":    oauth2Type,
	},
}

// providerType is the tftypes representation of the full provider config object.
var providerType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"endpoint":             tftypes.String,
		"ca_cert_file":         tftypes.String,
		"ca_cert":              tftypes.String,
		"insecure_skip_verify": tftypes.Bool,
		"max_retries":          tftypes.Number,
		"auth": authType,
	},
}

// nullOAuth2 is a null oauth2 nested value.
var nullOAuth2 = tftypes.NewValue(oauth2Type, nil)

// nullAuth is a null auth nested value (no auth block provided).
var nullAuth = tftypes.NewValue(authType, nil)

// authWithAPIToken builds an auth block value with api_token set.
func authWithAPIToken(token string) tftypes.Value {
	return tftypes.NewValue(authType, map[string]tftypes.Value{
		"api_token": tftypes.NewValue(tftypes.String, token),
		"oauth2":    nullOAuth2,
	})
}

// baseProviderConfig returns a config map with all attributes set to null
// except those provided in overrides.
func baseProviderConfig(overrides map[string]tftypes.Value) tftypes.Value {
	base := map[string]tftypes.Value{
		"endpoint":             tftypes.NewValue(tftypes.String, nil),
		"ca_cert_file":         tftypes.NewValue(tftypes.String, nil),
		"ca_cert":              tftypes.NewValue(tftypes.String, nil),
		"insecure_skip_verify": tftypes.NewValue(tftypes.Bool, nil),
		"max_retries":          tftypes.NewValue(tftypes.Number, nil),
		"auth": nullAuth,
	}
	for k, v := range overrides {
		base[k] = v
	}
	return tftypes.NewValue(providerType, base)
}

// newTestSchema returns the provider schema for use in test config construction.
func newTestSchema(t *testing.T) schema.Schema {
	t.Helper()
	p := &FlashBladeProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)
	return resp.Schema
}

// configureWithValue calls Configure with the given tftypes.Value and returns the response.
func configureWithValue(t *testing.T, val tftypes.Value) provider.ConfigureResponse {
	t.Helper()
	s := newTestSchema(t)
	p := &FlashBladeProvider{version: "test"}
	req := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw:    val,
			Schema: s,
		},
	}
	var resp provider.ConfigureResponse
	p.Configure(context.Background(), req, &resp)
	return resp
}

// mockFlashBladeServer creates a test HTTP server that handles FlashBlade API endpoints.
// It returns (server, closeFunc).
func mockFlashBladeServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/api_version":
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"versions": []string{"2.10", "2.22"},
			}); err != nil {
				http.Error(w, "encode error", http.StatusInternalServerError)
			}
		case "/api/login":
			w.Header().Set("x-auth-token", "test-session-token")
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestUnit_ProviderSchema verifies that the schema exposes all expected attributes:
// endpoint, auth block, ca_cert_file, ca_cert, insecure_skip_verify, max_retries,
// and that auth has api_token and oauth2 sub-block.
func TestUnit_ProviderSchema(t *testing.T) {
	p := &FlashBladeProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"endpoint", "ca_cert_file", "ca_cert", "insecure_skip_verify", "max_retries", "auth"}
	for _, name := range required {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}

	// Verify auth is a SingleNestedAttribute with the expected sub-attributes.
	authAttr, ok := attrs["auth"]
	if !ok {
		t.Fatal("schema missing 'auth' attribute")
	}
	nestedAuth, ok := authAttr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("'auth' attribute is not a SingleNestedAttribute, got %T", authAttr)
	}
	if _, ok := nestedAuth.Attributes["api_token"]; !ok {
		t.Error("auth block missing 'api_token' attribute")
	}
	if _, ok := nestedAuth.Attributes["oauth2"]; !ok {
		t.Error("auth block missing 'oauth2' attribute")
	}

	// Verify oauth2 is a SingleNestedAttribute with the expected sub-attributes.
	oauth2Attr, ok := nestedAuth.Attributes["oauth2"]
	if !ok {
		t.Fatal("auth block missing 'oauth2' attribute")
	}
	nestedOAuth2, ok := oauth2Attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("'oauth2' attribute is not a SingleNestedAttribute, got %T", oauth2Attr)
	}
	for _, name := range []string{"client_id", "key_id", "issuer"} {
		if _, ok := nestedOAuth2.Attributes[name]; !ok {
			t.Errorf("oauth2 block missing %q attribute", name)
		}
	}
}

// TestUnit_SensitiveAttributes verifies that api_token, client_id, and key_id are
// marked Sensitive in the schema (PROV-05).
func TestUnit_SensitiveAttributes(t *testing.T) {
	p := &FlashBladeProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	authAttr, ok := resp.Schema.Attributes["auth"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("auth is not a SingleNestedAttribute")
	}

	apiTokenAttr, ok := authAttr.Attributes["api_token"].(schema.StringAttribute)
	if !ok {
		t.Fatal("api_token is not a StringAttribute")
	}
	if !apiTokenAttr.Sensitive {
		t.Error("api_token should be Sensitive")
	}

	oauth2Attr, ok := authAttr.Attributes["oauth2"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("oauth2 is not a SingleNestedAttribute")
	}

	for _, name := range []string{"client_id", "key_id"} {
		attr, ok := oauth2Attr.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("oauth2.%s is not a StringAttribute", name)
		}
		if !attr.Sensitive {
			t.Errorf("oauth2.%s should be Sensitive", name)
		}
	}
}

// TestUnit_Configure_MissingEndpoint verifies that Configure returns a diagnostic error
// when no endpoint is set in config and FLASHBLADE_HOST env var is absent.
func TestUnit_Configure_MissingEndpoint(t *testing.T) {
	t.Setenv("FLASHBLADE_HOST", "")
	t.Setenv("FLASHBLADE_API_TOKEN", "")

	val := baseProviderConfig(nil)
	resp := configureWithValue(t, val)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for missing endpoint, got none")
	}
}

// TestUnit_Configure_MissingAuth verifies that Configure returns a diagnostic error
// when an endpoint is set but no auth credentials are provided or present in env vars.
func TestUnit_Configure_MissingAuth(t *testing.T) {
	srv := mockFlashBladeServer(t)

	t.Setenv("FLASHBLADE_API_TOKEN", "")
	t.Setenv("FLASHBLADE_OAUTH2_CLIENT_ID", "")
	t.Setenv("FLASHBLADE_OAUTH2_KEY_ID", "")
	t.Setenv("FLASHBLADE_OAUTH2_ISSUER", "")

	val := baseProviderConfig(map[string]tftypes.Value{
		"endpoint": tftypes.NewValue(tftypes.String, srv.URL),
	})
	resp := configureWithValue(t, val)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for missing auth, got none")
	}
}

// TestUnit_EnvVarFallback verifies that Configure reads FLASHBLADE_* env vars
// when all config values are null. This covers:
//   - FLASHBLADE_HOST (endpoint)
//   - FLASHBLADE_API_TOKEN (api_token)
func TestUnit_EnvVarFallback(t *testing.T) {
	srv := mockFlashBladeServer(t)

	t.Setenv("FLASHBLADE_HOST", srv.URL)
	t.Setenv("FLASHBLADE_API_TOKEN", "env-api-token")

	// All config values null — should fall back to env vars.
	val := baseProviderConfig(nil)
	resp := configureWithValue(t, val)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error when env vars set, got: %s", resp.Diagnostics)
	}
}

// TestUnit_TflogOutput verifies that Configure completes without panic when both
// endpoint and auth are set — tflog.Info calls are fire-and-forget.
func TestUnit_TflogOutput(t *testing.T) {
	srv := mockFlashBladeServer(t)

	t.Setenv("FLASHBLADE_HOST", "")
	t.Setenv("FLASHBLADE_API_TOKEN", "")

	val := baseProviderConfig(map[string]tftypes.Value{
		"endpoint": tftypes.NewValue(tftypes.String, srv.URL),
		"auth":     authWithAPIToken("test-token"),
	})
	resp := configureWithValue(t, val)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error with valid config: %s", resp.Diagnostics)
	}
}

// TestUnit_Configure_VersionMismatch verifies that Configure returns a diagnostic
// error when the FlashBlade endpoint does not support API version v2.22.
func TestUnit_Configure_VersionMismatch(t *testing.T) {
	// Server that only supports old API versions.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/api_version":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"versions": []string{"2.10", "2.15"},
			})
		case "/api/login":
			w.Header().Set("x-auth-token", "test-token")
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)

	t.Setenv("FLASHBLADE_HOST", "")
	t.Setenv("FLASHBLADE_API_TOKEN", "")

	val := baseProviderConfig(map[string]tftypes.Value{
		"endpoint": tftypes.NewValue(tftypes.String, srv.URL),
		"auth":     authWithAPIToken("test-token"),
	})
	resp := configureWithValue(t, val)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error diagnostic for unsupported API version, got none")
	}
}

// TestUnit_Configure_OAuth2 verifies that OAuth2 provider configuration (client_id + key_id)
// initializes without errors when a mock API version endpoint is available.
func TestUnit_Configure_OAuth2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/api_version":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"versions": []string{"2.10", "2.22"},
			})
		case "/oauth2/1.0/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "mock-oauth2-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	t.Setenv("FLASHBLADE_HOST", "")
	t.Setenv("FLASHBLADE_API_TOKEN", "")
	t.Setenv("FLASHBLADE_OAUTH2_CLIENT_ID", "")
	t.Setenv("FLASHBLADE_OAUTH2_KEY_ID", "")
	t.Setenv("FLASHBLADE_OAUTH2_ISSUER", "")

	val := baseProviderConfig(map[string]tftypes.Value{
		"endpoint":             tftypes.NewValue(tftypes.String, srv.URL),
		"insecure_skip_verify": tftypes.NewValue(tftypes.Bool, true),
		"auth": tftypes.NewValue(authType, map[string]tftypes.Value{
			"api_token": tftypes.NewValue(tftypes.String, nil),
			"oauth2": tftypes.NewValue(oauth2Type, map[string]tftypes.Value{
				"client_id": tftypes.NewValue(tftypes.String, "test-client-id"),
				"key_id":    tftypes.NewValue(tftypes.String, "test-key-id"),
				"issuer":    tftypes.NewValue(tftypes.String, "test-issuer"),
			}),
		}),
	})
	resp := configureWithValue(t, val)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error with OAuth2 config, got: %s", resp.Diagnostics)
	}
	if resp.ResourceData == nil {
		t.Error("expected ResourceData to be non-nil after OAuth2 configuration")
	}
}

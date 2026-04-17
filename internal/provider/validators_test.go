package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAlphanumericValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"simple lowercase", "myname", false},
		{"mixed case with digits", "Rule01", false},
		{"uppercase only", "ABC", false},
		{"digits only", "123", false},
		{"contains hyphen", "my-name", true},
		{"contains dot", "my.name", true},
		{"contains space", "my name", true},
		{"empty string", "", true},
		{"contains underscore", "my_name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}

			AlphanumericValidator().ValidateString(context.Background(), req, resp)

			if tt.wantErr && !resp.Diagnostics.HasError() {
				t.Errorf("expected error for %q but got none", tt.value)
			}
			if !tt.wantErr && resp.Diagnostics.HasError() {
				t.Errorf("expected no error for %q but got: %s", tt.value, resp.Diagnostics.Errors())
			}
		})
	}
}

func TestHostnameNoDotValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"with hyphen", "my-host", false},
		{"with underscore and digits", "host_01", false},
		{"simple", "host", false},
		{"uppercase", "MyHost", false},
		{"contains dot", "my.host.com", true},
		{"single dot", "host.local", true},
		{"empty string", "", true},
		{"contains space", "my host", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: types.StringValue(tt.value),
			}
			resp := &validator.StringResponse{}

			HostnameNoDotValidator().ValidateString(context.Background(), req, resp)

			if tt.wantErr && !resp.Diagnostics.HasError() {
				t.Errorf("expected error for %q but got none", tt.value)
			}
			if !tt.wantErr && resp.Diagnostics.HasError() {
				t.Errorf("expected no error for %q but got: %s", tt.value, resp.Diagnostics.Errors())
			}
		})
	}
}

func TestLDAPURIValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		values   []string
		wantErrs int
	}{
		{"all ldap", []string{"ldap://a", "ldap://b"}, 0},
		{"all ldaps", []string{"ldaps://a:636", "ldaps://b:636"}, 0},
		{"mixed ldap ldaps", []string{"ldap://a", "ldaps://b"}, 0},
		{"http rejected", []string{"http://a"}, 1},
		{"one bad in middle", []string{"ldap://a", "ftp://b", "ldaps://c"}, 1},
		{"two bad", []string{"http://a", "https://b"}, 2},
		{"empty list", []string{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			elems := make([]attr.Value, len(tt.values))
			for i, v := range tt.values {
				elems[i] = types.StringValue(v)
			}
			lv, d := types.ListValue(types.StringType, elems)
			if d.HasError() {
				t.Fatalf("ListValue: %s", d)
			}
			req := validator.ListRequest{
				Path:        path.Root("uris"),
				ConfigValue: lv,
			}
			resp := &validator.ListResponse{}
			LDAPURIValidator().ValidateList(context.Background(), req, resp)
			if got := len(resp.Diagnostics.Errors()); got != tt.wantErrs {
				t.Errorf("want %d errors, got %d: %s", tt.wantErrs, got, resp.Diagnostics.Errors())
			}
		})
	}
}

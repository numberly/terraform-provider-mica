package provider

import (
	"context"
	"testing"

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

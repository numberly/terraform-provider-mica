package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// alphanumericStringValidator rejects any string not matching ^[a-zA-Z0-9]+$.
type alphanumericStringValidator struct{}

// AlphanumericValidator returns a validator that ensures the string value
// contains only alphanumeric characters (a-z, A-Z, 0-9).
func AlphanumericValidator() validator.String {
	return alphanumericStringValidator{}
}

func (v alphanumericStringValidator) Description(_ context.Context) string {
	return "value must contain only alphanumeric characters (a-z, A-Z, 0-9)"
}

func (v alphanumericStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v alphanumericStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			"value must contain only alphanumeric characters (a-z, A-Z, 0-9)",
		)
	}
}

// hostnameNoDotStringValidator rejects any string containing a dot.
// Allowed characters: alphanumeric, hyphen, underscore.
type hostnameNoDotStringValidator struct{}

// HostnameNoDotValidator returns a validator that ensures the string value
// contains only alphanumeric characters, hyphens, and underscores (no dots).
func HostnameNoDotValidator() validator.String {
	return hostnameNoDotStringValidator{}
}

func (v hostnameNoDotStringValidator) Description(_ context.Context) string {
	return "value must contain only alphanumeric characters, hyphens, and underscores (no dots)"
}

func (v hostnameNoDotStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v hostnameNoDotStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			"value must contain only alphanumeric characters, hyphens, and underscores (no dots)",
		)
	}
}

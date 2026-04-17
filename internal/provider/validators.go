package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)


// Pre-compiled regex patterns used by validators.
// Compiled once at package init time to avoid per-call overhead.
var (
	reAlphanumeric  = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	reHostnameNoDot = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
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
	if !reAlphanumeric.MatchString(value) {
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
	if !reHostnameNoDot.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			"value must contain only alphanumeric characters, hyphens, and underscores (no dots)",
		)
	}
}

// ldapURIListValidator rejects any list element not starting with ldap:// or ldaps://.
type ldapURIListValidator struct{}

// LDAPURIValidator returns a list validator that ensures every string element
// starts with "ldap://" or "ldaps://". Used on the `uris` attribute of
// flashblade_directory_service_management.
func LDAPURIValidator() validator.List {
	return ldapURIListValidator{}
}

func (v ldapURIListValidator) Description(_ context.Context) string {
	return "each element must start with ldap:// or ldaps://"
}

func (v ldapURIListValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ldapURIListValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	elems := req.ConfigValue.Elements()
	for i, el := range elems {
		sv, ok := el.(types.String)
		if !ok {
			continue
		}
		if sv.IsNull() || sv.IsUnknown() {
			continue
		}
		val := sv.ValueString()
		if !strings.HasPrefix(val, "ldap://") && !strings.HasPrefix(val, "ldaps://") {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(i),
				"Invalid Value",
				fmt.Sprintf("uris[%d] must start with ldap:// or ldaps://", i),
			)
		}
	}
}

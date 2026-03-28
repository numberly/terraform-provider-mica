package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// listToStrings extracts a []string from a types.List of string elements.
// Null and unknown lists are treated as empty slices.
func listToStrings(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return []string{}, diags
	}
	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)
	return result, diags
}

// emptyStringList returns a types.List with zero string elements.
func emptyStringList() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{})
}

// compositeID joins parts with "/" to form a composite Terraform import ID.
func compositeID(parts ...string) string {
	return strings.Join(parts, "/")
}

// parseCompositeID splits a composite import ID into exactly n parts.
// Uses strings.SplitN with "/" separator to preserve embedded separators in the last part.
// Returns a descriptive error if the part count does not match.
func parseCompositeID(id string, n int) ([]string, error) {
	parts := strings.SplitN(id, "/", n)
	if len(parts) != n {
		return nil, fmt.Errorf("expected %d parts separated by '/', got %d in %q", n, len(parts), id)
	}
	return parts, nil
}

// stringOrNull maps empty Go strings to types.StringNull() and non-empty strings to types.StringValue().
// Use this for API fields that return JSON null (decoded as "") when unset,
// so Terraform state reflects null rather than an empty string.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

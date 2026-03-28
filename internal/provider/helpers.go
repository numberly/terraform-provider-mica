package provider

import (
	"context"

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

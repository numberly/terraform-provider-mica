package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
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

// doublePointerRefForPatch encodes the three-state PATCH semantics for
// clearable reference fields (CONVENTIONS.md §Pointer rules).
//
// Returns:
//   - nil                                          when state == plan (OMIT: do not send)
//   - &(*NamedReference)(nil)                      when plan is null and state was set (CLEAR: send JSON null)
//   - &(&NamedReference{Name: plan.ValueString()}) when plan has a new value (SET: send value)
//
// Use this helper from Update() to build PATCH bodies with `**NamedReference`
// fields so that users can clear, set, or leave untouched any reference.
func doublePointerRefForPatch(state, plan types.String) **client.NamedReference {
	if state.Equal(plan) {
		return nil
	}
	if plan.IsNull() {
		var null *client.NamedReference
		return &null
	}
	ref := &client.NamedReference{Name: plan.ValueString()}
	return &ref
}

// ---------- shared helpers (DUP-01 through DUP-05) -------------------------

// spaceAttrTypes returns the attribute type map for the shared "space" nested object
// used by filesystem, bucket, and object store account resources.
func spaceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"data_reduction":      types.Float64Type,
		"snapshots":           types.Int64Type,
		"total_physical":      types.Int64Type,
		"unique":              types.Int64Type,
		"virtual":             types.Int64Type,
		"snapshots_effective": types.Int64Type,
	}
}

// mapSpaceToObject builds a types.Object from a client.Space struct.
// Returns the object and diagnostics (never panics).
func mapSpaceToObject(space client.Space) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(spaceAttrTypes(), map[string]attr.Value{
		"data_reduction":      types.Float64Value(space.DataReduction),
		"snapshots":           types.Int64Value(space.Snapshots),
		"total_physical":      types.Int64Value(space.TotalPhysical),
		"unique":              types.Int64Value(space.Unique),
		"virtual":             types.Int64Value(space.Virtual),
		"snapshots_effective": types.Int64Value(space.SnapshotsEffective),
	})
}

// nullTimeoutsValue returns a timeouts.Value initialized with a null Object
// containing the standard CRUD timeout attribute types. Use in ImportState methods
// to initialize the Timeouts field before reading from the API.
func nullTimeoutsValue() timeouts.Value {
	return timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}
}

// nullTimeoutsValueNoUpdate returns a timeouts.Value initialized with a null Object
// containing only Create, Read, and Delete timeout attribute types. Use in ImportState
// methods for CRD-only resources (no Update operation).
func nullTimeoutsValueNoUpdate() timeouts.Value {
	return timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}
}

// namedRefsToNames extracts the Name field from each NamedReference.
func namedRefsToNames(refs []client.NamedReference) []string {
	if len(refs) == 0 {
		return []string{}
	}
	names := make([]string, len(refs))
	for i, ref := range refs {
		names[i] = ref.Name
	}
	return names
}

// namedRefsToListValue converts a slice of NamedReference to a types.List of
// their Name values. Returns an empty list (not null) when refs is empty.
func namedRefsToListValue(refs []client.NamedReference) types.List {
	return stringsToListValue(namedRefsToNames(refs))
}

// stringsToListValue converts a []string to a types.List of string values.
// Returns an empty list (not null) when the slice is empty.
func stringsToListValue(values []string) types.List {
	vals := make([]attr.Value, len(values))
	for i, v := range values {
		vals[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, vals)
}

// stringSlicesEqual returns true if two string slices have the same elements in the same order.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}


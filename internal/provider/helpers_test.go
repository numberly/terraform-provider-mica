package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

func TestUnit_CompositeID(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{"two parts", []string{"my-policy", "rule-abc"}, "my-policy/rule-abc"},
		{"single part", []string{"a"}, "a"},
		{"three parts", []string{"a", "b", "c"}, "a/b/c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compositeID(tt.parts...)
			if got != tt.want {
				t.Errorf("compositeID(%v) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}

func TestUnit_ParseCompositeID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		n       int
		want    []string
		wantErr bool
	}{
		{"valid two parts", "my-policy/0", 2, []string{"my-policy", "0"}, false},
		{"missing separator", "my-policy", 2, nil, true},
		{"preserves embedded separator", "a/b/c", 2, []string{"a", "b/c"}, false},
		{"empty string", "", 2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCompositeID(tt.id, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCompositeID(%q, %d) error = %v, wantErr %v", tt.id, tt.n, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseCompositeID(%q, %d) = %v, want %v", tt.id, tt.n, got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("parseCompositeID(%q, %d)[%d] = %q, want %q", tt.id, tt.n, i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestUnit_StringOrNull(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want types.String
	}{
		{"empty returns null", "", types.StringNull()},
		{"non-empty returns value", "allow", types.StringValue("allow")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringOrNull(tt.s)
			if !got.Equal(tt.want) {
				t.Errorf("stringOrNull(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// TestUnit_Helpers_DoublePointerRefForPatch_Omit_SameValue verifies that identical
// state and plan values produce a nil outer pointer (OMIT).
func TestUnit_Helpers_DoublePointerRefForPatch_Omit_SameValue(t *testing.T) {
	state := types.StringValue("foo")
	plan := types.StringValue("foo")
	got := doublePointerRefForPatch(state, plan)
	if got != nil {
		t.Fatalf("expected nil (omit) when state == plan, got %#v", got)
	}
}

// TestUnit_Helpers_DoublePointerRefForPatch_Omit_BothNull verifies that two null
// values collapse to OMIT (no spurious PATCH field).
func TestUnit_Helpers_DoublePointerRefForPatch_Omit_BothNull(t *testing.T) {
	state := types.StringNull()
	plan := types.StringNull()
	got := doublePointerRefForPatch(state, plan)
	if got != nil {
		t.Fatalf("expected nil (omit) when both null, got %#v", got)
	}
}

// TestUnit_Helpers_DoublePointerRefForPatch_Clear verifies that transitioning
// from a set state to a null plan produces the CLEAR encoding
// (outer non-nil, inner nil).
func TestUnit_Helpers_DoublePointerRefForPatch_Clear(t *testing.T) {
	state := types.StringValue("foo")
	plan := types.StringNull()
	got := doublePointerRefForPatch(state, plan)
	if got == nil {
		t.Fatal("expected non-nil outer pointer for CLEAR, got nil")
	}
	if *got != nil {
		t.Fatalf("expected inner nil for CLEAR, got %#v", *got)
	}
}

// TestUnit_Helpers_DoublePointerRefForPatch_Set_FromNull verifies that
// transitioning from null state to a set plan produces the SET encoding
// with the plan's name value.
func TestUnit_Helpers_DoublePointerRefForPatch_Set_FromNull(t *testing.T) {
	state := types.StringNull()
	plan := types.StringValue("foo")
	got := doublePointerRefForPatch(state, plan)
	if got == nil || *got == nil {
		t.Fatalf("expected non-nil outer and inner for SET, got %#v", got)
	}
	if (*got).Name != "foo" {
		t.Errorf("expected inner.Name=%q, got %q", "foo", (*got).Name)
	}
	// Ensure client package is referenced to avoid unused import in some layouts.
	var _ = (*client.NamedReference)(*got)
}

// TestUnit_Helpers_DoublePointerRefForPatch_Set_Changed verifies that changing
// from one non-null value to another produces SET with the new value.
func TestUnit_Helpers_DoublePointerRefForPatch_Set_Changed(t *testing.T) {
	state := types.StringValue("foo")
	plan := types.StringValue("bar")
	got := doublePointerRefForPatch(state, plan)
	if got == nil || *got == nil {
		t.Fatalf("expected non-nil outer and inner for SET, got %#v", got)
	}
	if (*got).Name != "bar" {
		t.Errorf("expected inner.Name=%q, got %q", "bar", (*got).Name)
	}
}

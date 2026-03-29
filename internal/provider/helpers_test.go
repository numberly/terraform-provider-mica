package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCompositeID(t *testing.T) {
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

func TestParseCompositeID(t *testing.T) {
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

func TestInt64UseStateForUnknown(t *testing.T) {
	ctx := context.Background()
	mod := int64UseStateForUnknown()

	t.Run("preserves state when plan is unknown", func(t *testing.T) {
		req := planmodifier.Int64Request{
			PlanValue:  types.Int64Unknown(),
			StateValue: types.Int64Value(42),
		}
		resp := &planmodifier.Int64Response{PlanValue: req.PlanValue}
		mod.PlanModifyInt64(ctx, req, resp)
		if !resp.PlanValue.Equal(types.Int64Value(42)) {
			t.Errorf("expected 42, got %v", resp.PlanValue)
		}
	})

	t.Run("does nothing when plan is not unknown", func(t *testing.T) {
		req := planmodifier.Int64Request{
			PlanValue:  types.Int64Value(10),
			StateValue: types.Int64Value(42),
		}
		resp := &planmodifier.Int64Response{PlanValue: req.PlanValue}
		mod.PlanModifyInt64(ctx, req, resp)
		if !resp.PlanValue.Equal(types.Int64Value(10)) {
			t.Errorf("expected 10, got %v", resp.PlanValue)
		}
	})

	t.Run("does nothing when state is null", func(t *testing.T) {
		req := planmodifier.Int64Request{
			PlanValue:  types.Int64Unknown(),
			StateValue: types.Int64Null(),
		}
		resp := &planmodifier.Int64Response{PlanValue: req.PlanValue}
		mod.PlanModifyInt64(ctx, req, resp)
		if !resp.PlanValue.IsUnknown() {
			t.Errorf("expected unknown, got %v", resp.PlanValue)
		}
	})
}

func TestFloat64UseStateForUnknown(t *testing.T) {
	ctx := context.Background()
	mod := float64UseStateForUnknown()

	t.Run("preserves state when plan is unknown", func(t *testing.T) {
		req := planmodifier.Float64Request{
			PlanValue:  types.Float64Unknown(),
			StateValue: types.Float64Value(3.14),
		}
		resp := &planmodifier.Float64Response{PlanValue: req.PlanValue}
		mod.PlanModifyFloat64(ctx, req, resp)
		if !resp.PlanValue.Equal(types.Float64Value(3.14)) {
			t.Errorf("expected 3.14, got %v", resp.PlanValue)
		}
	})

	t.Run("does nothing when plan is not unknown", func(t *testing.T) {
		req := planmodifier.Float64Request{
			PlanValue:  types.Float64Value(1.0),
			StateValue: types.Float64Value(3.14),
		}
		resp := &planmodifier.Float64Response{PlanValue: req.PlanValue}
		mod.PlanModifyFloat64(ctx, req, resp)
		if !resp.PlanValue.Equal(types.Float64Value(1.0)) {
			t.Errorf("expected 1.0, got %v", resp.PlanValue)
		}
	})

	t.Run("does nothing when state is null", func(t *testing.T) {
		req := planmodifier.Float64Request{
			PlanValue:  types.Float64Unknown(),
			StateValue: types.Float64Null(),
		}
		resp := &planmodifier.Float64Response{PlanValue: req.PlanValue}
		mod.PlanModifyFloat64(ctx, req, resp)
		if !resp.PlanValue.IsUnknown() {
			t.Errorf("expected unknown, got %v", resp.PlanValue)
		}
	})
}

func TestStringOrNull(t *testing.T) {
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

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
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

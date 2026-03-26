//go:build generate

package tools

import (
	// tfplugindocs is used via go generate to produce registry-compatible documentation.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

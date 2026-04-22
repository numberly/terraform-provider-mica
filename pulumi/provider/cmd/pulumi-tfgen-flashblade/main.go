package main

import (
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfgen"

	flashblade "github.com/numberly/opentofu-provider-flashblade/pulumi/provider"
)

func main() {
	// PF tfgen does not take a version parameter (differs from SDK v2 bridge).
	// Version is injected into the runtime binary via ldflags -X.
	tfgen.Main("flashblade", flashblade.Provider())
}

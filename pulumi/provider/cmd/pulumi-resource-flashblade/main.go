package main

import (
	"context"
	_ "embed"

	pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"

	flashblade "github.com/numberly/opentofu-provider-flashblade/pulumi/provider"
)

//go:embed schema-embed.json
var pulumiSchema []byte

//go:embed bridge-metadata.json
var bridgeMetadata []byte

func main() {
	meta := pftfbridge.ProviderMetadata{
		PackageSchema:  pulumiSchema,
		BridgeMetadata: bridgeMetadata,
	}
	pftfbridge.Main(
		context.Background(),
		"flashblade",
		flashblade.Provider(),
		meta,
	)
}

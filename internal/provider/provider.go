package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure FlashBladeProvider satisfies the provider.Provider interface.
var _ provider.Provider = &FlashBladeProvider{}

// FlashBladeProvider is the root provider struct.
type FlashBladeProvider struct {
	version string
}

// New returns a factory function that creates a FlashBladeProvider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FlashBladeProvider{version: version}
	}
}

// Metadata sets the provider type name.
func (p *FlashBladeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "flashblade"
	resp.Version = p.version
}

// Schema returns an empty schema placeholder — fleshed out in Plan 02.
func (p *FlashBladeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Configure is a no-op stub — implementation in Plan 02.
func (p *FlashBladeProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

// Resources returns an empty slice — resources added in Plan 03+.
func (p *FlashBladeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

// DataSources returns an empty slice — data sources added in Plan 03+.
func (p *FlashBladeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

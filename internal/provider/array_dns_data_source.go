package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure arrayDnsDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &arrayDnsDataSource{}
var _ datasource.DataSourceWithConfigure = &arrayDnsDataSource{}

// arrayDnsDataSource implements the flashblade_array_dns data source.
type arrayDnsDataSource struct {
	client *client.FlashBladeClient
}

// NewArrayDnsDataSource is the factory function registered in the provider.
func NewArrayDnsDataSource() datasource.DataSource {
	return &arrayDnsDataSource{}
}

// ---------- model structs ----------------------------------------------------

// arrayDnsDataSourceModel is the top-level model for the flashblade_array_dns data source.
type arrayDnsDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.List   `tfsdk:"nameservers"`
	Services    types.List   `tfsdk:"services"`
	Sources     types.List   `tfsdk:"sources"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *arrayDnsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_array_dns"
}

// Schema defines the data source schema.
func (d *arrayDnsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a DNS configuration entry from a FlashBlade array by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the DNS configuration.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the DNS configuration to look up.",
			},
			"domain": schema.StringAttribute{
				Computed:    true,
				Description: "The domain suffix appended by the array to unqualified hostnames.",
			},
			"nameservers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of DNS server IP addresses.",
			},
			"services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Services that use this DNS configuration.",
			},
			"sources": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Network interfaces used for DNS traffic.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *arrayDnsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.FlashBladeClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.FlashBladeClient, got: %T. This is a bug in the provider.", req.ProviderData),
		)
		return
	}
	d.client = c
}

// Read fetches the DNS configuration by name and populates state.
func (d *arrayDnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config arrayDnsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	dns, err := d.client.GetArrayDns(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError("DNS configuration not found", fmt.Sprintf("No DNS configuration named %q", name))
		} else {
			resp.Diagnostics.AddError("Error reading array DNS configuration", err.Error())
		}
		return
	}

	config.ID = types.StringValue(dns.ID)
	config.Name = types.StringValue(dns.Name)
	config.Domain = types.StringValue(dns.Domain)

	if len(dns.Nameservers) > 0 {
		ns, diags := types.ListValueFrom(ctx, types.StringType, dns.Nameservers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.Nameservers = ns
	} else {
		config.Nameservers = emptyStringList()
	}

	if len(dns.Services) > 0 {
		svc, diags := types.ListValueFrom(ctx, types.StringType, dns.Services)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.Services = svc
	} else {
		config.Services = emptyStringList()
	}

	if len(dns.Sources) > 0 {
		src, diags := types.ListValueFrom(ctx, types.StringType, dns.Sources)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.Sources = src
	} else {
		config.Sources = emptyStringList()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

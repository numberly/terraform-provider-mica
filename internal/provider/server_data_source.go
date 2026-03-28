package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure serverDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &serverDataSource{}
var _ datasource.DataSourceWithConfigure = &serverDataSource{}

// serverDataSource implements the flashblade_server data source.
type serverDataSource struct {
	client *client.FlashBladeClient
}

// NewServerDataSource is the factory function registered in the provider.
func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

// ---------- model structs ----------------------------------------------------

// serverDataSourceModel is the top-level model for the flashblade_server data source.
type serverDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Created types.Int64  `tfsdk:"created"`
	DNS     types.List   `tfsdk:"dns"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *serverDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_server"
}

// Schema defines the data source schema.
func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade server by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the server.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the server to look up.",
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the server was created.",
			},
			"dns": schema.ListNestedAttribute{
				Computed:    true,
				Description: "DNS configuration for the server.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Computed:    true,
							Description: "DNS domain suffix.",
						},
						"nameservers": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of DNS nameserver IP addresses.",
						},
						"services": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of DNS service types.",
						},
					},
				},
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches server data by name and populates state.
func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	srv, err := d.client.GetServer(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Server not found",
				fmt.Sprintf("No server with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading server", err.Error())
		return
	}

	config.ID = types.StringValue(srv.ID)
	config.Name = types.StringValue(srv.Name)
	config.Created = types.Int64Value(srv.Created)

	mapServerDNSToDataSourceModel(ctx, srv, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapServerDNSToDataSourceModel maps DNS from client.Server to the data source model.
func mapServerDNSToDataSourceModel(ctx context.Context, srv *client.Server, data *serverDataSourceModel, diags *diag.Diagnostics) {
	if len(srv.DNS) > 0 {
		dnsObjs := make([]attr.Value, 0, len(srv.DNS))
		for _, d := range srv.DNS {
			var nameservers types.List
			if len(d.Nameservers) > 0 {
				ns, nsDiags := types.ListValueFrom(ctx, types.StringType, d.Nameservers)
				diags.Append(nsDiags...)
				if diags.HasError() {
					return
				}
				nameservers = ns
			} else {
				nameservers = types.ListNull(types.StringType)
			}

			var services types.List
			if len(d.Services) > 0 {
				svc, svcDiags := types.ListValueFrom(ctx, types.StringType, d.Services)
				diags.Append(svcDiags...)
				if diags.HasError() {
					return
				}
				services = svc
			} else {
				services = types.ListNull(types.StringType)
			}

			obj, objDiags := types.ObjectValue(serverDNSAttrTypes(), map[string]attr.Value{
				"domain":      types.StringValue(d.Domain),
				"nameservers": nameservers,
				"services":    services,
			})
			diags.Append(objDiags...)
			if diags.HasError() {
				return
			}
			dnsObjs = append(dnsObjs, obj)
		}

		dnsList, listDiags := types.ListValue(serverDNSObjectType(), dnsObjs)
		diags.Append(listDiags...)
		if diags.HasError() {
			return
		}
		data.DNS = dnsList
	} else {
		data.DNS = types.ListNull(serverDNSObjectType())
	}
}

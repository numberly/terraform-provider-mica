package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &syslogServerDataSource{}
var _ datasource.DataSourceWithConfigure = &syslogServerDataSource{}

// syslogServerDataSource implements the flashblade_syslog_server data source.
type syslogServerDataSource struct {
	client *client.FlashBladeClient
}

func NewSyslogServerDataSource() datasource.DataSource {
	return &syslogServerDataSource{}
}

// ---------- model structs ----------------------------------------------------

// syslogServerDataSourceModel is the top-level model for the data source.
type syslogServerDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	URI      types.String `tfsdk:"uri"`
	Services types.List   `tfsdk:"services"`
	Sources  types.List   `tfsdk:"sources"`
}

// ---------- data source interface methods -----------------------------------

func (d *syslogServerDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_syslog_server"
}

// Schema defines the data source schema.
func (d *syslogServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade syslog server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the syslog server.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the syslog server to look up.",
			},
			"uri": schema.StringAttribute{
				Computed:    true,
				Description: "Syslog server URI in format PROTOCOL://HOST:PORT.",
			},
			"services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of services configured for this syslog server.",
			},
			"sources": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of sources configured for this syslog server.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *syslogServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches syslog server data and populates state.
func (d *syslogServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config syslogServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	srv, err := d.client.GetSyslogServer(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Syslog server not found",
				fmt.Sprintf("No syslog server with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading syslog server", err.Error())
		return
	}

	config.ID = types.StringValue(srv.ID)
	config.Name = types.StringValue(srv.Name)
	config.URI = types.StringValue(srv.URI)

	// Map Services.
	if len(srv.Services) > 0 {
		vals := make([]attr.Value, len(srv.Services))
		for i, s := range srv.Services {
			vals[i] = types.StringValue(s)
		}
		config.Services = types.ListValueMust(types.StringType, vals)
	} else {
		config.Services = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Map Sources.
	if len(srv.Sources) > 0 {
		vals := make([]attr.Value, len(srv.Sources))
		for i, s := range srv.Sources {
			vals[i] = types.StringValue(s)
		}
		config.Sources = types.ListValueMust(types.StringType, vals)
	} else {
		config.Sources = types.ListValueMust(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

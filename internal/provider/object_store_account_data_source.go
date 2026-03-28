package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure objectStoreAccountDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &objectStoreAccountDataSource{}
var _ datasource.DataSourceWithConfigure = &objectStoreAccountDataSource{}

// objectStoreAccountDataSource implements the flashblade_object_store_account data source.
type objectStoreAccountDataSource struct {
	client *client.FlashBladeClient
}

// NewObjectStoreAccountDataSource is the factory function registered in the provider.
func NewObjectStoreAccountDataSource() datasource.DataSource {
	return &objectStoreAccountDataSource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccountDataSourceSpaceModel maps space attributes for the data source.
type objectStoreAccountDataSourceSpaceModel struct {
	DataReduction      types.Float64 `tfsdk:"data_reduction"`
	Snapshots          types.Int64   `tfsdk:"snapshots"`
	TotalPhysical      types.Int64   `tfsdk:"total_physical"`
	Unique             types.Int64   `tfsdk:"unique"`
	Virtual            types.Int64   `tfsdk:"virtual"`
	SnapshotsEffective types.Int64   `tfsdk:"snapshots_effective"`
}

// objectStoreAccountDataSourceModel is the top-level model for the flashblade_object_store_account data source.
type objectStoreAccountDataSourceModel struct {
	ID               types.String                            `tfsdk:"id"`
	Name             types.String                            `tfsdk:"name"`
	Created          types.Int64                             `tfsdk:"created"`
	QuotaLimit       types.String                            `tfsdk:"quota_limit"`
	HardLimitEnabled types.Bool                              `tfsdk:"hard_limit_enabled"`
	ObjectCount      types.Int64                             `tfsdk:"object_count"`
	Space            *objectStoreAccountDataSourceSpaceModel `tfsdk:"space"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *objectStoreAccountDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_account"
}

// Schema defines the data source schema.
func (d *objectStoreAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade object store account by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store account.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store account to look up.",
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the account was created.",
			},
			"quota_limit": schema.StringAttribute{
				Computed:    true,
				Description: "The effective quota limit applied against the size of the account, in bytes.",
			},
			"hard_limit_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "If true, the account's size cannot exceed the quota limit.",
			},
			"object_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The count of objects within the account.",
			},
			"space": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Storage space breakdown.",
				Attributes: map[string]schema.Attribute{
					"data_reduction": schema.Float64Attribute{
						Computed:    true,
						Description: "Data reduction ratio.",
					},
					"snapshots": schema.Int64Attribute{
						Computed:    true,
						Description: "Physical space used by snapshots in bytes.",
					},
					"total_physical": schema.Int64Attribute{
						Computed:    true,
						Description: "Total physical space used in bytes.",
					},
					"unique": schema.Int64Attribute{
						Computed:    true,
						Description: "Unique physical space used in bytes.",
					},
					"virtual": schema.Int64Attribute{
						Computed:    true,
						Description: "Virtual (logical) space used in bytes.",
					},
					"snapshots_effective": schema.Int64Attribute{
						Computed:    true,
						Description: "Effective snapshot space used in bytes.",
					},
				},
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *objectStoreAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches object store account data by name and populates state.
func (d *objectStoreAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectStoreAccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	acct, err := d.client.GetObjectStoreAccount(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Object store account not found",
				fmt.Sprintf("No object store account with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading object store account", err.Error())
		return
	}

	config.ID = types.StringValue(acct.ID)
	config.Name = types.StringValue(acct.Name)
	config.Created = types.Int64Value(acct.Created)
	config.QuotaLimit = types.StringValue(acct.QuotaLimit)
	config.HardLimitEnabled = types.BoolValue(acct.HardLimitEnabled)
	config.ObjectCount = types.Int64Value(acct.ObjectCount)

	config.Space = &objectStoreAccountDataSourceSpaceModel{
		DataReduction:      types.Float64Value(acct.Space.DataReduction),
		Snapshots:          types.Int64Value(acct.Space.Snapshots),
		TotalPhysical:      types.Int64Value(acct.Space.TotalPhysical),
		Unique:             types.Int64Value(acct.Space.Unique),
		Virtual:            types.Int64Value(acct.Space.Virtual),
		SnapshotsEffective: types.Int64Value(acct.Space.SnapshotsEffective),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

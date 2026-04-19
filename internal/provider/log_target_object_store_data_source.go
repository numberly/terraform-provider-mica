package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &logTargetObjectStoreDataSource{}
var _ datasource.DataSourceWithConfigure = &logTargetObjectStoreDataSource{}

// logTargetObjectStoreDataSource implements the flashblade_log_target_object_store data source.
type logTargetObjectStoreDataSource struct {
	client *client.FlashBladeClient
}

func NewLogTargetObjectStoreDataSource() datasource.DataSource {
	return &logTargetObjectStoreDataSource{}
}

// ---------- model structs ----------------------------------------------------

// logTargetObjectStoreDataSourceModel is the top-level model for the data source.
type logTargetObjectStoreDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	BucketName        types.String `tfsdk:"bucket_name"`
	LogNamePrefix     types.String `tfsdk:"log_name_prefix"`
	LogRotateDuration types.Int64  `tfsdk:"log_rotate_duration"`
}

// ---------- data source interface methods -----------------------------------

func (d *logTargetObjectStoreDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_log_target_object_store"
}

// Schema defines the data source schema.
func (d *logTargetObjectStoreDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade log target object store.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the log target object store.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the log target object store to look up.",
			},
			"bucket_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the bucket where audit logs are stored.",
			},
			"log_name_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "The prefix of audit log object names in the bucket.",
			},
			"log_rotate_duration": schema.Int64Attribute{
				Computed:    true,
				Description: "The rotation interval for audit logs in milliseconds.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *logTargetObjectStoreDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches log target object store data and populates state.
func (d *logTargetObjectStoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config logTargetObjectStoreDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	item, err := d.client.GetLogTargetObjectStore(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Log target object store not found",
				fmt.Sprintf("No log target object store with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading log target object store", err.Error())
		return
	}

	config.ID = types.StringValue(item.ID)
	config.Name = types.StringValue(item.Name)
	config.BucketName = types.StringValue(item.Bucket.Name)
	config.LogNamePrefix = types.StringValue(item.LogNamePrefix.Prefix)
	config.LogRotateDuration = types.Int64Value(item.LogRotate.Duration)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

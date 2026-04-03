package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure bucketReplicaLinkDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &bucketReplicaLinkDataSource{}
var _ datasource.DataSourceWithConfigure = &bucketReplicaLinkDataSource{}

// bucketReplicaLinkDataSource implements the flashblade_bucket_replica_link data source.
type bucketReplicaLinkDataSource struct {
	client *client.FlashBladeClient
}

// NewBucketReplicaLinkDataSource is the factory function registered in the provider.
func NewBucketReplicaLinkDataSource() datasource.DataSource {
	return &bucketReplicaLinkDataSource{}
}

// ---------- model structs ----------------------------------------------------

// bucketReplicaLinkDataSourceModel is the model for the flashblade_bucket_replica_link data source.
type bucketReplicaLinkDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	LocalBucketName        types.String `tfsdk:"local_bucket_name"`
	RemoteBucketName       types.String `tfsdk:"remote_bucket_name"`
	RemoteCredentialsName  types.String `tfsdk:"remote_credentials_name"`
	RemoteName             types.String `tfsdk:"remote_name"`
	Paused                 types.Bool   `tfsdk:"paused"`
	CascadingEnabled       types.Bool   `tfsdk:"cascading_enabled"`
	Direction              types.String `tfsdk:"direction"`
	Status                 types.String `tfsdk:"status"`
	StatusDetails          types.String `tfsdk:"status_details"`
	Lag                    types.Int64  `tfsdk:"lag"`
	RecoveryPoint          types.Int64  `tfsdk:"recovery_point"`
	ObjectBacklogCount     types.Int64  `tfsdk:"object_backlog_count"`
	ObjectBacklogTotalSize types.Int64  `tfsdk:"object_backlog_total_size"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *bucketReplicaLinkDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_replica_link"
}

// Schema defines the data source schema.
func (d *bucketReplicaLinkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade bucket replica link by local and remote bucket names.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the bucket replica link. Use to look up a specific link unambiguously.",
			},
			"local_bucket_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the local bucket. Required when looking up by bucket names.",
			},
			"remote_bucket_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the remote bucket. Required when looking up by bucket names.",
			},
			"remote_credentials_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the remote credentials. Use to disambiguate when multiple links exist for the same bucket pair.",
			},
			"remote_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the remote array connection.",
			},
			"paused": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the replica link is paused.",
			},
			"cascading_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether cascading replication is enabled.",
			},
			"direction": schema.StringAttribute{
				Computed:    true,
				Description: "The replication direction.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The replication status.",
			},
			"status_details": schema.StringAttribute{
				Computed:    true,
				Description: "Additional status details.",
			},
			"lag": schema.Int64Attribute{
				Computed:    true,
				Description: "Replication lag in milliseconds.",
			},
			"recovery_point": schema.Int64Attribute{
				Computed:    true,
				Description: "Recovery point timestamp in milliseconds.",
			},
			"object_backlog_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of objects in the replication backlog.",
			},
			"object_backlog_total_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Total size of objects in the replication backlog in bytes.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *bucketReplicaLinkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches a bucket replica link by local and remote bucket names and populates state.
func (d *bucketReplicaLinkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bucketReplicaLinkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one lookup method is provided.
	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasBucketNames := !config.LocalBucketName.IsNull() && !config.RemoteBucketName.IsNull()
	if !hasID && !hasBucketNames {
		resp.Diagnostics.AddError("Missing lookup criteria",
			"Either id or both local_bucket_name and remote_bucket_name must be specified.")
		return
	}

	// If an ID is provided, use it directly (unambiguous).
	// Otherwise fall back to bucket names + optional remote_credentials_name filter.
	var link *client.BucketReplicaLink
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		var err error
		link, err = d.client.GetBucketReplicaLinkByID(ctx, config.ID.ValueString())
		if err != nil {
			if client.IsNotFound(err) {
				resp.Diagnostics.AddError("Bucket replica link not found",
					fmt.Sprintf("No bucket replica link with ID %q exists.", config.ID.ValueString()))
				return
			}
			resp.Diagnostics.AddError("Error reading bucket replica link", err.Error())
			return
		}
	} else {
		links, err := d.client.ListBucketReplicaLinks(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error listing bucket replica links", err.Error())
			return
		}

		localName := config.LocalBucketName.ValueString()
		remoteName := config.RemoteBucketName.ValueString()
		rcFilter := ""
		if !config.RemoteCredentialsName.IsNull() {
			rcFilter = config.RemoteCredentialsName.ValueString()
		}

		var matches []client.BucketReplicaLink
		for _, l := range links {
			if l.LocalBucket.Name != localName || l.RemoteBucket.Name != remoteName {
				continue
			}
			if rcFilter != "" {
				if l.RemoteCredentials == nil || l.RemoteCredentials.Name != rcFilter {
					continue
				}
			}
			matches = append(matches, l)
		}

		if len(matches) == 0 {
			resp.Diagnostics.AddError("Bucket replica link not found",
				fmt.Sprintf("No bucket replica link from %q to %q exists on the FlashBlade array.", localName, remoteName))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple bucket replica links found",
				fmt.Sprintf("Found %d bucket replica links from %q to %q. Set remote_credentials_name or id to disambiguate.", len(matches), localName, remoteName))
			return
		}
		link = &matches[0]
	}

	config.ID = types.StringValue(link.ID)
	config.LocalBucketName = types.StringValue(link.LocalBucket.Name)
	config.RemoteBucketName = types.StringValue(link.RemoteBucket.Name)
	config.RemoteName = stringOrNull(link.Remote.Name)
	config.Paused = types.BoolValue(link.Paused)
	config.CascadingEnabled = types.BoolValue(link.CascadingEnabled)
	config.Direction = stringOrNull(link.Direction)
	config.Status = stringOrNull(link.Status)
	config.StatusDetails = stringOrNull(link.StatusDetails)
	config.Lag = types.Int64Value(link.Lag)
	config.RecoveryPoint = types.Int64Value(link.RecoveryPoint)

	if link.RemoteCredentials != nil {
		config.RemoteCredentialsName = types.StringValue(link.RemoteCredentials.Name)
	} else {
		config.RemoteCredentialsName = types.StringNull()
	}

	if link.ObjectBacklog != nil {
		config.ObjectBacklogCount = types.Int64Value(link.ObjectBacklog.Count)
		config.ObjectBacklogTotalSize = types.Int64Value(link.ObjectBacklog.TotalSize)
	} else {
		config.ObjectBacklogCount = types.Int64Value(0)
		config.ObjectBacklogTotalSize = types.Int64Value(0)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure bucketReplicaLinkResource satisfies the resource interfaces.
var _ resource.Resource = &bucketReplicaLinkResource{}
var _ resource.ResourceWithConfigure = &bucketReplicaLinkResource{}
var _ resource.ResourceWithImportState = &bucketReplicaLinkResource{}
var _ resource.ResourceWithUpgradeState = &bucketReplicaLinkResource{}

// bucketReplicaLinkResource implements the flashblade_bucket_replica_link resource.
type bucketReplicaLinkResource struct {
	client *client.FlashBladeClient
}

// NewBucketReplicaLinkResource is the factory function registered in the provider.
func NewBucketReplicaLinkResource() resource.Resource {
	return &bucketReplicaLinkResource{}
}

// ---------- model structs ----------------------------------------------------

// bucketReplicaLinkModel is the Terraform state model for the flashblade_bucket_replica_link resource.
type bucketReplicaLinkModel struct {
	ID                    types.String   `tfsdk:"id"`
	LocalBucketName       types.String   `tfsdk:"local_bucket_name"`
	RemoteBucketName      types.String   `tfsdk:"remote_bucket_name"`
	RemoteCredentialsName types.String   `tfsdk:"remote_credentials_name"`
	RemoteName            types.String   `tfsdk:"remote_name"`
	Paused                types.Bool     `tfsdk:"paused"`
	CascadingEnabled      types.Bool     `tfsdk:"cascading_enabled"`
	Direction             types.String   `tfsdk:"direction"`
	Status                types.String   `tfsdk:"status"`
	StatusDetails         types.String   `tfsdk:"status_details"`
	Lag                   types.Int64    `tfsdk:"lag"`
	RecoveryPoint         types.Int64    `tfsdk:"recovery_point"`
	ObjectBacklogCount    types.Int64    `tfsdk:"object_backlog_count"`
	ObjectBacklogTotalSize types.Int64   `tfsdk:"object_backlog_total_size"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *bucketReplicaLinkResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_replica_link"
}

// Schema defines the resource schema.
func (r *bucketReplicaLinkResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade bucket replica link for cross-array bucket replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the bucket replica link.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"local_bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the local bucket. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remote_bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the remote bucket. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remote_credentials_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the remote credentials (for S3 replication targets). Omit for FlashBlade-to-FlashBlade replication.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"remote_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the remote array connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"paused": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the replica link is paused. Defaults to false.",
			},
			"cascading_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether cascading replication is enabled. Immutable after creation. Defaults to false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Computed:    true,
				Description: "The replication direction (e.g. 'outbound').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The replication status (e.g. 'replicating').",
			},
			"status_details": schema.StringAttribute{
				Computed:    true,
				Description: "Additional status details.",
			},
			"lag": schema.Int64Attribute{
				Computed:    true,
				Description: "Replication lag in milliseconds.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"recovery_point": schema.Int64Attribute{
				Computed:    true,
				Description: "Recovery point timestamp in milliseconds.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"object_backlog_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of objects in the replication backlog.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"object_backlog_total_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Total size of objects in the replication backlog in bytes.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}


// UpgradeState returns state upgraders for schema migrations.
func (r *bucketReplicaLinkResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *bucketReplicaLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = c
}

// ---------- CRUD methods ----------------------------------------------------

// Create creates a new bucket replica link.
func (r *bucketReplicaLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketReplicaLinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	body := client.BucketReplicaLinkPost{
		Paused:           data.Paused.ValueBool(),
		CascadingEnabled: data.CascadingEnabled.ValueBool(),
	}

	remoteCredentialsName := ""
	if !data.RemoteCredentialsName.IsNull() && !data.RemoteCredentialsName.IsUnknown() {
		remoteCredentialsName = data.RemoteCredentialsName.ValueString()
	}

	link, err := r.client.PostBucketReplicaLink(ctx, data.LocalBucketName.ValueString(), data.RemoteBucketName.ValueString(), remoteCredentialsName, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket replica link", err.Error())
		return
	}

	mapBucketReplicaLinkToModel(link, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *bucketReplicaLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketReplicaLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	link, err := r.client.GetBucketReplicaLink(ctx, data.LocalBucketName.ValueString(), data.RemoteBucketName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket replica link", err.Error())
		return
	}

	mapBucketReplicaLinkToModel(link, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing bucket replica link.
// Only paused is mutable via PATCH.
func (r *bucketReplicaLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state bucketReplicaLinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	if !plan.Paused.Equal(state.Paused) {
		paused := plan.Paused.ValueBool()
		patch := client.BucketReplicaLinkPatch{
			Paused: &paused,
		}
		_, err := r.client.PatchBucketReplicaLink(ctx, state.ID.ValueString(), patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating bucket replica link", err.Error())
			return
		}
	}

	// Re-read to refresh computed fields.
	link, err := r.client.GetBucketReplicaLink(ctx, plan.LocalBucketName.ValueString(), plan.RemoteBucketName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket replica link after update", err.Error())
		return
	}

	mapBucketReplicaLinkToModel(link, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a bucket replica link.
func (r *bucketReplicaLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketReplicaLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	err := r.client.DeleteBucketReplicaLink(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting bucket replica link", err.Error())
		return
	}
}

// ImportState imports an existing bucket replica link by composite ID "localBucket/remoteBucket".
func (r *bucketReplicaLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: localBucket/remoteBucket. Error: %s", err))
		return
	}

	link, err := r.client.GetBucketReplicaLink(ctx, parts[0], parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error importing bucket replica link", err.Error())
		return
	}

	var data bucketReplicaLinkModel
	data.Timeouts = nullTimeoutsValue()

	mapBucketReplicaLinkToModel(link, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapBucketReplicaLinkToModel maps a client.BucketReplicaLink to the Terraform model.
func mapBucketReplicaLinkToModel(link *client.BucketReplicaLink, data *bucketReplicaLinkModel) {
	data.ID = types.StringValue(link.ID)
	data.LocalBucketName = types.StringValue(link.LocalBucket.Name)
	data.RemoteBucketName = types.StringValue(link.RemoteBucket.Name)
	data.RemoteName = stringOrNull(link.Remote.Name)
	data.Paused = types.BoolValue(link.Paused)
	data.CascadingEnabled = types.BoolValue(link.CascadingEnabled)
	data.Direction = stringOrNull(link.Direction)
	data.Status = stringOrNull(link.Status)
	data.StatusDetails = stringOrNull(link.StatusDetails)
	data.Lag = types.Int64Value(link.Lag)
	data.RecoveryPoint = types.Int64Value(link.RecoveryPoint)

	if link.RemoteCredentials != nil {
		data.RemoteCredentialsName = types.StringValue(link.RemoteCredentials.Name)
	}
	// If RemoteCredentials is nil, preserve the existing state value for this Optional field.

	if link.ObjectBacklog != nil {
		data.ObjectBacklogCount = types.Int64Value(link.ObjectBacklog.Count)
		data.ObjectBacklogTotalSize = types.Int64Value(link.ObjectBacklog.TotalSize)
	} else {
		data.ObjectBacklogCount = types.Int64Value(0)
		data.ObjectBacklogTotalSize = types.Int64Value(0)
	}
}

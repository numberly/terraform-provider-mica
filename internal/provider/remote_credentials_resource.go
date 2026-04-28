package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &remoteCredentialsResource{}
var _ resource.ResourceWithConfigure = &remoteCredentialsResource{}
var _ resource.ResourceWithImportState = &remoteCredentialsResource{}
var _ resource.ResourceWithUpgradeState = &remoteCredentialsResource{}

// remoteCredentialsResource implements the flashblade_object_store_remote_credentials resource.
type remoteCredentialsResource struct {
	client *client.FlashBladeClient
}

func NewRemoteCredentialsResource() resource.Resource {
	return &remoteCredentialsResource{}
}

// ---------- model structs ----------------------------------------------------

// remoteCredentialsModel is the top-level model for the flashblade_object_store_remote_credentials resource.
type remoteCredentialsModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	AccessKeyID     types.String   `tfsdk:"access_key_id"`
	SecretAccessKey types.String   `tfsdk:"secret_access_key"`
	RemoteName      types.String   `tfsdk:"remote_name"`
	TargetName      types.String   `tfsdk:"target_name"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// remoteCredentialsV0Model is the v0 state model (no target_name field).
type remoteCredentialsV0Model struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	AccessKeyID     types.String   `tfsdk:"access_key_id"`
	SecretAccessKey types.String   `tfsdk:"secret_access_key"`
	RemoteName      types.String   `tfsdk:"remote_name"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *remoteCredentialsResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_remote_credentials"
}

// Schema defines the resource schema.
func (r *remoteCredentialsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages FlashBlade object store remote credentials for cross-array bucket replication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the remote credentials.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the remote credentials. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_key_id": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The access key ID for the remote S3 credentials.",
			},
			"secret_access_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The secret access key for the remote S3 credentials.",
			},
			"remote_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the remote array connection. Populated automatically from the API response. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the target (S3-compatible endpoint). Mutually exclusive with remote_name. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
func (r *remoteCredentialsResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// v0 -> v1: add target_name attribute (null for all existing resources).
		0: {
			PriorSchema: &schema.Schema{
				Version:     0,
				Description: "Manages FlashBlade object store remote credentials for cross-array bucket replication.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Required: true,
					},
					"access_key_id": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
					"secret_access_key": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
					"remote_name": schema.StringAttribute{
						Required: true,
					},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
						Create: true,
						Read:   true,
						Update: true,
						Delete: true,
					}),
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState remoteCredentialsV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				newState := remoteCredentialsModel{
					ID:              oldState.ID,
					Name:            oldState.Name,
					AccessKeyID:     oldState.AccessKeyID,
					SecretAccessKey: oldState.SecretAccessKey,
					RemoteName:      oldState.RemoteName,
					TargetName:      types.StringNull(),
					Timeouts:        oldState.Timeouts,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *remoteCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates new remote credentials.
func (r *remoteCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data remoteCredentialsModel
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

	post := client.ObjectStoreRemoteCredentialsPost{
		AccessKeyID:     data.AccessKeyID.ValueString(),
		SecretAccessKey: data.SecretAccessKey.ValueString(),
	}

	// Route to PostRemoteCredentialsForTarget or PostRemoteCredentialsForRemote
	// based on which attribute is set.
	var cred *client.ObjectStoreRemoteCredentials
	var err error
	switch {
	case !data.TargetName.IsNull() && !data.TargetName.IsUnknown() && data.TargetName.ValueString() != "":
		cred, err = r.client.PostRemoteCredentialsForTarget(ctx, data.Name.ValueString(), data.TargetName.ValueString(), post)
	case !data.RemoteName.IsNull() && !data.RemoteName.IsUnknown() && data.RemoteName.ValueString() != "":
		cred, err = r.client.PostRemoteCredentialsForRemote(ctx, data.Name.ValueString(), data.RemoteName.ValueString(), post)
	default:
		resp.Diagnostics.AddError("Error creating remote credentials", "either target_name or remote_name must be set")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error creating remote credentials", err.Error())
		return
	}

	// Preserve target_name from plan (not returned by API).
	planTargetName := data.TargetName

	mapRemoteCredentialsToModel(cred, &data)
	// Preserve user-provided secrets in state (API may not return them on subsequent reads).
	data.AccessKeyID = types.StringValue(post.AccessKeyID)
	data.SecretAccessKey = types.StringValue(post.SecretAccessKey)
	// Preserve target_name from plan (API does not return it).
	data.TargetName = planTargetName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *remoteCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data remoteCredentialsModel
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

	name := data.Name.ValueString()
	cred, err := r.client.GetRemoteCredentials(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading remote credentials", err.Error())
		return
	}

	// Preserve secret_access_key and target_name from state — GET does not return them.
	existingSecret := data.SecretAccessKey
	existingTargetName := data.TargetName

	mapRemoteCredentialsToModel(cred, &data)
	data.SecretAccessKey = existingSecret
	data.TargetName = existingTargetName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to existing remote credentials (key rotation).
func (r *remoteCredentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state remoteCredentialsModel
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

	patch := client.ObjectStoreRemoteCredentialsPatch{}

	if !plan.AccessKeyID.Equal(state.AccessKeyID) {
		v := plan.AccessKeyID.ValueString()
		patch.AccessKeyID = &v
	}
	if !plan.SecretAccessKey.Equal(state.SecretAccessKey) {
		v := plan.SecretAccessKey.ValueString()
		patch.SecretAccessKey = &v
	}

	_, err := r.client.PatchRemoteCredentials(ctx, state.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating remote credentials", err.Error())
		return
	}

	// Re-read to refresh computed fields.
	cred, err := r.client.GetRemoteCredentials(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading remote credentials after update", err.Error())
		return
	}

	// Preserve target_name from plan (API does not return it).
	planTargetName := plan.TargetName

	mapRemoteCredentialsToModel(cred, &plan)
	// Preserve user-provided secrets in state from plan values.
	plan.AccessKeyID = types.StringValue(plan.AccessKeyID.ValueString())
	plan.SecretAccessKey = types.StringValue(plan.SecretAccessKey.ValueString())
	// Preserve target_name from plan (API does not return it).
	plan.TargetName = planTargetName

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes remote credentials.
func (r *remoteCredentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data remoteCredentialsModel
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

	err := r.client.DeleteRemoteCredentials(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting remote credentials", err.Error())
		return
	}
}

func (r *remoteCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	cred, err := r.client.GetRemoteCredentials(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing remote credentials", err.Error())
		return
	}

	var data remoteCredentialsModel
	data.Timeouts = nullTimeoutsValue()

	mapRemoteCredentialsToModel(cred, &data)
	// secret_access_key will be empty after import — user must provide it in config or use ignore_changes.
	data.SecretAccessKey = types.StringValue("")
	// target_name is not returned by GET — set to null on import.
	data.TargetName = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapRemoteCredentialsToModel maps a client.ObjectStoreRemoteCredentials to a remoteCredentialsModel.
// It preserves user-managed fields (Timeouts, SecretAccessKey, TargetName).
// IMPORTANT: Does NOT set SecretAccessKey or TargetName — GET does not return them.
func mapRemoteCredentialsToModel(cred *client.ObjectStoreRemoteCredentials, data *remoteCredentialsModel) {
	data.ID = types.StringValue(cred.ID)
	data.Name = types.StringValue(cred.Name)
	data.AccessKeyID = types.StringValue(cred.AccessKeyID)
	data.RemoteName = types.StringValue(cred.Remote.Name)
}

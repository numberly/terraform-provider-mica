package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure remoteCredentialsResource satisfies the resource interfaces.
var _ resource.Resource = &remoteCredentialsResource{}
var _ resource.ResourceWithConfigure = &remoteCredentialsResource{}
var _ resource.ResourceWithImportState = &remoteCredentialsResource{}
var _ resource.ResourceWithUpgradeState = &remoteCredentialsResource{}

// remoteCredentialsResource implements the flashblade_object_store_remote_credentials resource.
type remoteCredentialsResource struct {
	client *client.FlashBladeClient
}

// NewRemoteCredentialsResource is the factory function registered in the provider.
func NewRemoteCredentialsResource() resource.Resource {
	return &remoteCredentialsResource{}
}

// ---------- model structs ----------------------------------------------------

// remoteCredentialsModel is the top-level model for the flashblade_object_store_remote_credentials resource.
type remoteCredentialsModel struct {
	ID             types.String   `tfsdk:"id"`
	Name           types.String   `tfsdk:"name"`
	AccessKeyID    types.String   `tfsdk:"access_key_id"`
	SecretAccessKey types.String  `tfsdk:"secret_access_key"`
	RemoteName     types.String   `tfsdk:"remote_name"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *remoteCredentialsResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_remote_credentials"
}

// Schema defines the resource schema.
func (r *remoteCredentialsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages FlashBlade object store remote credentials for cross-array bucket replication.",
		Version:     0,
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
				Required:    true,
				Description: "The name of the remote array connection. Changing this forces a new resource.",
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

func (r *remoteCredentialsResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
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
		AccessKeyID:    data.AccessKeyID.ValueString(),
		SecretAccessKey: data.SecretAccessKey.ValueString(),
	}

	cred, err := r.client.PostRemoteCredentials(ctx, data.Name.ValueString(), data.RemoteName.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating remote credentials", err.Error())
		return
	}

	mapRemoteCredentialsToModel(cred, &data)
	// Preserve user-provided secrets in state (API may not return them on subsequent reads).
	data.AccessKeyID = types.StringValue(post.AccessKeyID)
	data.SecretAccessKey = types.StringValue(post.SecretAccessKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
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

	// Preserve secret_access_key from state — GET does not return it.
	existingSecret := data.SecretAccessKey

	mapRemoteCredentialsToModel(cred, &data)
	data.SecretAccessKey = existingSecret

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

	mapRemoteCredentialsToModel(cred, &plan)
	// Preserve user-provided secrets in state from plan values.
	plan.AccessKeyID = types.StringValue(plan.AccessKeyID.ValueString())
	plan.SecretAccessKey = types.StringValue(plan.SecretAccessKey.ValueString())

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

// ImportState imports existing remote credentials by name.
func (r *remoteCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	cred, err := r.client.GetRemoteCredentials(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing remote credentials", err.Error())
		return
	}

	var data remoteCredentialsModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}

	mapRemoteCredentialsToModel(cred, &data)
	// secret_access_key will be empty after import — user must provide it in config or use ignore_changes.
	data.SecretAccessKey = types.StringValue("")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapRemoteCredentialsToModel maps a client.ObjectStoreRemoteCredentials to a remoteCredentialsModel.
// It preserves user-managed fields (Timeouts, SecretAccessKey).
// IMPORTANT: Does NOT set SecretAccessKey — GET does not return it.
func mapRemoteCredentialsToModel(cred *client.ObjectStoreRemoteCredentials, data *remoteCredentialsModel) {
	data.ID = types.StringValue(cred.ID)
	data.Name = types.StringValue(cred.Name)
	data.AccessKeyID = types.StringValue(cred.AccessKeyID)
	data.RemoteName = types.StringValue(cred.Remote.Name)
}

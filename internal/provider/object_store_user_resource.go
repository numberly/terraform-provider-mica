package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure objectStoreUserResource satisfies the resource interfaces.
var _ resource.Resource = &objectStoreUserResource{}
var _ resource.ResourceWithConfigure = &objectStoreUserResource{}
var _ resource.ResourceWithImportState = &objectStoreUserResource{}
var _ resource.ResourceWithUpgradeState = &objectStoreUserResource{}

// objectStoreUserResource implements the flashblade_object_store_user resource.
type objectStoreUserResource struct {
	client *client.FlashBladeClient
}

// NewObjectStoreUserResource is the factory function registered in the provider.
func NewObjectStoreUserResource() resource.Resource {
	return &objectStoreUserResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreUserModel is the top-level model for the flashblade_object_store_user resource.
type objectStoreUserModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	FullAccess types.Bool     `tfsdk:"full_access"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *objectStoreUserResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_user"
}

// Schema defines the resource schema.
func (r *objectStoreUserResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade object store user (S3 user within an account).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store user in the format account/username. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"full_access": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the user has full access to all object store operations. Defaults to false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

// UpgradeState returns state upgraders for schema migrations.
func (r *objectStoreUserResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *objectStoreUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new object store user.
func (r *objectStoreUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreUserModel
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

	body := client.ObjectStoreUserPost{}
	if !data.FullAccess.IsNull() && !data.FullAccess.IsUnknown() {
		v := data.FullAccess.ValueBool()
		body.FullAccess = &v
	}

	_, err := r.client.PostObjectStoreUser(ctx, data.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store user", err.Error())
		return
	}

	// Re-read from API to get the actual state (POST response may not reflect full_access).
	user, err := r.client.GetObjectStoreUser(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading object store user after create", err.Error())
		return
	}

	// full_access is write-only — preserve from plan; default to false if unset.
	if data.FullAccess.IsNull() || data.FullAccess.IsUnknown() {
		data.FullAccess = types.BoolValue(false)
	}
	mapObjectStoreUserToModel(user, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *objectStoreUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreUserModel
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
	user, err := r.client.GetObjectStoreUser(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store user", err.Error())
		return
	}

	// full_access is write-only (not returned by API) — preserve from state.
	// If state has no value yet (e.g. after import), default to false.
	if data.FullAccess.IsNull() || data.FullAccess.IsUnknown() {
		data.FullAccess = types.BoolValue(false)
	}
	mapObjectStoreUserToModel(user, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported — the FlashBlade API does not allow PATCH on object store users.
func (r *objectStoreUserResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"flashblade_object_store_user does not support in-place updates. All attributes that changed require replacement.",
	)
}

// Delete removes an object store user.
func (r *objectStoreUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreUserModel
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

	if err := r.client.DeleteObjectStoreUser(ctx, data.Name.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting object store user", err.Error())
		return
	}
}

// ImportState imports an existing object store user by name (format: account/username).
func (r *objectStoreUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	user, err := r.client.GetObjectStoreUser(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Object store user not found",
				fmt.Sprintf("No object store user with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error importing object store user", err.Error())
		return
	}

	var data objectStoreUserModel
	// Initialize timeouts with a null object (no update timeout — CRD only).
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	mapObjectStoreUserToModel(user, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapObjectStoreUserToModel maps a client.ObjectStoreUser to an objectStoreUserModel.
// Note: full_access is a write-only POST parameter — the API does not return it
// in GET/POST responses. The caller must preserve FullAccess from plan or state.
func mapObjectStoreUserToModel(user *client.ObjectStoreUser, data *objectStoreUserModel) {
	data.ID = types.StringValue(user.ID)
	data.Name = types.StringValue(user.Name)
}

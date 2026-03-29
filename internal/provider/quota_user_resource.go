package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure quotaUserResource satisfies the resource interfaces.
var _ resource.Resource = &quotaUserResource{}
var _ resource.ResourceWithConfigure = &quotaUserResource{}
var _ resource.ResourceWithImportState = &quotaUserResource{}
var _ resource.ResourceWithUpgradeState = &quotaUserResource{}

// quotaUserResource implements the flashblade_quota_user resource.
type quotaUserResource struct {
	client *client.FlashBladeClient
}

// NewQuotaUserResource is the factory function registered in the provider.
func NewQuotaUserResource() resource.Resource {
	return &quotaUserResource{}
}

// ---------- model structs ----------------------------------------------------

// quotaUserModel is the top-level model for the flashblade_quota_user resource.
type quotaUserModel struct {
	ID             types.String   `tfsdk:"id"`
	FileSystemName types.String   `tfsdk:"file_system_name"`
	UID            types.String   `tfsdk:"uid"`
	Quota          types.Int64    `tfsdk:"quota"`
	Usage          types.Int64    `tfsdk:"usage"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *quotaUserResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_quota_user"
}

// Schema defines the resource schema.
func (r *quotaUserResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a per-filesystem user quota on a FlashBlade array.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic composite identifier in the form file_system_name/uid.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_system_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the file system this quota applies to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uid": schema.StringAttribute{
				Required:    true,
				Description: "User ID (UID) the quota applies to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"quota": schema.Int64Attribute{
				Required:    true,
				Description: "Quota limit in bytes.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"usage": schema.Int64Attribute{
				Computed:    true,
				Description: "Current usage in bytes (read-only, API-managed).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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

func (r *quotaUserResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *quotaUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new user quota.
func (r *quotaUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data quotaUserModel
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

	fsName := data.FileSystemName.ValueString()
	uid := data.UID.ValueString()

	_, err := r.client.PostQuotaUser(ctx, fsName, uid, client.QuotaUserPost{
		Quota: data.Quota.ValueInt64(),
	})
	if err != nil {
		// FlashBlade pre-creates implicit zero-quota entries for UIDs that access the FS.
		// Fall back to PATCH if the quota already exists.
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 400 {
			q := data.Quota.ValueInt64()
			_, patchErr := r.client.PatchQuotaUser(ctx, fsName, uid, client.QuotaUserPatch{
				Quota: &q,
			})
			if patchErr != nil {
				resp.Diagnostics.AddError("Error creating user quota", patchErr.Error())
				return
			}
		} else {
			resp.Diagnostics.AddError("Error creating user quota", err.Error())
			return
		}
	}

	r.readIntoState(ctx, fsName, uid, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *quotaUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data quotaUserModel
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

	fsName := data.FileSystemName.ValueString()
	uid := data.UID.ValueString()

	qu, err := r.client.GetQuotaUser(ctx, fsName, uid)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user quota", err.Error())
		return
	}

	// Drift detection on quota.
	if !data.Quota.IsNull() && !data.Quota.IsUnknown() {
		if data.Quota.ValueInt64() != qu.Quota {
			tflog.Info(ctx, "drift detected on user quota", map[string]any{
				"resource":    fsName + "/" + uid,
				"field":       "quota",
				"state_value": data.Quota.ValueInt64(),
				"api_value":   qu.Quota,
			})
		}
	}

	mapQuotaUserToModel(fsName, uid, qu, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing user quota (only quota is patchable).
func (r *quotaUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state quotaUserModel
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

	fsName := state.FileSystemName.ValueString()
	uid := state.UID.ValueString()

	patch := client.QuotaUserPatch{}
	if !plan.Quota.Equal(state.Quota) {
		v := plan.Quota.ValueInt64()
		patch.Quota = &v
	}

	_, err := r.client.PatchQuotaUser(ctx, fsName, uid, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user quota", err.Error())
		return
	}

	r.readIntoState(ctx, fsName, uid, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a user quota.
func (r *quotaUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data quotaUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	fsName := data.FileSystemName.ValueString()
	uid := data.UID.ValueString()

	if err := r.client.DeleteQuotaUser(ctx, fsName, uid); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting user quota", err.Error())
		return
	}
}

// ImportState imports an existing user quota by composite ID "file_system_name/uid".
func (r *quotaUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the form 'file_system_name/uid', got: %q", req.ID),
		)
		return
	}
	fsName := parts[0]
	uid := parts[1]

	var data quotaUserModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}
	data.FileSystemName = types.StringValue(fsName)
	data.UID = types.StringValue(uid)

	r.readIntoState(ctx, fsName, uid, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetQuotaUser and maps the result into the provided model.
func (r *quotaUserResource) readIntoState(ctx context.Context, fsName, uid string, data *quotaUserModel, diags interface {
	AddError(string, string)
	HasError() bool
}) {
	qu, err := r.client.GetQuotaUser(ctx, fsName, uid)
	if err != nil {
		diags.AddError("Error reading user quota after write", err.Error())
		return
	}
	mapQuotaUserToModel(fsName, uid, qu, data)
}

// mapQuotaUserToModel maps a client.QuotaUser to a quotaUserModel.
// Preserves user-managed fields (Timeouts).
func mapQuotaUserToModel(fsName, uid string, qu *client.QuotaUser, data *quotaUserModel) {
	data.ID = types.StringValue(compositeID(fsName, uid))
	data.FileSystemName = types.StringValue(fsName)
	data.UID = types.StringValue(uid)
	data.Quota = types.Int64Value(qu.Quota)
	data.Usage = types.Int64Value(qu.Usage)
}

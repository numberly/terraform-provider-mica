package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &quotaGroupResource{}
var _ resource.ResourceWithConfigure = &quotaGroupResource{}
var _ resource.ResourceWithImportState = &quotaGroupResource{}
var _ resource.ResourceWithUpgradeState = &quotaGroupResource{}

// quotaGroupResource implements the flashblade_quota_group resource.
type quotaGroupResource struct {
	client *client.FlashBladeClient
}

func NewQuotaGroupResource() resource.Resource {
	return &quotaGroupResource{}
}

// ---------- model structs ----------------------------------------------------

// quotaGroupModel is the top-level model for the flashblade_quota_group resource.
type quotaGroupModel struct {
	ID             types.String   `tfsdk:"id"`
	FileSystemName types.String   `tfsdk:"file_system_name"`
	GID            types.String   `tfsdk:"gid"`
	Quota          types.Int64    `tfsdk:"quota"`
	Usage          types.Int64    `tfsdk:"usage"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *quotaGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_quota_group"
}

// Schema defines the resource schema.
func (r *quotaGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a per-filesystem group quota on a FlashBlade array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic composite identifier in the form file_system_name/gid.",
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
			"gid": schema.StringAttribute{
				Required:    true,
				Description: "Group ID (GID) the quota applies to. Changing this forces a new resource.",
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


// UpgradeState returns state upgraders for schema migrations.
func (r *quotaGroupResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *quotaGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *quotaGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data quotaGroupModel
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
	gid := data.GID.ValueString()

	_, err := r.client.PostQuotaGroup(ctx, fsName, gid, client.QuotaGroupPost{
		Quota: data.Quota.ValueInt64(),
	})
	if err != nil {
		// FlashBlade pre-creates implicit zero-quota entries for GIDs that access the FS.
		// Fall back to PATCH if the quota already exists.
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 400 {
			q := data.Quota.ValueInt64()
			_, patchErr := r.client.PatchQuotaGroup(ctx, fsName, gid, client.QuotaGroupPatch{
				Quota: &q,
			})
			if patchErr != nil {
				resp.Diagnostics.AddError("Error creating group quota", patchErr.Error())
				return
			}
		} else {
			resp.Diagnostics.AddError("Error creating group quota", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, fsName, gid, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *quotaGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data quotaGroupModel
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
	gid := data.GID.ValueString()

	qg, err := r.client.GetQuotaGroup(ctx, fsName, gid)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group quota", err.Error())
		return
	}

	// Drift detection on quota.
	if !data.Quota.IsNull() && !data.Quota.IsUnknown() {
		if data.Quota.ValueInt64() != qg.Quota {
			tflog.Debug(ctx, "drift detected on group quota", map[string]any{
				"resource":    fsName + "/" + gid,
				"field":       "quota",
				"was":         data.Quota.ValueInt64(),
				"now":           qg.Quota,
			})
		}
	}

	mapQuotaGroupToModel(fsName, gid, qg, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing group quota (only quota is patchable).
func (r *quotaGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state quotaGroupModel
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
	gid := state.GID.ValueString()

	patch := client.QuotaGroupPatch{}
	if !plan.Quota.Equal(state.Quota) {
		v := plan.Quota.ValueInt64()
		patch.Quota = &v
	}

	_, err := r.client.PatchQuotaGroup(ctx, fsName, gid, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating group quota", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, fsName, gid, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a group quota.
func (r *quotaGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data quotaGroupModel
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
	gid := data.GID.ValueString()

	if err := r.client.DeleteQuotaGroup(ctx, fsName, gid); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting group quota", err.Error())
		return
	}
}

// ImportState imports an existing group quota by composite ID "file_system_name/gid".
func (r *quotaGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the form 'file_system_name/gid', got: %q", req.ID),
		)
		return
	}
	fsName := parts[0]
	gid := parts[1]

	var data quotaGroupModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	data.FileSystemName = types.StringValue(fsName)
	data.GID = types.StringValue(gid)

	resp.Diagnostics.Append(r.readIntoState(ctx, fsName, gid, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetQuotaGroup and maps the result into the provided model.
func (r *quotaGroupResource) readIntoState(ctx context.Context, fsName, gid string, data *quotaGroupModel) diag.Diagnostics {
	var diags diag.Diagnostics

	qg, err := r.client.GetQuotaGroup(ctx, fsName, gid)
	if err != nil {
		diags.AddError("Error reading group quota after write", err.Error())
		return diags
	}
	mapQuotaGroupToModel(fsName, gid, qg, data)
	return diags
}


// mapQuotaGroupToModel maps a client.QuotaGroup to a quotaGroupModel.
// Preserves user-managed fields (Timeouts).
func mapQuotaGroupToModel(fsName, gid string, qg *client.QuotaGroup, data *quotaGroupModel) {
	data.ID = types.StringValue(compositeID(fsName, gid))
	data.FileSystemName = types.StringValue(fsName)
	data.GID = types.StringValue(gid)
	data.Quota = types.Int64Value(qg.Quota)
	data.Usage = types.Int64Value(qg.Usage)
}

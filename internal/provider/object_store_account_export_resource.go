package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &objectStoreAccountExportResource{}
var _ resource.ResourceWithConfigure = &objectStoreAccountExportResource{}
var _ resource.ResourceWithImportState = &objectStoreAccountExportResource{}
var _ resource.ResourceWithUpgradeState = &objectStoreAccountExportResource{}

// objectStoreAccountExportResource implements the flashblade_object_store_account_export resource.
type objectStoreAccountExportResource struct {
	client *client.FlashBladeClient
}

func NewObjectStoreAccountExportResource() resource.Resource {
	return &objectStoreAccountExportResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccountExportModel is the top-level model for the flashblade_object_store_account_export resource.
type objectStoreAccountExportModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	AccountName types.String   `tfsdk:"account_name"`
	ServerName  types.String   `tfsdk:"server_name"`
	Enabled     types.Bool     `tfsdk:"enabled"`
	PolicyName  types.String   `tfsdk:"policy_name"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *objectStoreAccountExportResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_account_export"
}

// Schema defines the resource schema.
func (r *objectStoreAccountExportResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade object store account export.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store account export.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The combined name of the export (e.g. 'account/export_name').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store account to export. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the server to export to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the export is enabled. Defaults to true.",
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the S3 export policy to apply to the export.",
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
func (r *objectStoreAccountExportResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *objectStoreAccountExportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *objectStoreAccountExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreAccountExportModel
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

	post := client.ObjectStoreAccountExportPost{
		ExportEnabled: data.Enabled.ValueBool(),
		Server:        &client.NamedReference{Name: data.ServerName.ValueString()},
	}

	export, err := r.client.PostObjectStoreAccountExport(ctx, data.AccountName.ValueString(), data.PolicyName.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store account export", err.Error())
		return
	}

	mapObjectStoreAccountExportToModel(export, &data)

	// Apply policy_name if set, since POST does not support it.
	if !data.PolicyName.IsNull() && !data.PolicyName.IsUnknown() && data.PolicyName.ValueString() != "" {
		patch := client.ObjectStoreAccountExportPatch{
			Policy: &client.NamedReference{Name: data.PolicyName.ValueString()},
		}
		updated, patchErr := r.client.PatchObjectStoreAccountExport(ctx, export.ID, patch)
		if patchErr != nil {
			resp.Diagnostics.AddError("Error setting policy on object store account export", patchErr.Error())
			return
		}
		mapObjectStoreAccountExportToModel(updated, &data)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *objectStoreAccountExportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreAccountExportModel
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
	export, err := r.client.GetObjectStoreAccountExport(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store account export", err.Error())
		return
	}

	mapObjectStoreAccountExportToModel(export, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing object store account export.
func (r *objectStoreAccountExportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state objectStoreAccountExportModel
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

	patch := client.ObjectStoreAccountExportPatch{}

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.ExportEnabled = &v
	}
	if !plan.PolicyName.Equal(state.PolicyName) {
		patch.Policy = &client.NamedReference{Name: plan.PolicyName.ValueString()}
	}

	_, err := r.client.PatchObjectStoreAccountExport(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating object store account export", err.Error())
		return
	}

	r.readIntoState(ctx, state.Name.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an object store account export.
func (r *objectStoreAccountExportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreAccountExportModel
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

	memberName := data.AccountName.ValueString()
	combinedName := data.Name.ValueString()
	// The API returns Name as "account/export" but Delete expects just the short export name.
	exportName := combinedName
	if idx := strings.LastIndex(combinedName, "/"); idx >= 0 {
		exportName = combinedName[idx+1:]
	}

	err := r.client.DeleteObjectStoreAccountExport(ctx, memberName, exportName)
	if err != nil {
		if client.IsNotFound(err) {
			// Already gone — no error.
			return
		}
		resp.Diagnostics.AddError("Error deleting object store account export", err.Error())
		return
	}
}

// ImportState imports an existing object store account export by its combined name.
func (r *objectStoreAccountExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	export, err := r.client.GetObjectStoreAccountExport(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing object store account export", err.Error())
		return
	}

	var data objectStoreAccountExportModel
	// Initialize timeouts with a proper null value.
	data.Timeouts = nullTimeoutsValue()

	mapObjectStoreAccountExportToModel(export, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetObjectStoreAccountExport and maps the result into the provided model.
func (r *objectStoreAccountExportResource) readIntoState(ctx context.Context, name string, data *objectStoreAccountExportModel, diags DiagnosticReporter) {
	export, err := r.client.GetObjectStoreAccountExport(ctx, name)
	if err != nil {
		diags.AddError("Error reading object store account export after write", err.Error())
		return
	}
	mapObjectStoreAccountExportToModel(export, data)
}

// mapObjectStoreAccountExportToModel maps a client.ObjectStoreAccountExport to an objectStoreAccountExportModel.
// It preserves user-managed fields (Timeouts).
func mapObjectStoreAccountExportToModel(export *client.ObjectStoreAccountExport, data *objectStoreAccountExportModel) {
	data.ID = types.StringValue(export.ID)
	data.Name = types.StringValue(export.Name)
	data.Enabled = types.BoolValue(export.Enabled)

	if export.Member != nil {
		data.AccountName = types.StringValue(export.Member.Name)
	}
	if export.Server != nil {
		data.ServerName = types.StringValue(export.Server.Name)
	}
	if export.Policy != nil {
		data.PolicyName = types.StringValue(export.Policy.Name)
	} else {
		data.PolicyName = types.StringNull()
	}
}

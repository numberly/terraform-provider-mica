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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure fileSystemExportResource satisfies the resource interfaces.
var _ resource.Resource = &fileSystemExportResource{}
var _ resource.ResourceWithConfigure = &fileSystemExportResource{}
var _ resource.ResourceWithImportState = &fileSystemExportResource{}
var _ resource.ResourceWithUpgradeState = &fileSystemExportResource{}

// fileSystemExportResource implements the flashblade_file_system_export resource.
type fileSystemExportResource struct {
	client *client.FlashBladeClient
}

// NewFileSystemExportResource is the factory function registered in the provider.
func NewFileSystemExportResource() resource.Resource {
	return &fileSystemExportResource{}
}

// ---------- model structs ----------------------------------------------------

// fileSystemExportModel is the top-level model for the flashblade_file_system_export resource.
type fileSystemExportModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	ExportName      types.String   `tfsdk:"export_name"`
	FileSystemName  types.String   `tfsdk:"file_system_name"`
	ServerName      types.String   `tfsdk:"server_name"`
	PolicyName      types.String   `tfsdk:"policy_name"`
	SharePolicyName types.String   `tfsdk:"share_policy_name"`
	Enabled         types.Bool     `tfsdk:"enabled"`
	PolicyType      types.String   `tfsdk:"policy_type"`
	Status          types.String   `tfsdk:"status"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *fileSystemExportResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_file_system_export"
}

// Schema defines the resource schema.
func (r *fileSystemExportResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade file system export.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the file system export.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The combined name of the export (e.g. 'filesystem/export_name').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"export_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The export name part. Defaults to the file system name if not set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_system_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the file system to export. Changing this forces a new resource.",
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
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the NFS export policy to apply to the export.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"share_policy_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the SMB share policy to apply to the export.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the export is enabled.",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The policy type ('nfs' or 'smb').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the export.",
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
func (r *fileSystemExportResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *fileSystemExportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new file system export.
func (r *fileSystemExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fileSystemExportModel
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

	exportName := data.FileSystemName.ValueString() // default to FS name
	if !data.ExportName.IsNull() && !data.ExportName.IsUnknown() {
		exportName = data.ExportName.ValueString()
	}
	post := client.FileSystemExportPost{
		ExportName: exportName,
		Server:     &client.NamedReference{Name: data.ServerName.ValueString()},
	}
	if !data.SharePolicyName.IsNull() && !data.SharePolicyName.IsUnknown() {
		post.SharePolicy = &client.NamedReference{Name: data.SharePolicyName.ValueString()}
	}

	export, err := r.client.PostFileSystemExport(ctx, data.FileSystemName.ValueString(), data.PolicyName.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating file system export", err.Error())
		return
	}

	mapFileSystemExportToModel(export, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *fileSystemExportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data fileSystemExportModel
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
	export, err := r.client.GetFileSystemExport(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading file system export", err.Error())
		return
	}

	// Drift detection on user-configurable fields.
	if !data.ExportName.IsNull() && !data.ExportName.IsUnknown() {
		if data.ExportName.ValueString() != export.ExportName {
			tflog.Info(ctx, "drift detected on file system export", map[string]any{
				"resource":    name,
				"field":       "export_name",
				"state_value": data.ExportName.ValueString(),
				"api_value":   export.ExportName,
			})
		}
	}
	if !data.SharePolicyName.IsNull() && !data.SharePolicyName.IsUnknown() {
		apiSharePolicy := ""
		if export.SharePolicy != nil {
			apiSharePolicy = export.SharePolicy.Name
		}
		if data.SharePolicyName.ValueString() != apiSharePolicy {
			tflog.Info(ctx, "drift detected on file system export", map[string]any{
				"resource":    name,
				"field":       "share_policy_name",
				"state_value": data.SharePolicyName.ValueString(),
				"api_value":   apiSharePolicy,
			})
		}
	}

	mapFileSystemExportToModel(export, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing file system export.
func (r *fileSystemExportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state fileSystemExportModel
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

	patch := client.FileSystemExportPatch{}

	if !plan.ExportName.Equal(state.ExportName) {
		v := plan.ExportName.ValueString()
		patch.ExportName = &v
	}
	if !plan.SharePolicyName.Equal(state.SharePolicyName) {
		v := plan.SharePolicyName.ValueString()
		patch.SharePolicy = &client.NamedReference{Name: v}
	}

	_, err := r.client.PatchFileSystemExport(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating file system export", err.Error())
		return
	}

	r.readIntoState(ctx, state.Name.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a file system export.
func (r *fileSystemExportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data fileSystemExportModel
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

	memberName := data.FileSystemName.ValueString()
	exportName := data.ExportName.ValueString()

	err := r.client.DeleteFileSystemExport(ctx, memberName, exportName)
	if err != nil {
		if client.IsNotFound(err) {
			// Already gone — no error.
			return
		}
		resp.Diagnostics.AddError("Error deleting file system export", err.Error())
		return
	}
}

// ImportState imports an existing file system export by its combined name.
func (r *fileSystemExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	export, err := r.client.GetFileSystemExport(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing file system export", err.Error())
		return
	}

	var data fileSystemExportModel
	// Initialize timeouts with a proper null value.
	data.Timeouts = nullTimeoutsValue()

	mapFileSystemExportToModel(export, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetFileSystemExport and maps the result into the provided model.
func (r *fileSystemExportResource) readIntoState(ctx context.Context, name string, data *fileSystemExportModel, diags DiagnosticReporter) {
	export, err := r.client.GetFileSystemExport(ctx, name)
	if err != nil {
		diags.AddError("Error reading file system export after write", err.Error())
		return
	}
	mapFileSystemExportToModel(export, data)
}

// mapFileSystemExportToModel maps a client.FileSystemExport to a fileSystemExportModel.
// It preserves user-managed fields (Timeouts).
func mapFileSystemExportToModel(export *client.FileSystemExport, data *fileSystemExportModel) {
	data.ID = types.StringValue(export.ID)
	data.Name = types.StringValue(export.Name)
	data.ExportName = types.StringValue(export.ExportName)
	data.Enabled = types.BoolValue(export.Enabled)
	data.PolicyType = types.StringValue(export.PolicyType)
	data.Status = types.StringValue(export.Status)

	if export.Member != nil {
		data.FileSystemName = types.StringValue(export.Member.Name)
	}
	if export.Server != nil {
		data.ServerName = types.StringValue(export.Server.Name)
	}
	if export.Policy != nil && export.Policy.Name != "" {
		data.PolicyName = types.StringValue(export.Policy.Name)
	}
	if export.SharePolicy != nil && export.SharePolicy.Name != "" {
		data.SharePolicyName = types.StringValue(export.SharePolicy.Name)
	} else {
		data.SharePolicyName = types.StringNull()
	}
}

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure filesystemResource satisfies the resource.Resource interface.
var _ resource.Resource = &filesystemResource{}
var _ resource.ResourceWithConfigure = &filesystemResource{}
var _ resource.ResourceWithImportState = &filesystemResource{}

// filesystemResource implements the flashblade_file_system resource.
type filesystemResource struct {
	client *client.FlashBladeClient
}

// NewFilesystemResource is the factory function registered in the provider.
func NewFilesystemResource() resource.Resource {
	return &filesystemResource{}
}

// ---------- nested model structs --------------------------------------------

// filesystemSpaceModel maps space attributes.
type filesystemSpaceModel struct {
	DataReduction      types.Float64 `tfsdk:"data_reduction"`
	Snapshots          types.Int64   `tfsdk:"snapshots"`
	TotalPhysical      types.Int64   `tfsdk:"total_physical"`
	Unique             types.Int64   `tfsdk:"unique"`
	Virtual            types.Int64   `tfsdk:"virtual"`
	SnapshotsEffective types.Int64   `tfsdk:"snapshots_effective"`
}

// filesystemNFSModel maps NFS attributes.
type filesystemNFSModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	V3Enabled  types.Bool   `tfsdk:"v3_enabled"`
	V41Enabled types.Bool   `tfsdk:"v4_1_enabled"`
	Rules      types.String `tfsdk:"rules"`
	Transport  types.String `tfsdk:"transport"`
}

// filesystemSMBModel maps SMB attributes.
type filesystemSMBModel struct {
	Enabled                       types.Bool `tfsdk:"enabled"`
	AccessBasedEnumerationEnabled types.Bool `tfsdk:"access_based_enumeration_enabled"`
	ContinuousAvailabilityEnabled types.Bool `tfsdk:"continuous_availability_enabled"`
	SMBEncryptionEnabled          types.Bool `tfsdk:"smb_encryption_enabled"`
}

// filesystemHTTPModel maps HTTP attributes.
type filesystemHTTPModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// filesystemMultiProtocolModel maps multi_protocol attributes.
type filesystemMultiProtocolModel struct {
	AccessControlStyle types.String `tfsdk:"access_control_style"`
	SafeguardACLs      types.Bool   `tfsdk:"safeguard_acls"`
}

// filesystemDefaultQuotasModel maps default_quotas attributes.
type filesystemDefaultQuotasModel struct {
	GroupQuota types.Int64 `tfsdk:"group_quota"`
	UserQuota  types.Int64 `tfsdk:"user_quota"`
}

// filesystemSourceModel maps the source reference.
type filesystemSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// filesystemModel is the top-level model for the flashblade_file_system resource.
type filesystemModel struct {
	ID                       types.String                  `tfsdk:"id"`
	Name                     types.String                  `tfsdk:"name"`
	Provisioned              types.Int64                   `tfsdk:"provisioned"`
	Destroyed                types.Bool                    `tfsdk:"destroyed"`
	DestroyEradicateOnDelete types.Bool                    `tfsdk:"destroy_eradicate_on_delete"`
	TimeRemaining            types.Int64                   `tfsdk:"time_remaining"`
	Created                  types.Int64                   `tfsdk:"created"`
	PromotionStatus          types.String                  `tfsdk:"promotion_status"`
	Writable                 types.Bool                    `tfsdk:"writable"`
	NFSExportPolicy          types.String                  `tfsdk:"nfs_export_policy"`
	SMBSharePolicy           types.String                  `tfsdk:"smb_share_policy"`
	Space                    *filesystemSpaceModel         `tfsdk:"space"`
	NFS                      *filesystemNFSModel           `tfsdk:"nfs"`
	SMB                      *filesystemSMBModel           `tfsdk:"smb"`
	HTTP                     *filesystemHTTPModel          `tfsdk:"http"`
	MultiProtocol            *filesystemMultiProtocolModel `tfsdk:"multi_protocol"`
	DefaultQuotas            *filesystemDefaultQuotasModel `tfsdk:"default_quotas"`
	Source                   *filesystemSourceModel        `tfsdk:"source"`
	Timeouts                 timeouts.Value                `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *filesystemResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_file_system"
}

// Schema defines the resource schema.
func (r *filesystemResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade file system.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the file system.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the file system. Supports in-place rename.",
			},
			"provisioned": schema.Int64Attribute{
				Required:    true,
				Description: "Provisioned size of the file system in bytes.",
			},
			"destroyed": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the file system is soft-deleted.",
			},
			"destroy_eradicate_on_delete": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "When true (default), Terraform will eradicate the file system on destroy. When false, only soft-deletes.",
			},
			"time_remaining": schema.Int64Attribute{
				Computed:    true,
				Description: "Milliseconds remaining until auto-eradication of a soft-deleted file system.",
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the file system was created.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"promotion_status": schema.StringAttribute{
				Computed:    true,
				Description: "Replication promotion status of the file system.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"writable": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the file system is writable.",
			},
			"nfs_export_policy": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the NFS export policy to apply to this file system.",
			},
			"smb_share_policy": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the SMB share policy to apply to this file system.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"space": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Storage space breakdown (read-only, API-managed).",
				Attributes: map[string]schema.Attribute{
					"data_reduction": schema.Float64Attribute{
						Computed:    true,
						Description: "Data reduction ratio.",
					},
					"snapshots": schema.Int64Attribute{
						Computed:    true,
						Description: "Physical space used by snapshots in bytes.",
					},
					"total_physical": schema.Int64Attribute{
						Computed:    true,
						Description: "Total physical space used in bytes.",
					},
					"unique": schema.Int64Attribute{
						Computed:    true,
						Description: "Unique physical space used in bytes.",
					},
					"virtual": schema.Int64Attribute{
						Computed:    true,
						Description: "Virtual (logical) space used in bytes.",
					},
					"snapshots_effective": schema.Int64Attribute{
						Computed:    true,
						Description: "Effective snapshot space used in bytes.",
					},
				},
			},
			"nfs": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "NFS protocol configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Whether NFS is enabled on this file system.",
					},
					"v3_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether NFSv3 is enabled.",
					},
					"v4_1_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Whether NFSv4.1 is enabled.",
					},
					"rules": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "NFS export rules string (e.g. '*(rw,no_root_squash)').",
					},
					"transport": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "NFS transport protocol ('tcp' or 'udp').",
					},
				},
			},
			"smb": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SMB protocol configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether SMB is enabled on this file system.",
					},
					"access_based_enumeration_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether access-based enumeration is enabled for SMB.",
					},
					"continuous_availability_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether continuous availability is enabled for SMB.",
					},
					"smb_encryption_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether SMB encryption is enabled.",
					},
				},
			},
			"http": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "HTTP protocol configuration (read-only, API-managed).",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether HTTP is enabled on this file system.",
					},
				},
			},
			"multi_protocol": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Multi-protocol access configuration.",
				Attributes: map[string]schema.Attribute{
					"access_control_style": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Access control style for multi-protocol access ('nfs' or 'smb').",
					},
					"safeguard_acls": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to safeguard ACLs during multi-protocol access.",
					},
				},
			},
			"default_quotas": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default quota settings.",
				Attributes: map[string]schema.Attribute{
					"group_quota": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Default quota per group in bytes.",
					},
					"user_quota": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Default quota per user in bytes.",
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Source file system reference (for clones/replicas, read-only).",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "Source file system ID.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "Source file system name.",
					},
				},
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *filesystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new file system.
func (r *filesystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data filesystemModel
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

	post := client.FileSystemPost{
		Name:        data.Name.ValueString(),
		Provisioned: data.Provisioned.ValueInt64(),
	}

	if data.NFS != nil {
		post.NFS = client.NFSConfig{
			Enabled:    data.NFS.Enabled.ValueBool(),
			V3Enabled:  data.NFS.V3Enabled.ValueBool(),
			V41Enabled: data.NFS.V41Enabled.ValueBool(),
			Rules:      data.NFS.Rules.ValueString(),
			Transport:  data.NFS.Transport.ValueString(),
		}
	}
	if data.SMB != nil {
		post.SMB = client.SMBConfig{
			Enabled:                       data.SMB.Enabled.ValueBool(),
			AccessBasedEnumerationEnabled: data.SMB.AccessBasedEnumerationEnabled.ValueBool(),
			ContinuousAvailabilityEnabled: data.SMB.ContinuousAvailabilityEnabled.ValueBool(),
			SMBEncryptionEnabled:          data.SMB.SMBEncryptionEnabled.ValueBool(),
		}
	}

	_, err := r.client.PostFileSystem(ctx, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating file system", err.Error())
		return
	}

	r.readIntoState(ctx, data.Name.ValueString(), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *filesystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data filesystemModel
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
	fs, err := r.client.GetFileSystem(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading file system", err.Error())
		return
	}

	// Drift detection: compare user-configurable fields against current state.
	if !data.Provisioned.IsNull() && !data.Provisioned.IsUnknown() {
		if data.Provisioned.ValueInt64() != fs.Provisioned {
			tflog.Info(ctx, "drift detected on file system", map[string]any{
				"resource":    name,
				"field":       "provisioned",
				"state_value": data.Provisioned.ValueInt64(),
				"api_value":   fs.Provisioned,
			})
		}
	}

	// Map API response to model.
	mapFSToModel(fs, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing file system.
func (r *filesystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state filesystemModel
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

	patch := client.FileSystemPatch{}

	// Only patch changed fields.
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Provisioned.Equal(state.Provisioned) {
		v := plan.Provisioned.ValueInt64()
		patch.Provisioned = &v
	}

	// NFS changes.
	if plan.NFS != nil {
		nfs := client.NFSConfig{
			Enabled:    plan.NFS.Enabled.ValueBool(),
			V3Enabled:  plan.NFS.V3Enabled.ValueBool(),
			V41Enabled: plan.NFS.V41Enabled.ValueBool(),
			Rules:      plan.NFS.Rules.ValueString(),
			Transport:  plan.NFS.Transport.ValueString(),
		}
		patch.NFS = &nfs
	}

	// SMB changes.
	if plan.SMB != nil {
		smb := client.SMBConfig{
			Enabled:                       plan.SMB.Enabled.ValueBool(),
			AccessBasedEnumerationEnabled: plan.SMB.AccessBasedEnumerationEnabled.ValueBool(),
			ContinuousAvailabilityEnabled: plan.SMB.ContinuousAvailabilityEnabled.ValueBool(),
			SMBEncryptionEnabled:          plan.SMB.SMBEncryptionEnabled.ValueBool(),
		}
		patch.SMB = &smb
	}

	_, err := r.client.PatchFileSystem(ctx, state.ID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating file system", err.Error())
		return
	}

	// Use the new name for read-at-end-of-write (in case of rename).
	newName := plan.Name.ValueString()
	r.readIntoState(ctx, newName, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a file system via two-phase soft-delete + eradicate.
func (r *filesystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data filesystemModel
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

	id := data.ID.ValueString()
	name := data.Name.ValueString()

	// Phase 1: Soft-delete.
	destroyed := true
	_, err := r.client.PatchFileSystem(ctx, id, client.FileSystemPatch{Destroyed: &destroyed})
	if err != nil {
		if client.IsNotFound(err) {
			// Already gone — no error.
			return
		}
		resp.Diagnostics.AddError("Error soft-deleting file system", err.Error())
		return
	}

	// Phase 2 + 3: Eradicate only if destroy_eradicate_on_delete is true (or null/default).
	eradicate := data.DestroyEradicateOnDelete.IsNull() || data.DestroyEradicateOnDelete.ValueBool()
	if eradicate {
		if err := r.client.DeleteFileSystem(ctx, id); err != nil {
			if !client.IsNotFound(err) {
				resp.Diagnostics.AddError("Error eradicating file system", err.Error())
				return
			}
		}
		if err := r.client.PollUntilEradicated(ctx, name); err != nil {
			resp.Diagnostics.AddError("Error waiting for file system eradication", err.Error())
			return
		}
	}
}

// ImportState imports an existing file system by name.
func (r *filesystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	fs, err := r.client.GetFileSystem(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing file system", err.Error())
		return
	}

	// Start with an empty model and fill it from the API.
	var data filesystemModel
	// Set defaults for optional fields not returned by the API.
	data.DestroyEradicateOnDelete = types.BoolValue(true)
	// Initialize timeouts with a proper null value so the framework can serialize it.
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}

	mapFSToModel(fs, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetFileSystem and maps the result into the provided model.
// This is the Read-at-end-of-write pattern ensuring state reflects true API state.
func (r *filesystemResource) readIntoState(ctx context.Context, name string, data *filesystemModel, diags interface {
	AddError(string, string)
	HasError() bool
}) {
	fs, err := r.client.GetFileSystem(ctx, name)
	if err != nil {
		diags.AddError("Error reading file system after write", err.Error())
		return
	}
	mapFSToModel(fs, data)
}

// mapFSToModel maps a client.FileSystem to a filesystemModel.
// It preserves user-managed fields (DestroyEradicateOnDelete, Timeouts, policy fields).
func mapFSToModel(fs *client.FileSystem, data *filesystemModel) {
	data.ID = types.StringValue(fs.ID)
	data.Name = types.StringValue(fs.Name)
	data.Provisioned = types.Int64Value(fs.Provisioned)
	data.Destroyed = types.BoolValue(fs.Destroyed)
	data.Created = types.Int64Value(fs.Created)
	data.PromotionStatus = types.StringValue(fs.PromotionStatus)
	data.Writable = types.BoolValue(fs.Writable)

	if fs.TimeRemaining != 0 {
		data.TimeRemaining = types.Int64Value(fs.TimeRemaining)
	} else {
		data.TimeRemaining = types.Int64Value(0)
	}

	// Space (always set — Computed-only block).
	data.Space = &filesystemSpaceModel{
		DataReduction:      types.Float64Value(fs.Space.DataReduction),
		Snapshots:          types.Int64Value(fs.Space.Snapshots),
		TotalPhysical:      types.Int64Value(fs.Space.TotalPhysical),
		Unique:             types.Int64Value(fs.Space.Unique),
		Virtual:            types.Int64Value(fs.Space.Virtual),
		SnapshotsEffective: types.Int64Value(fs.Space.SnapshotsEffective),
	}

	// NFS block — always set from API.
	data.NFS = &filesystemNFSModel{
		Enabled:    types.BoolValue(fs.NFS.Enabled),
		V3Enabled:  types.BoolValue(fs.NFS.V3Enabled),
		V41Enabled: types.BoolValue(fs.NFS.V41Enabled),
		Rules:      types.StringValue(fs.NFS.Rules),
		Transport:  types.StringValue(fs.NFS.Transport),
	}

	// SMB block — always set from API.
	data.SMB = &filesystemSMBModel{
		Enabled:                       types.BoolValue(fs.SMB.Enabled),
		AccessBasedEnumerationEnabled: types.BoolValue(fs.SMB.AccessBasedEnumerationEnabled),
		ContinuousAvailabilityEnabled: types.BoolValue(fs.SMB.ContinuousAvailabilityEnabled),
		SMBEncryptionEnabled:          types.BoolValue(fs.SMB.SMBEncryptionEnabled),
	}

	// HTTP block — always set from API (Computed-only).
	data.HTTP = &filesystemHTTPModel{
		Enabled: types.BoolValue(fs.HTTP.Enabled),
	}

	// Source block — only if present in API response.
	if fs.Source != nil {
		data.Source = &filesystemSourceModel{
			ID:   types.StringValue(fs.Source.ID),
			Name: types.StringValue(fs.Source.Name),
		}
	} else {
		data.Source = nil
	}

	// MultiProtocol — only set if data contains it (Optional/Computed).
	if data.MultiProtocol == nil && (fs.MultiProtocol.AccessControlStyle != "" || fs.MultiProtocol.SafeguardACLsOnDestroy) {
		data.MultiProtocol = &filesystemMultiProtocolModel{
			AccessControlStyle: types.StringValue(fs.MultiProtocol.AccessControlStyle),
			SafeguardACLs:      types.BoolValue(fs.MultiProtocol.SafeguardACLsOnDestroy),
		}
	} else if data.MultiProtocol != nil {
		data.MultiProtocol.AccessControlStyle = types.StringValue(fs.MultiProtocol.AccessControlStyle)
		data.MultiProtocol.SafeguardACLs = types.BoolValue(fs.MultiProtocol.SafeguardACLsOnDestroy)
	}

	// DefaultQuotas — only set if data contains it (Optional/Computed).
	if data.DefaultQuotas == nil && (fs.DefaultQuotas.GroupQuota != 0 || fs.DefaultQuotas.UserQuota != 0) {
		data.DefaultQuotas = &filesystemDefaultQuotasModel{
			GroupQuota: types.Int64Value(fs.DefaultQuotas.GroupQuota),
			UserQuota:  types.Int64Value(fs.DefaultQuotas.UserQuota),
		}
	} else if data.DefaultQuotas != nil {
		data.DefaultQuotas.GroupQuota = types.Int64Value(fs.DefaultQuotas.GroupQuota)
		data.DefaultQuotas.UserQuota = types.Int64Value(fs.DefaultQuotas.UserQuota)
	}
}

// ---------- plan modifier helpers -------------------------------------------

// int64UseStateForUnknown returns an Int64 plan modifier that preserves state value
// when the planned value is unknown (equivalent to stringplanmodifier.UseStateForUnknown).
func int64UseStateForUnknown() planmodifier.Int64 {
	return &int64UseStateForUnknownModifier{}
}

type int64UseStateForUnknownModifier struct{}

func (m *int64UseStateForUnknownModifier) Description(_ context.Context) string {
	return "Use state value for unknown planned values."
}

func (m *int64UseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Use state value for unknown planned values."
}

func (m *int64UseStateForUnknownModifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

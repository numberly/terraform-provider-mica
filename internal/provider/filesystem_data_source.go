package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure filesystemDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &filesystemDataSource{}
var _ datasource.DataSourceWithConfigure = &filesystemDataSource{}

// filesystemDataSource implements the flashblade_file_system data source.
type filesystemDataSource struct {
	client *client.FlashBladeClient
}

// NewFilesystemDataSource is the factory function registered in the provider.
func NewFilesystemDataSource() datasource.DataSource {
	return &filesystemDataSource{}
}

// ---------- nested model structs (data source) ------------------------------

// filesystemDataSourceModel is the top-level model for the flashblade_file_system data source.
type filesystemDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Provisioned     types.Int64  `tfsdk:"provisioned"`
	Destroyed       types.Bool   `tfsdk:"destroyed"`
	TimeRemaining   types.Int64  `tfsdk:"time_remaining"`
	Created         types.Int64  `tfsdk:"created"`
	PromotionStatus types.String `tfsdk:"promotion_status"`
	Writable        types.Bool   `tfsdk:"writable"`
	Space           types.Object `tfsdk:"space"`
	NFS             types.Object `tfsdk:"nfs"`
	SMB             types.Object `tfsdk:"smb"`
	HTTP            types.Object `tfsdk:"http"`
	MultiProtocol   types.Object `tfsdk:"multi_protocol"`
	DefaultQuotas   types.Object `tfsdk:"default_quotas"`
	Source          types.Object `tfsdk:"source"`
}

// ---------- data source interface methods -----------------------------------

// Metadata sets the Terraform type name.
func (d *filesystemDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_file_system"
}

// Schema defines the data source schema.
// All attributes are Computed except name (Required, used as lookup key).
func (d *filesystemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade file system by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the file system.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the file system to look up.",
			},
			"provisioned": schema.Int64Attribute{
				Computed:    true,
				Description: "Provisioned size of the file system in bytes.",
			},
			"destroyed": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the file system is soft-deleted.",
			},
			"time_remaining": schema.Int64Attribute{
				Computed:    true,
				Description: "Milliseconds remaining until auto-eradication of a soft-deleted file system.",
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the file system was created.",
			},
			"promotion_status": schema.StringAttribute{
				Computed:    true,
				Description: "Replication promotion status of the file system.",
			},
			"writable": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the file system is writable.",
			},
			"space": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Storage space breakdown.",
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
				Computed:    true,
				Description: "NFS protocol configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether NFS is enabled on this file system.",
					},
					"v3_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether NFSv3 is enabled.",
					},
					"v4_1_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether NFSv4.1 is enabled.",
					},
					"rules": schema.StringAttribute{
						Computed:    true,
						Description: "NFS export rules string.",
					},
					"transport": schema.StringAttribute{
						Computed:    true,
						Description: "NFS transport protocol.",
					},
				},
			},
			"smb": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "SMB protocol configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether SMB is enabled on this file system.",
					},
					"access_based_enumeration_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether access-based enumeration is enabled for SMB.",
					},
					"continuous_availability_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether continuous availability is enabled for SMB.",
					},
					"smb_encryption_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether SMB encryption is enabled.",
					},
				},
			},
			"http": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "HTTP protocol configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether HTTP is enabled on this file system.",
					},
				},
			},
			"multi_protocol": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Multi-protocol access configuration.",
				Attributes: map[string]schema.Attribute{
					"access_control_style": schema.StringAttribute{
						Computed:    true,
						Description: "Access control style for multi-protocol access.",
					},
					"safeguard_acls": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether ACLs are safeguarded during multi-protocol access.",
					},
				},
			},
			"default_quotas": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Default quota settings.",
				Attributes: map[string]schema.Attribute{
					"group_quota": schema.Int64Attribute{
						Computed:    true,
						Description: "Default quota per group in bytes.",
					},
					"user_quota": schema.Int64Attribute{
						Computed:    true,
						Description: "Default quota per user in bytes.",
					},
				},
			},
			"source": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Source file system reference (for clones/replicas).",
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

// Configure injects the FlashBladeClient into the data source.
func (d *filesystemDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches file system data by name and populates state.
func (d *filesystemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config filesystemDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	fs, err := d.client.GetFileSystem(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"File system not found",
				fmt.Sprintf("No file system with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading file system", err.Error())
		return
	}

	// Map API response to model.
	config.ID = types.StringValue(fs.ID)
	config.Name = types.StringValue(fs.Name)
	config.Provisioned = types.Int64Value(fs.Provisioned)
	config.Destroyed = types.BoolValue(fs.Destroyed)
	config.Created = types.Int64Value(fs.Created)
	config.PromotionStatus = types.StringValue(fs.PromotionStatus)
	config.Writable = types.BoolValue(fs.Writable)
	config.TimeRemaining = types.Int64Value(fs.TimeRemaining)

	// Space (always set — Computed-only block).
	config.Space = mustObjectValue(fsSpaceAttrTypes(), map[string]attr.Value{
		"data_reduction":      types.Float64Value(fs.Space.DataReduction),
		"snapshots":           types.Int64Value(fs.Space.Snapshots),
		"total_physical":      types.Int64Value(fs.Space.TotalPhysical),
		"unique":              types.Int64Value(fs.Space.Unique),
		"virtual":             types.Int64Value(fs.Space.Virtual),
		"snapshots_effective": types.Int64Value(fs.Space.SnapshotsEffective),
	})

	// NFS block — always set from API.
	config.NFS = mustObjectValue(fsNFSAttrTypes(), map[string]attr.Value{
		"enabled":      types.BoolValue(fs.NFS.Enabled),
		"v3_enabled":   types.BoolValue(fs.NFS.V3Enabled),
		"v4_1_enabled": types.BoolValue(fs.NFS.V41Enabled),
		"rules":        types.StringValue(fs.NFS.Rules),
		"transport":    types.StringValue(fs.NFS.Transport),
	})

	// SMB block — always set from API.
	config.SMB = mustObjectValue(fsSMBAttrTypes(), map[string]attr.Value{
		"enabled":                          types.BoolValue(fs.SMB.Enabled),
		"access_based_enumeration_enabled": types.BoolValue(fs.SMB.AccessBasedEnumerationEnabled),
		"continuous_availability_enabled":  types.BoolValue(fs.SMB.ContinuousAvailabilityEnabled),
		"smb_encryption_enabled":           types.BoolValue(fs.SMB.SMBEncryptionEnabled),
	})

	// HTTP block — always set from API (Computed-only).
	config.HTTP = mustObjectValue(fsHTTPAttrTypes(), map[string]attr.Value{
		"enabled": types.BoolValue(fs.HTTP.Enabled),
	})

	// Source block — only if present in API response.
	if fs.Source != nil {
		config.Source = mustObjectValue(fsSourceAttrTypes(), map[string]attr.Value{
			"id":   types.StringValue(fs.Source.ID),
			"name": types.StringValue(fs.Source.Name),
		})
	} else {
		config.Source = types.ObjectNull(fsSourceAttrTypes())
	}

	// MultiProtocol — always set from API.
	config.MultiProtocol = mustObjectValue(fsMultiProtocolAttrTypes(), map[string]attr.Value{
		"access_control_style": types.StringValue(fs.MultiProtocol.AccessControlStyle),
		"safeguard_acls":       types.BoolValue(fs.MultiProtocol.SafeguardACLsOnDestroy),
	})

	// DefaultQuotas — always set from API.
	config.DefaultQuotas = mustObjectValue(fsDefaultQuotasAttrTypes(), map[string]attr.Value{
		"group_quota": types.Int64Value(fs.DefaultQuotas.GroupQuota),
		"user_quota":  types.Int64Value(fs.DefaultQuotas.UserQuota),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

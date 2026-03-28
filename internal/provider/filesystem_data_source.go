package provider

import (
	"context"
	"fmt"

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

// filesystemDataSourceSpaceModel mirrors the resource space model.
type filesystemDataSourceSpaceModel struct {
	DataReduction      types.Float64 `tfsdk:"data_reduction"`
	Snapshots          types.Int64   `tfsdk:"snapshots"`
	TotalPhysical      types.Int64   `tfsdk:"total_physical"`
	Unique             types.Int64   `tfsdk:"unique"`
	Virtual            types.Int64   `tfsdk:"virtual"`
	SnapshotsEffective types.Int64   `tfsdk:"snapshots_effective"`
}

// filesystemDataSourceNFSModel mirrors the resource NFS model.
type filesystemDataSourceNFSModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	V3Enabled  types.Bool   `tfsdk:"v3_enabled"`
	V41Enabled types.Bool   `tfsdk:"v4_1_enabled"`
	Rules      types.String `tfsdk:"rules"`
	Transport  types.String `tfsdk:"transport"`
}

// filesystemDataSourceSMBModel mirrors the resource SMB model.
type filesystemDataSourceSMBModel struct {
	Enabled                       types.Bool `tfsdk:"enabled"`
	AccessBasedEnumerationEnabled types.Bool `tfsdk:"access_based_enumeration_enabled"`
	ContinuousAvailabilityEnabled types.Bool `tfsdk:"continuous_availability_enabled"`
	SMBEncryptionEnabled          types.Bool `tfsdk:"smb_encryption_enabled"`
}

// filesystemDataSourceHTTPModel mirrors the resource HTTP model.
type filesystemDataSourceHTTPModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// filesystemDataSourceMultiProtocolModel mirrors the resource multi_protocol model.
type filesystemDataSourceMultiProtocolModel struct {
	AccessControlStyle types.String `tfsdk:"access_control_style"`
	SafeguardACLs      types.Bool   `tfsdk:"safeguard_acls"`
}

// filesystemDataSourceDefaultQuotasModel mirrors the resource default_quotas model.
type filesystemDataSourceDefaultQuotasModel struct {
	GroupQuota types.Int64 `tfsdk:"group_quota"`
	UserQuota  types.Int64 `tfsdk:"user_quota"`
}

// filesystemDataSourceSourceModel mirrors the resource source model.
type filesystemDataSourceSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// filesystemDataSourceModel is the top-level model for the flashblade_file_system data source.
type filesystemDataSourceModel struct {
	ID              types.String                            `tfsdk:"id"`
	Name            types.String                            `tfsdk:"name"`
	Provisioned     types.Int64                             `tfsdk:"provisioned"`
	Destroyed       types.Bool                              `tfsdk:"destroyed"`
	TimeRemaining   types.Int64                             `tfsdk:"time_remaining"`
	Created         types.Int64                             `tfsdk:"created"`
	PromotionStatus types.String                            `tfsdk:"promotion_status"`
	Writable        types.Bool                              `tfsdk:"writable"`
	Space           *filesystemDataSourceSpaceModel         `tfsdk:"space"`
	NFS             *filesystemDataSourceNFSModel           `tfsdk:"nfs"`
	SMB             *filesystemDataSourceSMBModel           `tfsdk:"smb"`
	HTTP            *filesystemDataSourceHTTPModel          `tfsdk:"http"`
	MultiProtocol   *filesystemDataSourceMultiProtocolModel `tfsdk:"multi_protocol"`
	DefaultQuotas   *filesystemDataSourceDefaultQuotasModel `tfsdk:"default_quotas"`
	Source          *filesystemDataSourceSourceModel        `tfsdk:"source"`
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
		},
		Blocks: map[string]schema.Block{
			"space": schema.SingleNestedBlock{
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
			"nfs": schema.SingleNestedBlock{
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
			"smb": schema.SingleNestedBlock{
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
			"http": schema.SingleNestedBlock{
				Description: "HTTP protocol configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether HTTP is enabled on this file system.",
					},
				},
			},
			"multi_protocol": schema.SingleNestedBlock{
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
			"default_quotas": schema.SingleNestedBlock{
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
			"source": schema.SingleNestedBlock{
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

	config.Space = &filesystemDataSourceSpaceModel{
		DataReduction:      types.Float64Value(fs.Space.DataReduction),
		Snapshots:          types.Int64Value(fs.Space.Snapshots),
		TotalPhysical:      types.Int64Value(fs.Space.TotalPhysical),
		Unique:             types.Int64Value(fs.Space.Unique),
		Virtual:            types.Int64Value(fs.Space.Virtual),
		SnapshotsEffective: types.Int64Value(fs.Space.SnapshotsEffective),
	}

	config.NFS = &filesystemDataSourceNFSModel{
		Enabled:    types.BoolValue(fs.NFS.Enabled),
		V3Enabled:  types.BoolValue(fs.NFS.V3Enabled),
		V41Enabled: types.BoolValue(fs.NFS.V41Enabled),
		Rules:      types.StringValue(fs.NFS.Rules),
		Transport:  types.StringValue(fs.NFS.Transport),
	}

	config.SMB = &filesystemDataSourceSMBModel{
		Enabled:                       types.BoolValue(fs.SMB.Enabled),
		AccessBasedEnumerationEnabled: types.BoolValue(fs.SMB.AccessBasedEnumerationEnabled),
		ContinuousAvailabilityEnabled: types.BoolValue(fs.SMB.ContinuousAvailabilityEnabled),
		SMBEncryptionEnabled:          types.BoolValue(fs.SMB.SMBEncryptionEnabled),
	}

	config.HTTP = &filesystemDataSourceHTTPModel{
		Enabled: types.BoolValue(fs.HTTP.Enabled),
	}

	if fs.Source != nil {
		config.Source = &filesystemDataSourceSourceModel{
			ID:   types.StringValue(fs.Source.ID),
			Name: types.StringValue(fs.Source.Name),
		}
	} else {
		config.Source = nil
	}

	if fs.MultiProtocol.AccessControlStyle != "" || fs.MultiProtocol.SafeguardACLsOnDestroy {
		config.MultiProtocol = &filesystemDataSourceMultiProtocolModel{
			AccessControlStyle: types.StringValue(fs.MultiProtocol.AccessControlStyle),
			SafeguardACLs:      types.BoolValue(fs.MultiProtocol.SafeguardACLsOnDestroy),
		}
	}

	if fs.DefaultQuotas.GroupQuota != 0 || fs.DefaultQuotas.UserQuota != 0 {
		config.DefaultQuotas = &filesystemDataSourceDefaultQuotasModel{
			GroupQuota: types.Int64Value(fs.DefaultQuotas.GroupQuota),
			UserQuota:  types.Int64Value(fs.DefaultQuotas.UserQuota),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

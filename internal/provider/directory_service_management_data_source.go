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

// Ensure directoryServiceManagementDataSource satisfies the datasource interfaces.
var _ datasource.DataSource = &directoryServiceManagementDataSource{}
var _ datasource.DataSourceWithConfigure = &directoryServiceManagementDataSource{}

// directoryServiceManagementDataSource implements the flashblade_directory_service_management data source.
type directoryServiceManagementDataSource struct {
	client *client.FlashBladeClient
}

// NewDirectoryServiceManagementDataSource is the factory function registered in the provider.
func NewDirectoryServiceManagementDataSource() datasource.DataSource {
	return &directoryServiceManagementDataSource{}
}

// ---------- model structs ----------------------------------------------------

// directoryServiceManagementDataSourceModel is the top-level model for the
// flashblade_directory_service_management data source.
// Computed-only: no name (singleton always reads "management"), no bind_password (write-only).
type directoryServiceManagementDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	URIs                  types.List   `tfsdk:"uris"`
	BaseDN                types.String `tfsdk:"base_dn"`
	BindUser              types.String `tfsdk:"bind_user"`
	CACertificate         types.Object `tfsdk:"ca_certificate"`
	CACertificateGroup    types.Object `tfsdk:"ca_certificate_group"`
	UserLoginAttribute    types.String `tfsdk:"user_login_attribute"`
	UserObjectClass       types.String `tfsdk:"user_object_class"`
	SSHPublicKeyAttribute types.String `tfsdk:"ssh_public_key_attribute"`
	Services              types.List   `tfsdk:"services"`
}

// ---------- helpers ----------------------------------------------------------

// namedRefAttrTypes returns the attr.Type map for a NamedReference nested object.
func namedRefAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{"name": types.StringType}
}

// namedRefObjectValue converts a *NamedReference to a types.Object.
// Returns a null object when ref is nil.
func namedRefObjectValue(ref *client.NamedReference) types.Object {
	if ref == nil {
		return types.ObjectNull(namedRefAttrTypes())
	}
	obj, _ := types.ObjectValue(namedRefAttrTypes(), map[string]attr.Value{
		"name": types.StringValue(ref.Name),
	})
	return obj
}

// ---------- data source interface methods ------------------------------------

// Metadata sets the Terraform type name.
func (d *directoryServiceManagementDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_directory_service_management"
}

// Schema defines the data source schema.
// All attributes are Computed (read-only view of the singleton management directory service).
// ca_certificate and ca_certificate_group are SingleNestedAttribute with a "name" sub-attribute (D-06).
// No name argument (always reads "management"). No bind_password (write-only, never readable).
// No timeouts block (data sources do not use timeouts per CONVENTIONS.md).
func (d *directoryServiceManagementDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	nestedRef := schema.SingleNestedAttribute{
		Computed:    true,
		Description: "Reference to a named object. Null when the reference is not set.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the referenced object. Null when the reference is not set.",
			},
		},
	}
	resp.Schema = schema.Schema{
		Description: "Reads the current FlashBlade management directory service (LDAP admin) configuration. Singleton — no arguments needed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the management directory service.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the management directory service is enabled.",
			},
			"uris": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of LDAP server URIs (e.g. ldaps://ldap.example.com:636).",
			},
			"base_dn": schema.StringAttribute{
				Computed:    true,
				Description: "Base Distinguished Name used when searching the directory.",
			},
			"bind_user": schema.StringAttribute{
				Computed:    true,
				Description: "DN of the user used to bind to the directory.",
			},
			"ca_certificate":       nestedRef,
			"ca_certificate_group": nestedRef,
			"user_login_attribute": schema.StringAttribute{
				Computed:    true,
				Description: "LDAP attribute holding the user's login name (e.g. sAMAccountName or uid).",
			},
			"user_object_class": schema.StringAttribute{
				Computed:    true,
				Description: "LDAP object class for management users (e.g. User, posixAccount).",
			},
			"ssh_public_key_attribute": schema.StringAttribute{
				Computed:    true,
				Description: "LDAP attribute holding the user's SSH public key.",
			},
			"services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Services using this directory service configuration.",
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *directoryServiceManagementDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the management directory service configuration and populates state.
// Always reads the singleton with name "management" — no user-supplied name argument.
func (d *directoryServiceManagementDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config directoryServiceManagementDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := d.client.GetDirectoryServiceManagement(ctx, "management")
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Directory service management not configured",
				"No management directory service configured on this FlashBlade array.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading directory service management config", err.Error())
		return
	}

	// Identity + connection fields
	config.ID = types.StringValue(ds.ID)
	config.Enabled = types.BoolValue(ds.Enabled)
	config.BaseDN = types.StringValue(ds.BaseDN)
	config.BindUser = types.StringValue(ds.BindUser)

	// Management sub-object fields
	config.UserLoginAttribute = types.StringValue(ds.Management.UserLoginAttribute)
	config.UserObjectClass = types.StringValue(ds.Management.UserObjectClass)
	config.SSHPublicKeyAttribute = types.StringValue(ds.Management.SSHPublicKeyAttribute)

	// URIs list
	if len(ds.URIs) > 0 {
		l, diags := types.ListValueFrom(ctx, types.StringType, ds.URIs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.URIs = l
	} else {
		config.URIs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Services list
	if len(ds.Services) > 0 {
		l, diags := types.ListValueFrom(ctx, types.StringType, ds.Services)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		config.Services = l
	} else {
		config.Services = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// TLS references as nested objects (D-06)
	config.CACertificate = namedRefObjectValue(ds.CACertificate)
	config.CACertificateGroup = namedRefObjectValue(ds.CACertificateGroup)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

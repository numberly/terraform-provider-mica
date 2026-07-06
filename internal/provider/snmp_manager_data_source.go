package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &snmpManagerDataSource{}
var _ datasource.DataSourceWithConfigure = &snmpManagerDataSource{}

// snmpManagerDataSource implements the flashblade_snmp_manager data source.
type snmpManagerDataSource struct {
	client *client.FlashBladeClient
}

func NewSnmpManagerDataSource() datasource.DataSource {
	return &snmpManagerDataSource{}
}

// ---------- model structs ----------------------------------------------------

type snmpManagerDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Host         types.String `tfsdk:"host"`
	Notification types.String `tfsdk:"notification"`
	Version      types.String `tfsdk:"version"`
	V2c          types.Object `tfsdk:"v2c"`
	V3           types.Object `tfsdk:"v3"`
}

// ---------- data source interface methods -----------------------------------

func (d *snmpManagerDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_snmp_manager"
}

func (d *snmpManagerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade SNMP manager (trap/inform destination) by name. Sensitive fields (community, auth_passphrase, privacy_passphrase) are never returned by the API and surface as null.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SNMP manager.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the SNMP manager to look up.",
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "DNS name or IP address (with optional :port) of the SNMP receiver.",
			},
			"notification": schema.StringAttribute{
				Computed:    true,
				Description: "Notification delivery mode (inform or trap).",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "SNMP protocol version (v2c or v3).",
			},
			"v2c": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "SNMPv2c configuration block.",
				Attributes: map[string]schema.Attribute{
					"community": schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "Community string. Always null on read (never returned by the API).",
					},
				},
			},
			"v3": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "SNMPv3 configuration block.",
				Attributes: map[string]schema.Attribute{
					"user": schema.StringAttribute{
						Computed:    true,
						Description: "SNMPv3 username.",
					},
					"auth_protocol": schema.StringAttribute{
						Computed:    true,
						Description: "Authentication protocol (MD5 or SHA).",
					},
					"auth_passphrase": schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "Authentication passphrase. Always null on read (never returned by the API).",
					},
					"privacy_protocol": schema.StringAttribute{
						Computed:    true,
						Description: "Privacy protocol (AES or DES).",
					},
					"privacy_passphrase": schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "Privacy passphrase. Always null on read (never returned by the API).",
					},
				},
			},
		},
	}
}

func (d *snmpManagerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *snmpManagerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config snmpManagerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	mgr, err := d.client.GetSnmpManager(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"SNMP manager not found",
				fmt.Sprintf("No SNMP manager with name %q exists on the FlashBlade array.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading SNMP manager", err.Error())
		return
	}

	config.ID = types.StringValue(mgr.ID)
	config.Name = types.StringValue(mgr.Name)
	config.Host = types.StringValue(mgr.Host)
	config.Notification = types.StringValue(mgr.Notification)
	config.Version = types.StringValue(mgr.Version)

	if mgr.V2c != nil {
		// Community is never returned by the API; surface as null.
		v2c, d := types.ObjectValue(snmpV2cAttrTypes, map[string]attr.Value{
			"community": types.StringNull(),
		})
		resp.Diagnostics.Append(d...)
		config.V2c = v2c
	} else {
		config.V2c = types.ObjectNull(snmpV2cAttrTypes)
	}

	if mgr.V3 != nil {
		v3, d := types.ObjectValue(snmpV3AttrTypes, map[string]attr.Value{
			"user":               stringFromAPI(mgr.V3.User),
			"auth_protocol":      stringFromAPI(mgr.V3.AuthProtocol),
			"auth_passphrase":    types.StringNull(),
			"privacy_protocol":   stringFromAPI(mgr.V3.PrivacyProtocol),
			"privacy_passphrase": types.StringNull(),
		})
		resp.Diagnostics.Append(d...)
		config.V3 = v3
	} else {
		config.V3 = types.ObjectNull(snmpV3AttrTypes)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

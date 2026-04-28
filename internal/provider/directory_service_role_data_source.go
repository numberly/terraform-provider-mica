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

var _ datasource.DataSource = &directoryServiceRoleDataSource{}
var _ datasource.DataSourceWithConfigure = &directoryServiceRoleDataSource{}

type directoryServiceRoleDataSource struct{ client *client.FlashBladeClient }

func NewDirectoryServiceRoleDataSource() datasource.DataSource {
	return &directoryServiceRoleDataSource{}
}

type directoryServiceRoleDataSourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Group                    types.String `tfsdk:"group"`
	GroupBase                types.String `tfsdk:"group_base"`
	ManagementAccessPolicies types.List   `tfsdk:"management_access_policies"`
	Role                     types.Object `tfsdk:"role"`
}

func (d *directoryServiceRoleDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_directory_service_role"
}

func (d *directoryServiceRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing FlashBlade directory service role by name.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"name":       schema.StringAttribute{Required: true, Description: "The role name to look up."},
			"group":      schema.StringAttribute{Computed: true},
			"group_base": schema.StringAttribute{Computed: true},
			"management_access_policies": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"role": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{Computed: true},
				},
			},
		},
	}
}

func (d *directoryServiceRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.FlashBladeClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected *client.FlashBladeClient, got: %T.", req.ProviderData))
		return
	}
	d.client = c
}

func (d *directoryServiceRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config directoryServiceRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	role, err := d.client.GetDirectoryServiceRole(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError("Directory service role not found", fmt.Sprintf("No directory service role named %q.", name))
		} else {
			resp.Diagnostics.AddError("Error reading directory service role", err.Error())
		}
		return
	}

	config.ID = types.StringValue(role.ID)
	config.Name = types.StringValue(role.Name)
	config.Group = types.StringValue(role.Group)
	config.GroupBase = types.StringValue(role.GroupBase)

	policyVals := make([]attr.Value, len(role.ManagementAccessPolicies))
	for i, p := range role.ManagementAccessPolicies {
		policyVals[i] = types.StringValue(p.Name)
	}
	list, diags := types.ListValue(types.StringType, policyVals)
	resp.Diagnostics.Append(diags...)
	config.ManagementAccessPolicies = list

	if role.Role == nil {
		config.Role = types.ObjectNull(map[string]attr.Type{"name": types.StringType})
	} else {
		obj, diags := types.ObjectValue(
			map[string]attr.Type{"name": types.StringType},
			map[string]attr.Value{"name": types.StringValue(role.Role.Name)},
		)
		resp.Diagnostics.Append(diags...)
		config.Role = obj
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

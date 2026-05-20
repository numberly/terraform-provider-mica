package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ datasource.DataSource = &resiliencyGroupMemberDataSource{}
var _ datasource.DataSourceWithConfigure = &resiliencyGroupMemberDataSource{}

// resiliencyGroupMemberDataSource implements the flashblade_resiliency_group_member
// data source. It looks up a single membership row by (group_name, member_name).
type resiliencyGroupMemberDataSource struct {
	client *client.FlashBladeClient
}

func NewResiliencyGroupMemberDataSource() datasource.DataSource {
	return &resiliencyGroupMemberDataSource{}
}

// ---------- model struct ----------------------------------------------------

// resiliencyGroupMemberDataSourceModel mirrors the data source schema.
// ID is a composite identifier "<group_name>/<member_name>" so the row is
// uniquely addressable in Terraform state (member rows have no global name).
type resiliencyGroupMemberDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ResiliencyGroupName types.String `tfsdk:"resiliency_group_name"`
	MemberName          types.String `tfsdk:"member_name"`
	GroupID             types.String `tfsdk:"group_id"`
	GroupResourceType   types.String `tfsdk:"group_resource_type"`
	MemberID            types.String `tfsdk:"member_id"`
	MemberResourceType  types.String `tfsdk:"member_resource_type"`
}

// ---------- data source interface methods -----------------------------------

func (d *resiliencyGroupMemberDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_resiliency_group_member"
}

func (d *resiliencyGroupMemberDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a single FlashBlade resiliency group member by parent group name and member name. " +
			"Member rows are hardware-managed group-to-resource associations reported by the array (API 2.23+).",
		MarkdownDescription: "Reads a single FlashBlade resiliency group member by parent group name and member name. " +
			"Member rows are hardware-managed group-to-resource associations reported by the array (API 2.23+).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form `<resiliency_group_name>/<member_name>`.",
			},
			"resiliency_group_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the parent resiliency group whose member rows should be searched.",
			},
			"member_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the member resource to look up within the resiliency group.",
			},
			"group_id": schema.StringAttribute{
				Computed:    true,
				Description: "Non-modifiable globally unique identifier of the parent resiliency group.",
			},
			"group_resource_type": schema.StringAttribute{
				Computed:    true,
				Description: "API resource type of the parent group reference (typically `resiliency-groups`).",
			},
			"member_id": schema.StringAttribute{
				Computed:    true,
				Description: "Non-modifiable globally unique identifier of the member resource.",
			},
			"member_resource_type": schema.StringAttribute{
				Computed:    true,
				Description: "API resource type of the member reference (e.g. `file-systems`).",
			},
		},
	}
}

func (d *resiliencyGroupMemberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *resiliencyGroupMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resiliencyGroupMemberDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := config.ResiliencyGroupName.ValueString()
	memberName := config.MemberName.ValueString()

	m, err := d.client.GetResiliencyGroupMember(ctx, groupName, memberName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Resiliency group member not found",
				fmt.Sprintf("No member named %q was found in resiliency group %q.", memberName, groupName),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading resiliency group member", err.Error())
		return
	}

	config.ID = types.StringValue(groupName + "/" + memberName)
	config.ResiliencyGroupName = types.StringValue(m.Group.Name)
	config.MemberName = types.StringValue(m.Member.Name)
	config.GroupID = stringOrNull(m.Group.ID)
	config.GroupResourceType = stringOrNull(m.Group.ResourceType)
	config.MemberID = stringOrNull(m.Member.ID)
	config.MemberResourceType = stringOrNull(m.Member.ResourceType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

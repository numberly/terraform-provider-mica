package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &subnetResource{}
var _ resource.ResourceWithConfigure = &subnetResource{}
var _ resource.ResourceWithImportState = &subnetResource{}
var _ resource.ResourceWithUpgradeState = &subnetResource{}

// subnetResource implements the flashblade_subnet resource.
type subnetResource struct {
	client *client.FlashBladeClient
}

func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

// ---------- model structs ----------------------------------------------------

// subnetResourceModel is the top-level model for the flashblade_subnet resource.
type subnetResourceModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Prefix     types.String   `tfsdk:"prefix"`
	Gateway    types.String   `tfsdk:"gateway"`
	MTU        types.Int64    `tfsdk:"mtu"`
	VLAN       types.Int64    `tfsdk:"vlan"`
	LagName    types.String   `tfsdk:"lag_name"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	Services   types.List     `tfsdk:"services"`
	Interfaces types.List     `tfsdk:"interfaces"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// subnetV0Model is the v0 state model. Identical attribute set to the current model —
// the v0→v1 bump is defensive (no on-disk shape change; only wire-format semantics changed
// for VLAN and LinkAggregationGroup). See CONVENTIONS.md §State Upgraders and D-51-04.
type subnetV0Model struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Prefix     types.String   `tfsdk:"prefix"`
	Gateway    types.String   `tfsdk:"gateway"`
	MTU        types.Int64    `tfsdk:"mtu"`
	VLAN       types.Int64    `tfsdk:"vlan"`
	LagName    types.String   `tfsdk:"lag_name"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	Services   types.List     `tfsdk:"services"`
	Interfaces types.List     `tfsdk:"interfaces"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *subnetResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_subnet"
}

// Schema defines the resource schema.
func (r *subnetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a FlashBlade subnet.",
		MarkdownDescription: `Manages a FlashBlade subnet.

## Example Usage

` + "```hcl" + `
resource "flashblade_subnet" "data" {
  name     = "data-subnet"
  prefix   = "10.21.200.0/24"
  gateway  = "10.21.200.1"
  mtu      = 9000
  vlan     = 2100
  lag_name = "lag0"
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the subnet.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the subnet. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				Required:    true,
				Description: "IPv4 or IPv6 subnet address in CIDR notation (e.g. 10.21.200.0/24).",
			},
			"gateway": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IPv4 or IPv6 gateway address for the subnet.",
			},
			"mtu": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum transmission unit (MTU) in bytes. Defaults to 1500.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"vlan": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "VLAN ID. 0 means untagged. Defaults to 0.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"lag_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the link aggregation group (LAG) this subnet is attached to.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the subnet is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of services associated with this subnet (e.g. data, replication).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"interfaces": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of network interface names attached to this subnet.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
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
// v0→v1: no-op identity — the Terraform attribute set is unchanged. The bump exists
// because wire-format semantics changed (VLAN *int64, LinkAggregationGroup **NamedReference
// in PATCH). See D-51-03 and D-51-04.
func (r *subnetResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Version:     0,
				Description: "Manages a FlashBlade subnet.",
				Attributes: map[string]schema.Attribute{
					"id":       schema.StringAttribute{Computed: true},
					"name":     schema.StringAttribute{Required: true},
					"prefix":   schema.StringAttribute{Required: true},
					"gateway":  schema.StringAttribute{Optional: true, Computed: true},
					"mtu":      schema.Int64Attribute{Optional: true, Computed: true},
					"vlan":     schema.Int64Attribute{Optional: true, Computed: true},
					"lag_name": schema.StringAttribute{Optional: true, Computed: true},
					"enabled":  schema.BoolAttribute{Computed: true},
					"services": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
					"interfaces": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
						Create: true,
						Read:   true,
						Update: true,
						Delete: true,
					}),
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState subnetV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Identity copy — no attribute shape change.
				newState := subnetResourceModel(oldState)
				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *subnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data subnetResourceModel
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

	body := client.SubnetPost{
		Prefix:               data.Prefix.ValueString(),
		Gateway:              data.Gateway.ValueString(),
		LinkAggregationGroup: lagNameToRef(data.LagName),
	}
	if !data.MTU.IsNull() && !data.MTU.IsUnknown() {
		body.MTU = data.MTU.ValueInt64()
	}
	if !data.VLAN.IsNull() && !data.VLAN.IsUnknown() {
		// Send VLAN as an explicit *int64 so VLAN=0 (untagged) is preserved
		// in the POST body instead of being dropped by omitempty (R-001).
		v := data.VLAN.ValueInt64()
		body.VLAN = &v
	}

	subnet, err := r.client.PostSubnet(ctx, data.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating subnet", err.Error())
		return
	}

	resp.Diagnostics.Append(mapSubnetToModel(ctx, subnet, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data subnetResourceModel
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
	subnet, err := r.client.GetSubnet(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading subnet", err.Error())
		return
	}

	// Log drift if key values differ from state.
	if data.Prefix.ValueString() != subnet.Prefix {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "prefix",
			"was": data.Prefix.ValueString(), "now": subnet.Prefix,
		})
	}
	if data.Gateway.ValueString() != subnet.Gateway {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "gateway",
			"was": data.Gateway.ValueString(), "now": subnet.Gateway,
		})
	}
	if data.MTU.ValueInt64() != subnet.MTU {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name, "field": "mtu",
			"was": data.MTU.ValueInt64(), "now": subnet.MTU,
		})
	}

	resp.Diagnostics.Append(mapSubnetToModel(ctx, subnet, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing subnet.
func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state subnetResourceModel
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

	patch := client.SubnetPatch{}

	if !plan.Prefix.Equal(state.Prefix) {
		v := plan.Prefix.ValueString()
		patch.Prefix = &v
	}
	if !plan.Gateway.Equal(state.Gateway) {
		v := plan.Gateway.ValueString()
		patch.Gateway = &v
	}
	if !plan.MTU.Equal(state.MTU) {
		v := plan.MTU.ValueInt64()
		patch.MTU = &v
	}
	if !plan.VLAN.Equal(state.VLAN) {
		v := plan.VLAN.ValueInt64()
		patch.VLAN = &v
	}
	// Use doublePointerRefForPatch so that:
	//   - unchanged lag_name         → omit (nil outer)
	//   - lag_name set → null        → CLEAR (non-nil outer, nil inner)
	//   - lag_name changed / set     → SET
	// R-002, D-51-01.
	patch.LinkAggregationGroup = doublePointerRefForPatch(state.LagName, plan.LagName)

	subnet, err := r.client.PatchSubnet(ctx, plan.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating subnet", err.Error())
		return
	}

	resp.Diagnostics.Append(mapSubnetToModel(ctx, subnet, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a subnet.
func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data subnetResourceModel
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

	if err := r.client.DeleteSubnet(ctx, data.Name.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting subnet", err.Error())
		return
	}
}

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data subnetResourceModel
	data.Timeouts = nullTimeoutsValue()
	data.Name = types.StringValue(name)

	subnet, err := r.client.GetSubnet(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing subnet", err.Error())
		return
	}

	resp.Diagnostics.Append(mapSubnetToModel(ctx, subnet, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// lagNameToRef converts a flat types.String lag name to a *client.NamedReference.
// Returns nil if the value is null, unknown, or empty.
func lagNameToRef(lagName types.String) *client.NamedReference {
	if lagName.IsNull() || lagName.IsUnknown() || lagName.ValueString() == "" {
		return nil
	}
	return &client.NamedReference{Name: lagName.ValueString()}
}

// refToLagName converts a *client.NamedReference to a flat types.String.
// Returns types.StringNull() if the reference is nil or has an empty name.
func refToLagName(ref *client.NamedReference) types.String {
	if ref == nil || ref.Name == "" {
		return types.StringNull()
	}
	return types.StringValue(ref.Name)
}

// mapSubnetToModel maps a client.Subnet response to a subnetResourceModel.
// It preserves user-managed fields (Timeouts).
func mapSubnetToModel(ctx context.Context, subnet *client.Subnet, data *subnetResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ID = types.StringValue(subnet.ID)
	data.Name = types.StringValue(subnet.Name)
	data.Prefix = types.StringValue(subnet.Prefix)
	data.Gateway = stringOrNull(subnet.Gateway)
	data.MTU = types.Int64Value(subnet.MTU)
	data.VLAN = types.Int64Value(subnet.VLAN)
	data.Enabled = types.BoolValue(subnet.Enabled)
	data.LagName = refToLagName(subnet.LinkAggregationGroup)

	if len(subnet.Services) > 0 {
		svcList, svcDiags := types.ListValueFrom(ctx, types.StringType, subnet.Services)
		diags.Append(svcDiags...)
		if diags.HasError() {
			return diags
		}
		data.Services = svcList
	} else {
		data.Services = types.ListNull(types.StringType)
	}

	if len(subnet.Interfaces) == 0 {
		data.Interfaces = types.ListNull(types.StringType)
	} else {
		data.Interfaces = namedRefsToListValue(subnet.Interfaces)
	}
	return diags
}

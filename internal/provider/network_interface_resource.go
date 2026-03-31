package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure networkInterfaceResource satisfies the resource interfaces.
var _ resource.Resource = &networkInterfaceResource{}
var _ resource.ResourceWithConfigure = &networkInterfaceResource{}
var _ resource.ResourceWithImportState = &networkInterfaceResource{}
var _ resource.ResourceWithConfigValidators = &networkInterfaceResource{}

// networkInterfaceResource implements the flashblade_network_interface resource.
type networkInterfaceResource struct {
	client *client.FlashBladeClient
}

// NewNetworkInterfaceResource is the factory function registered in the provider.
func NewNetworkInterfaceResource() resource.Resource {
	return &networkInterfaceResource{}
}

// ---------- model structs ----------------------------------------------------

// networkInterfaceResourceModel is the top-level model for the flashblade_network_interface resource.
type networkInterfaceResourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Address         types.String   `tfsdk:"address"`
	SubnetName      types.String   `tfsdk:"subnet_name"`
	Type            types.String   `tfsdk:"type"`
	Services        types.String   `tfsdk:"services"`
	AttachedServers types.List     `tfsdk:"attached_servers"`
	Enabled         types.Bool     `tfsdk:"enabled"`
	Gateway         types.String   `tfsdk:"gateway"`
	MTU             types.Int64    `tfsdk:"mtu"`
	Netmask         types.String   `tfsdk:"netmask"`
	VLAN            types.Int64    `tfsdk:"vlan"`
	Realms          types.List     `tfsdk:"realms"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *networkInterfaceResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_network_interface"
}

// Schema defines the resource schema.
func (r *networkInterfaceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade network interface (VIP).",
		MarkdownDescription: `Manages a FlashBlade network interface (VIP).

## Example Usage

` + "```hcl" + `
resource "flashblade_network_interface" "data_vip" {
  name        = "vip0"
  address     = "10.21.200.10"
  subnet_name = "data-subnet"
  type        = "vip"
  services    = "data"
  attached_servers = ["server1"]
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the network interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network interface. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the subnet this interface is attached to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The network interface type (e.g. vip). Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"services": schema.StringAttribute{
				Required:    true,
				Description: "The service type for this network interface. One of: data, sts, egress-only, replication.",
				Validators: []validator.String{
					serviceTypeValidator(),
				},
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "The IPv4 address for this network interface.",
			},
			"attached_servers": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of server names attached to this interface. Required for data/sts; forbidden for egress-only/replication.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the network interface is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway": schema.StringAttribute{
				Computed:    true,
				Description: "The gateway address for this network interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mtu": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum transmission unit (MTU) in bytes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"netmask": schema.StringAttribute{
				Computed:    true,
				Description: "The subnet mask for this network interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vlan": schema.Int64Attribute{
				Computed:    true,
				Description: "VLAN ID. 0 means untagged.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"realms": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of realms associated with this network interface.",
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

// Configure injects the FlashBladeClient into the resource.
func (r *networkInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ConfigValidators returns cross-field validators for the resource configuration.
func (r *networkInterfaceResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{networkInterfaceServicesValidator{}}
}

// ---------- CRUD methods ----------------------------------------------------

// Create creates a new network interface.
func (r *networkInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkInterfaceResourceModel
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

	body := client.NetworkInterfacePost{
		Address:  data.Address.ValueString(),
		Services: []string{data.Services.ValueString()},
		Type:     data.Type.ValueString(),
	}
	if !data.SubnetName.IsNull() && !data.SubnetName.IsUnknown() && data.SubnetName.ValueString() != "" {
		body.Subnet = &client.NamedReference{Name: data.SubnetName.ValueString()}
	}
	body.AttachedServers = niServersToNamedRefs(ctx, data.AttachedServers, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ni, err := r.client.PostNetworkInterface(ctx, data.Name.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating network interface", err.Error())
		return
	}

	mapNetworkInterfaceToModel(ctx, ni, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *networkInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkInterfaceResourceModel
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
	ni, err := r.client.GetNetworkInterface(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network interface", err.Error())
		return
	}

	// Log drift on mutable fields.
	if data.Address.ValueString() != ni.Address {
		tflog.Info(ctx, "network interface drift detected: address changed",
			map[string]any{"name": name, "state": data.Address.ValueString(), "api": ni.Address})
	}
	apiSvc := ""
	if len(ni.Services) > 0 {
		apiSvc = ni.Services[0]
	}
	if data.Services.ValueString() != apiSvc {
		tflog.Info(ctx, "network interface drift detected: services changed",
			map[string]any{"name": name, "state": data.Services.ValueString(), "api": apiSvc})
	}

	mapNetworkInterfaceToModel(ctx, ni, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing network interface.
func (r *networkInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state networkInterfaceResourceModel
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

	patch := client.NetworkInterfacePatch{
		// Always include services and attached_servers — full-replace semantics (no omitempty).
		Services:        []string{plan.Services.ValueString()},
		AttachedServers: niServersToNamedRefs(ctx, plan.AttachedServers, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Only include address when it has changed.
	if !plan.Address.Equal(state.Address) {
		v := plan.Address.ValueString()
		patch.Address = &v
	}

	ni, err := r.client.PatchNetworkInterface(ctx, plan.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating network interface", err.Error())
		return
	}

	mapNetworkInterfaceToModel(ctx, ni, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a network interface.
func (r *networkInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkInterfaceResourceModel
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

	if err := r.client.DeleteNetworkInterface(ctx, data.Name.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting network interface", err.Error())
		return
	}
}

// ImportState imports an existing network interface by name.
func (r *networkInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data networkInterfaceResourceModel
	data.Timeouts = nullTimeoutsValue()
	data.Name = types.StringValue(name)

	ni, err := r.client.GetNetworkInterface(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing network interface", err.Error())
		return
	}

	mapNetworkInterfaceToModel(ctx, ni, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapNetworkInterfaceToModel maps a *client.NetworkInterface to a *networkInterfaceResourceModel.
func mapNetworkInterfaceToModel(ctx context.Context, ni *client.NetworkInterface, data *networkInterfaceResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(ni.ID)
	data.Name = types.StringValue(ni.Name)
	data.Address = types.StringValue(ni.Address)
	data.Type = types.StringValue(ni.Type)
	data.Enabled = types.BoolValue(ni.Enabled)
	data.MTU = types.Int64Value(ni.MTU)
	data.VLAN = types.Int64Value(ni.VLAN)
	data.Gateway = stringOrNull(ni.Gateway)
	data.Netmask = stringOrNull(ni.Netmask)
	data.SubnetName = refToSubnetName(ni.Subnet)

	// Collapse []string services to a single string value.
	if len(ni.Services) > 0 {
		data.Services = types.StringValue(ni.Services[0])
	} else {
		data.Services = types.StringNull()
	}

	// AttachedServers: always use a list (empty, not null) to prevent spurious drift.
	if len(ni.AttachedServers) > 0 {
		names := make([]string, 0, len(ni.AttachedServers))
		for _, s := range ni.AttachedServers {
			names = append(names, s.Name)
		}
		serverList, serverDiags := types.ListValueFrom(ctx, types.StringType, names)
		diags.Append(serverDiags...)
		if diags.HasError() {
			return
		}
		data.AttachedServers = serverList
	} else {
		emptyList, emptyDiags := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(emptyDiags...)
		if diags.HasError() {
			return
		}
		data.AttachedServers = emptyList
	}

	// Realms: null when absent, list otherwise.
	if len(ni.Realms) > 0 {
		realmList, realmDiags := types.ListValueFrom(ctx, types.StringType, ni.Realms)
		diags.Append(realmDiags...)
		if diags.HasError() {
			return
		}
		data.Realms = realmList
	} else {
		data.Realms = types.ListNull(types.StringType)
	}
}

// refToSubnetName converts a *client.NamedReference to a flat types.String.
// Returns types.StringNull() if the reference is nil or has an empty name.
func refToSubnetName(ref *client.NamedReference) types.String {
	if ref == nil || ref.Name == "" {
		return types.StringNull()
	}
	return types.StringValue(ref.Name)
}

// niServersToNamedRefs converts a types.List of server names to []client.NamedReference.
// Returns nil if the list is null, unknown, or empty.
func niServersToNamedRefs(ctx context.Context, list types.List, diags *diag.Diagnostics) []client.NamedReference {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var names []string
	diags.Append(list.ElementsAs(ctx, &names, false)...)
	if diags.HasError() {
		return nil
	}
	if len(names) == 0 {
		return nil
	}
	refs := make([]client.NamedReference, 0, len(names))
	for _, n := range names {
		refs = append(refs, client.NamedReference{Name: n})
	}
	return refs
}

// ---------- validators -------------------------------------------------------

// serviceTypeValidator returns a validator.String that accepts exactly:
// "data", "sts", "egress-only", "replication".
func serviceTypeValidator() validator.String {
	return &niServiceTypeValidator{}
}

// niServiceTypeValidator validates the services attribute of a network interface.
type niServiceTypeValidator struct{}

func (v *niServiceTypeValidator) Description(_ context.Context) string {
	return `value must be one of: "data", "sts", "egress-only", "replication"`
}

func (v *niServiceTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *niServiceTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	valid := map[string]bool{
		"data":        true,
		"sts":         true,
		"egress-only": true,
		"replication": true,
	}
	if !valid[val] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Service Type",
			fmt.Sprintf("Expected one of: data, sts, egress-only, replication. Got: %q", val),
		)
	}
}

// ---------- cross-field config validator ------------------------------------

// networkInterfaceServicesValidator enforces the relationship between services and attached_servers:
//   - data/sts: requires exactly 1 attached_server
//   - egress-only/replication: requires attached_servers to be null or empty
type networkInterfaceServicesValidator struct{}

func (v networkInterfaceServicesValidator) Description(_ context.Context) string {
	return "data/sts services require attached_servers; egress-only/replication forbid attached_servers"
}

func (v networkInterfaceServicesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v networkInterfaceServicesValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Read each attribute individually to avoid crashing on unknown list elements.
	var services types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("services"), &services)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var attachedServers types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("attached_servers"), &attachedServers)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Defer validation if either value is unknown (not yet known at plan time).
	if services.IsUnknown() || attachedServers.IsUnknown() {
		return
	}

	svc := services.ValueString()
	requiresServers := svc == "data" || svc == "sts"
	forbidsServers := svc == "egress-only" || svc == "replication"

	// Count elements without converting to native strings (elements may be individually unknown).
	hasServers := !attachedServers.IsNull() && len(attachedServers.Elements()) > 0

	if requiresServers && !hasServers {
		resp.Diagnostics.AddError(
			"Missing attached_servers",
			fmt.Sprintf("services=%q requires at least one attached_server. Set attached_servers to a list with at least one server name.", svc),
		)
	}

	if forbidsServers && hasServers {
		resp.Diagnostics.AddError(
			"Invalid attached_servers",
			fmt.Sprintf("services=%q does not allow attached_servers. Remove attached_servers from the configuration.", svc),
		)
	}
}

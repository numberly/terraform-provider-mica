package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure arrayDnsResource satisfies the resource interfaces.
var _ resource.Resource = &arrayDnsResource{}
var _ resource.ResourceWithConfigure = &arrayDnsResource{}
var _ resource.ResourceWithImportState = &arrayDnsResource{}
var _ resource.ResourceWithUpgradeState = &arrayDnsResource{}

// arrayDnsResource implements the flashblade_array_dns resource.
type arrayDnsResource struct {
	client *client.FlashBladeClient
}

// NewArrayDnsResource is the factory function registered in the provider.
func NewArrayDnsResource() resource.Resource {
	return &arrayDnsResource{}
}

// ---------- model structs ----------------------------------------------------

// arrayDnsModel is the top-level model for the flashblade_array_dns resource (v1).
type arrayDnsModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Domain      types.String   `tfsdk:"domain"`
	Nameservers types.List     `tfsdk:"nameservers"`
	Services    types.List     `tfsdk:"services"`
	Sources     types.List     `tfsdk:"sources"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

// arrayDnsV0Model is the v0 state model (no name attribute).
type arrayDnsV0Model struct {
	ID          types.String   `tfsdk:"id"`
	Domain      types.String   `tfsdk:"domain"`
	Nameservers types.List     `tfsdk:"nameservers"`
	Services    types.List     `tfsdk:"services"`
	Sources     types.List     `tfsdk:"sources"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *arrayDnsResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_array_dns"
}

// Schema defines the resource schema.
func (r *arrayDnsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a DNS configuration entry on a FlashBlade array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the DNS configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the DNS configuration. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The domain suffix appended by the array to unqualified hostnames.",
			},
			"nameservers": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of DNS server IP addresses.",
			},
			"services": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Services that use this DNS configuration.",
			},
			"sources": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Network interfaces used for DNS traffic.",
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
func (r *arrayDnsResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// v0 -> v1: add name attribute.
		0: {
			PriorSchema: &schema.Schema{
				Version:     0,
				Description: "Manages the DNS configuration of a FlashBlade array. This is a singleton resource — Create/Delete patches the existing configuration rather than creating or deleting a record.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"domain": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					"nameservers": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"services": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"sources": schema.ListAttribute{
						Optional:    true,
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
				var oldState arrayDnsV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				newState := arrayDnsModel{
					ID:          oldState.ID,
					Name:        types.StringValue("default"),
					Domain:      oldState.Domain,
					Nameservers: oldState.Nameservers,
					Services:    oldState.Services,
					Sources:     oldState.Sources,
					Timeouts:    oldState.Timeouts,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *arrayDnsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new DNS configuration entry.
func (r *arrayDnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data arrayDnsModel
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

	name := data.Name.ValueString()

	nameservers, d := listToStrings(ctx, data.Nameservers)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	services, d := listToStrings(ctx, data.Services)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	sources, d := listToStrings(ctx, data.Sources)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	post := client.ArrayDnsPost{
		Domain:      data.Domain.ValueString(),
		Nameservers: nameservers,
		Services:    services,
		Sources:     sources,
	}
	result, err := r.client.PostArrayDns(ctx, name, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating array DNS configuration", err.Error())
		return
	}

	resp.Diagnostics.Append(mapArrayDnsToModel(ctx, result, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *arrayDnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data arrayDnsModel
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
	dns, err := r.client.GetArrayDns(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading array DNS configuration", err.Error())
		return
	}

	// Drift detection
	if data.Name.ValueString() != dns.Name {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "name",
			"was":      data.Name.ValueString(),
			"now":      dns.Name,
		})
	}
	if data.Domain.ValueString() != dns.Domain {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "domain",
			"was":      data.Domain.ValueString(),
			"now":      dns.Domain,
		})
	}

	resp.Diagnostics.Append(mapArrayDnsToModel(ctx, dns, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to the DNS configuration.
func (r *arrayDnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state arrayDnsModel
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

	name := plan.Name.ValueString()

	patch := client.ArrayDnsPatch{}
	if !plan.Domain.Equal(state.Domain) {
		v := plan.Domain.ValueString()
		patch.Domain = &v
	}
	if !plan.Nameservers.Equal(state.Nameservers) {
		ns, d := listToStrings(ctx, plan.Nameservers)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Nameservers = &ns
	}
	if !plan.Services.Equal(state.Services) {
		svc, d := listToStrings(ctx, plan.Services)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Services = &svc
	}
	if !plan.Sources.Equal(state.Sources) {
		src, d := listToStrings(ctx, plan.Sources)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Sources = &src
	}

	result, err := r.client.PatchArrayDns(ctx, name, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating array DNS configuration", err.Error())
		return
	}

	resp.Diagnostics.Append(mapArrayDnsToModel(ctx, result, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the DNS configuration entry.
func (r *arrayDnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data arrayDnsModel
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

	name := data.Name.ValueString()
	if err := r.client.DeleteArrayDns(ctx, name); err != nil {
		resp.Diagnostics.AddError("Error deleting array DNS configuration", err.Error())
		return
	}

	tflog.Info(ctx, "DNS configuration deleted", map[string]any{"name": name})
}

// ImportState imports a DNS configuration entry by name.
func (r *arrayDnsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	dns, err := r.client.GetArrayDns(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing array DNS configuration", err.Error())
		return
	}

	var data arrayDnsModel
	data.Timeouts = nullTimeoutsValue()

	resp.Diagnostics.Append(mapArrayDnsToModel(ctx, dns, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapArrayDnsToModel maps a client.ArrayDns to an arrayDnsModel.
// Returns any diagnostics generated by the framework list conversion.
func mapArrayDnsToModel(ctx context.Context, dns *client.ArrayDns, data *arrayDnsModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(dns.ID)
	data.Name = types.StringValue(dns.Name)
	data.Domain = types.StringValue(dns.Domain)

	if len(dns.Nameservers) > 0 {
		ns, d := types.ListValueFrom(ctx, types.StringType, dns.Nameservers)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Nameservers = ns
	} else {
		data.Nameservers = emptyStringList()
	}

	if len(dns.Services) > 0 {
		svc, d := types.ListValueFrom(ctx, types.StringType, dns.Services)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Services = svc
	} else {
		data.Services = emptyStringList()
	}

	if len(dns.Sources) > 0 {
		src, d := types.ListValueFrom(ctx, types.StringType, dns.Sources)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Sources = src
	} else {
		data.Sources = emptyStringList()
	}

	return diags
}

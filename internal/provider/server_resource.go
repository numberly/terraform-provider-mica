package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure serverResource satisfies the resource interfaces.
var _ resource.Resource = &serverResource{}
var _ resource.ResourceWithConfigure = &serverResource{}
var _ resource.ResourceWithImportState = &serverResource{}
var _ resource.ResourceWithUpgradeState = &serverResource{}

// serverResource implements the flashblade_server resource.
type serverResource struct {
	client *client.FlashBladeClient
}

// NewServerResource is the factory function registered in the provider.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

// ---------- model structs ----------------------------------------------------

// serverDNSModel maps a single DNS configuration block.
type serverDNSModel struct {
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.List   `tfsdk:"nameservers"`
	Services    types.List   `tfsdk:"services"`
}

// serverResourceModel is the top-level model for the flashblade_server resource.
type serverResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Created       types.Int64    `tfsdk:"created"`
	DNS           types.List     `tfsdk:"dns"`
	CascadeDelete types.List     `tfsdk:"cascade_delete"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

// ---------- attribute type helpers -------------------------------------------

// serverDNSAttrTypes returns the attribute types for a single DNS nested object.
func serverDNSAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":      types.StringType,
		"nameservers": types.ListType{ElemType: types.StringType},
		"services":    types.ListType{ElemType: types.StringType},
	}
}

// serverDNSObjectType returns the types.ObjectType for a DNS nested object.
func serverDNSObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: serverDNSAttrTypes()}
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *serverResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_server"
}

// Schema defines the resource schema.
func (r *serverResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the server. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the server was created.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"dns": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "DNS configuration for the server.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Optional:    true,
							Description: "DNS domain suffix.",
						},
						"nameservers": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "List of DNS nameserver IP addresses.",
						},
						"services": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "List of DNS service types.",
						},
					},
				},
			},
			"cascade_delete": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of export names to cascade-delete when destroying this server. Used only on delete, not stored in API state.",
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
func (r *serverResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new server.
func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverResourceModel
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

	post := client.ServerPost{
		DNS: mapModelDNSToClient(ctx, &data, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	srv, err := r.client.PostServer(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating server", err.Error())
		return
	}

	mapServerToModel(ctx, srv, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverResourceModel
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
	srv, err := r.client.GetServer(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading server", err.Error())
		return
	}

	mapServerToModel(ctx, srv, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing server.
func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	patch := client.ServerPatch{
		DNS: mapModelDNSToClient(ctx, &plan, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	srv, err := r.client.PatchServer(ctx, plan.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating server", err.Error())
		return
	}

	mapServerToModel(ctx, srv, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a server.
func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverResourceModel
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

	// Extract cascade_delete names from state.
	var cascadeNames []string
	if !data.CascadeDelete.IsNull() && !data.CascadeDelete.IsUnknown() {
		resp.Diagnostics.Append(data.CascadeDelete.ElementsAs(ctx, &cascadeNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if err := r.client.DeleteServer(ctx, data.Name.ValueString(), cascadeNames); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting server", err.Error())
		return
	}
}

// ImportState imports an existing server by name.
func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data serverResourceModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	// Initialize cascade_delete as null (write-only, never comes from API).
	data.CascadeDelete = types.ListNull(types.StringType)

	data.Name = types.StringValue(name)

	srv, err := r.client.GetServer(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing server", err.Error())
		return
	}

	mapServerToModel(ctx, srv, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapServerToModel maps a client.Server to a serverResourceModel.
// It preserves user-managed fields (Timeouts, CascadeDelete).
func mapServerToModel(ctx context.Context, srv *client.Server, data *serverResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(srv.ID)
	data.Name = types.StringValue(srv.Name)
	data.Created = types.Int64Value(srv.Created)

	// Map DNS list.
	if len(srv.DNS) > 0 {
		dnsObjs := make([]attr.Value, 0, len(srv.DNS))
		for _, d := range srv.DNS {
			// Build nameservers list.
			var nameservers types.List
			if len(d.Nameservers) > 0 {
				ns, nsDiags := types.ListValueFrom(ctx, types.StringType, d.Nameservers)
				diags.Append(nsDiags...)
				if diags.HasError() {
					return
				}
				nameservers = ns
			} else {
				nameservers = types.ListNull(types.StringType)
			}

			// Build services list.
			var services types.List
			if len(d.Services) > 0 {
				svc, svcDiags := types.ListValueFrom(ctx, types.StringType, d.Services)
				diags.Append(svcDiags...)
				if diags.HasError() {
					return
				}
				services = svc
			} else {
				services = types.ListNull(types.StringType)
			}

			obj, objDiags := types.ObjectValue(serverDNSAttrTypes(), map[string]attr.Value{
				"domain":      types.StringValue(d.Domain),
				"nameservers": nameservers,
				"services":    services,
			})
			diags.Append(objDiags...)
			if diags.HasError() {
				return
			}
			dnsObjs = append(dnsObjs, obj)
		}

		dnsList, listDiags := types.ListValue(serverDNSObjectType(), dnsObjs)
		diags.Append(listDiags...)
		if diags.HasError() {
			return
		}
		data.DNS = dnsList
	} else {
		data.DNS = types.ListNull(serverDNSObjectType())
	}
}

// mapModelDNSToClient extracts DNS from the Terraform model and converts to client types.
func mapModelDNSToClient(ctx context.Context, data *serverResourceModel, diags *diag.Diagnostics) []client.ServerDNS {
	if data.DNS.IsNull() || data.DNS.IsUnknown() || len(data.DNS.Elements()) == 0 {
		return nil
	}

	var dnsModels []serverDNSModel
	d := data.DNS.ElementsAs(ctx, &dnsModels, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	result := make([]client.ServerDNS, 0, len(dnsModels))
	for _, dm := range dnsModels {
		entry := client.ServerDNS{
			Domain: dm.Domain.ValueString(),
		}

		if !dm.Nameservers.IsNull() && !dm.Nameservers.IsUnknown() {
			diags.Append(dm.Nameservers.ElementsAs(ctx, &entry.Nameservers, false)...)
			if diags.HasError() {
				return nil
			}
		}

		if !dm.Services.IsNull() && !dm.Services.IsUnknown() {
			diags.Append(dm.Services.ElementsAs(ctx, &entry.Services, false)...)
			if diags.HasError() {
				return nil
			}
		}

		result = append(result, entry)
	}

	return result
}

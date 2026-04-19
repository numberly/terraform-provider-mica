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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &serverResource{}
var _ resource.ResourceWithConfigure = &serverResource{}
var _ resource.ResourceWithImportState = &serverResource{}
var _ resource.ResourceWithUpgradeState = &serverResource{}

// serverResource implements the flashblade_server resource.
type serverResource struct {
	client *client.FlashBladeClient
}

func NewServerResource() resource.Resource {
	return &serverResource{}
}

// ---------- model structs ----------------------------------------------------

// serverResourceModel is the top-level model for the flashblade_server resource (schema v2).
type serverResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Created           types.Int64    `tfsdk:"created"`
	DNS               types.List     `tfsdk:"dns"`
	DirectoryServices types.List     `tfsdk:"directory_services"`
	CascadeDelete     types.List     `tfsdk:"cascade_delete"`
	NetworkInterfaces types.List     `tfsdk:"network_interfaces"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

// serverV0StateModel is used exclusively for the v0 -> v1 state upgrade.
type serverV0StateModel struct {
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Created       types.Int64    `tfsdk:"created"`
	DNS           types.List     `tfsdk:"dns"`
	CascadeDelete types.List     `tfsdk:"cascade_delete"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

// serverV1StateModel mirrors the v1 schema shape (nested DNS + network_interfaces, no directory_services).
// Used as the OUTPUT of v0->v1 and INPUT of v1->v2 upgraders.
type serverV1StateModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Created           types.Int64    `tfsdk:"created"`
	DNS               types.List     `tfsdk:"dns"`
	CascadeDelete     types.List     `tfsdk:"cascade_delete"`
	NetworkInterfaces types.List     `tfsdk:"network_interfaces"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *serverResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_server"
}

// Schema defines the resource schema (version 2).
func (r *serverResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     2,
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
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dns": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of DNS configuration names associated with this server.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"directory_services": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of directory service names associated with this server.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"cascade_delete": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of export names to cascade-delete when destroying this server. Used only on delete, not stored in API state.",
			},
			"network_interfaces": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of network interfaces (VIPs) attached to this server. Discovered automatically from the array.",
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
func (r *serverResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// v1 nested DNS attribute types (shared between v0 PriorSchema and v1 PriorSchema).
	v1NestedDNS := schema.ListNestedAttribute{
		Optional: true,
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"domain": schema.StringAttribute{
					Optional: true,
				},
				"nameservers": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
				"services": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
	}

	return map[int64]resource.StateUpgrader{
		// v0 -> v1: add network_interfaces as empty list.
		0: {
			PriorSchema: &schema.Schema{
				Version:     0,
				Description: "Manages a FlashBlade server.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Required: true,
					},
					"created": schema.Int64Attribute{
						Computed: true,
					},
					"dns":            v1NestedDNS,
					"cascade_delete": schema.ListAttribute{Optional: true, ElementType: types.StringType},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
						Create: true,
						Read:   true,
						Update: true,
						Delete: true,
					}),
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var oldState serverV0StateModel
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Output v1 format: preserve nested DNS + add empty network_interfaces.
				newState := serverV1StateModel{
					ID:                oldState.ID,
					Name:              oldState.Name,
					Created:           oldState.Created,
					DNS:               oldState.DNS,
					CascadeDelete:     oldState.CascadeDelete,
					Timeouts:          oldState.Timeouts,
					NetworkInterfaces: types.ListValueMust(types.StringType, []attr.Value{}),
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},

		// v1 -> v2: convert nested DNS to flat string list, add directory_services.
		1: {
			PriorSchema: &schema.Schema{
				Version:     1,
				Description: "Manages a FlashBlade server.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Required: true,
					},
					"created": schema.Int64Attribute{
						Computed: true,
					},
					"dns":            v1NestedDNS,
					"cascade_delete": schema.ListAttribute{Optional: true, ElementType: types.StringType},
					"network_interfaces": schema.ListAttribute{
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
				var oldState serverV1StateModel
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// v1 DNS was nested objects without a "name" field — reset to null.
				// On next Read the provider will fetch real NamedReference data from the API.
				newState := serverResourceModel{
					ID:                oldState.ID,
					Name:              oldState.Name,
					Created:           oldState.Created,
					DNS:               types.ListNull(types.StringType),
					DirectoryServices: types.ListNull(types.StringType),
					CascadeDelete:     oldState.CascadeDelete,
					NetworkInterfaces: oldState.NetworkInterfaces,
					Timeouts:          oldState.Timeouts,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
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
		DNS: dnsNamesToRefs(ctx, &data, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	srv, err := r.client.PostServer(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating server", err.Error())
		return
	}

	resp.Diagnostics.Append(mapServerToModel(ctx, srv, &data)...)
	enrichServerNetworkInterfaces(ctx, r.client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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

	resp.Diagnostics.Append(mapServerToModel(ctx, srv, &data)...)
	enrichServerNetworkInterfaces(ctx, r.client, &data, &resp.Diagnostics)
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
		DNS: dnsNamesToRefs(ctx, &plan, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	srv, err := r.client.PatchServer(ctx, plan.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating server", err.Error())
		return
	}

	resp.Diagnostics.Append(mapServerToModel(ctx, srv, &plan)...)
	enrichServerNetworkInterfaces(ctx, r.client, &plan, &resp.Diagnostics)
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

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data serverResourceModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	// Initialize cascade_delete as null (write-only, never comes from API).
	data.CascadeDelete = types.ListNull(types.StringType)
	// Initialize network_interfaces as empty list (will be populated by mapServerToModel).
	data.NetworkInterfaces = types.ListValueMust(types.StringType, []attr.Value{})
	// Initialize directory_services as null (will be populated by mapServerToModel).
	data.DirectoryServices = types.ListNull(types.StringType)

	data.Name = types.StringValue(name)

	srv, err := r.client.GetServer(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing server", err.Error())
		return
	}

	resp.Diagnostics.Append(mapServerToModel(ctx, srv, &data)...)
	enrichServerNetworkInterfaces(ctx, r.client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapServerToModel is a pure transformer: it maps a client.Server to a
// serverResourceModel without performing any network I/O. Call
// enrichServerNetworkInterfaces separately to populate NetworkInterfaces
// from ListNetworkInterfaces. Preserves user-managed fields (Timeouts,
// CascadeDelete).
func mapServerToModel(ctx context.Context, srv *client.Server, data *serverResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.ID = types.StringValue(srv.ID)
	data.Name = types.StringValue(srv.Name)
	data.Created = types.Int64Value(srv.Created)

	if len(srv.DNS) > 0 {
		names := make([]string, len(srv.DNS))
		for i, d := range srv.DNS {
			names[i] = d.Name
		}
		dnsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
		diags.Append(listDiags...)
		if diags.HasError() {
			return diags
		}
		data.DNS = dnsList
	} else {
		data.DNS = types.ListNull(types.StringType)
	}

	if len(srv.DirectoryServices) > 0 {
		names := make([]string, len(srv.DirectoryServices))
		for i, ds := range srv.DirectoryServices {
			names[i] = ds.Name
		}
		dsList, listDiags := types.ListValueFrom(ctx, types.StringType, names)
		diags.Append(listDiags...)
		if diags.HasError() {
			return diags
		}
		data.DirectoryServices = dsList
	} else {
		data.DirectoryServices = types.ListNull(types.StringType)
	}
	return diags
}

// enrichServerNetworkInterfaces calls ListNetworkInterfaces and filters by server name.
// Sets data.NetworkInterfaces to an empty list (not null) if no VIPs are attached.
// Appends a warning diagnostic (not error) if the API call fails.
func enrichServerNetworkInterfaces(ctx context.Context, c *client.FlashBladeClient, data *serverResourceModel, diags *diag.Diagnostics) {
	nis, err := c.ListNetworkInterfaces(ctx)
	if err != nil {
		diags.AddWarning(
			"Could not list network interfaces",
			fmt.Sprintf("VIP enrichment for server %q failed: %s. network_interfaces will be empty.", data.Name.ValueString(), err.Error()),
		)
		data.NetworkInterfaces = types.ListValueMust(types.StringType, []attr.Value{})
		return
	}

	serverName := data.Name.ValueString()
	var matchingNames []string
	for _, ni := range nis {
		for _, as := range ni.AttachedServers {
			if as.Name == serverName {
				matchingNames = append(matchingNames, ni.Name)
				break
			}
		}
	}

	if matchingNames == nil {
		matchingNames = []string{}
	}

	niList, listDiags := types.ListValueFrom(ctx, types.StringType, matchingNames)
	diags.Append(listDiags...)
	if diags.HasError() {
		return
	}
	data.NetworkInterfaces = niList
}

// dnsNamesToRefs converts the flat string list of DNS names in the model to []client.NamedReference.
func dnsNamesToRefs(ctx context.Context, data *serverResourceModel, diags *diag.Diagnostics) []client.NamedReference {
	if data.DNS.IsNull() || data.DNS.IsUnknown() || len(data.DNS.Elements()) == 0 {
		return nil
	}
	var names []string
	diags.Append(data.DNS.ElementsAs(ctx, &names, false)...)
	if diags.HasError() {
		return nil
	}
	refs := make([]client.NamedReference, len(names))
	for i, n := range names {
		refs[i] = client.NamedReference{Name: n}
	}
	return refs
}

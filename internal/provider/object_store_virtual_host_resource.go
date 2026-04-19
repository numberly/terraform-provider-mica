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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &objectStoreVirtualHostResource{}
var _ resource.ResourceWithConfigure = &objectStoreVirtualHostResource{}
var _ resource.ResourceWithImportState = &objectStoreVirtualHostResource{}
var _ resource.ResourceWithUpgradeState = &objectStoreVirtualHostResource{}

// objectStoreVirtualHostResource implements the flashblade_object_store_virtual_host resource.
type objectStoreVirtualHostResource struct {
	client *client.FlashBladeClient
}

func NewObjectStoreVirtualHostResource() resource.Resource {
	return &objectStoreVirtualHostResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreVirtualHostModel is the top-level model for the flashblade_object_store_virtual_host resource.
type objectStoreVirtualHostModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Hostname        types.String   `tfsdk:"hostname"`
	AttachedServers types.List     `tfsdk:"attached_servers"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// objectStoreVirtualHostV0Model is the v0 state model (name was Computed).
type objectStoreVirtualHostV0Model struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Hostname        types.String   `tfsdk:"hostname"`
	AttachedServers types.List     `tfsdk:"attached_servers"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *objectStoreVirtualHostResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_virtual_host"
}

// Schema defines the resource schema.
func (r *objectStoreVirtualHostResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a FlashBlade object store virtual host for virtual-hosted-style S3 endpoints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the virtual host.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The user-specified name of the virtual host. Must contain only alphanumeric characters, hyphens, and underscores.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					HostnameNoDotValidator(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "The hostname (FQDN) for the virtual-hosted-style S3 endpoint.",
			},
			"attached_servers": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of server names attached to this virtual host. The API may auto-attach the default array server.",
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
func (r *objectStoreVirtualHostResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Computed: true,
					},
					"hostname": schema.StringAttribute{
						Required: true,
					},
					"attached_servers": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
						Create: true, Read: true, Update: true, Delete: true,
					}),
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var old objectStoreVirtualHostV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
				if resp.Diagnostics.HasError() {
					return
				}

				newState := objectStoreVirtualHostModel(old)
				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *objectStoreVirtualHostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *objectStoreVirtualHostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreVirtualHostModel
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

	post := client.ObjectStoreVirtualHostPost{
		Hostname:        data.Hostname.ValueString(),
		AttachedServers: modelServersToNamedRefs(ctx, &data, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	vh, err := r.client.PostObjectStoreVirtualHost(ctx, name, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store virtual host", err.Error())
		return
	}

	mapVirtualHostToModel(ctx, vh, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *objectStoreVirtualHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreVirtualHostModel
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
	vh, err := r.client.GetObjectStoreVirtualHost(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store virtual host", err.Error())
		return
	}

	// Drift detection on name.
	if data.Name.ValueString() != vh.Name {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "name",
			"was":      data.Name.ValueString(),
			"now":      vh.Name,
		})
	}

	// Drift detection on hostname.
	if !data.Hostname.IsNull() && !data.Hostname.IsUnknown() {
		if data.Hostname.ValueString() != vh.Hostname {
			tflog.Debug(ctx, "drift detected", map[string]any{
				"resource": name,
				"field":    "hostname",
				"was":      data.Hostname.ValueString(),
				"now":      vh.Hostname,
			})
		}
	}

	mapVirtualHostToModel(ctx, vh, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing object store virtual host.
func (r *objectStoreVirtualHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state objectStoreVirtualHostModel
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

	name := state.Name.ValueString()
	patch := client.ObjectStoreVirtualHostPatch{}
	needsPatch := false

	if !plan.Hostname.Equal(state.Hostname) {
		v := plan.Hostname.ValueString()
		patch.Hostname = &v
		needsPatch = true
	}

	if !plan.AttachedServers.Equal(state.AttachedServers) {
		// Full-replace semantics: always send the full desired list.
		patch.AttachedServers = modelServersToNamedRefs(ctx, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		needsPatch = true
	}

	if needsPatch {
		_, err := r.client.PatchObjectStoreVirtualHost(ctx, name, patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating object store virtual host", err.Error())
			return
		}
	}

	r.readIntoState(ctx, name, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an object store virtual host.
func (r *objectStoreVirtualHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreVirtualHostModel
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

	name := data.Name.ValueString()
	if err := r.client.DeleteObjectStoreVirtualHost(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting object store virtual host", err.Error())
		return
	}
}

// ImportState imports an existing object store virtual host by name or hostname.
// Tries GET by ?names= first; if not found, lists all and matches by name or hostname.
func (r *objectStoreVirtualHostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Try direct lookup by name first.
	vh, err := r.client.GetObjectStoreVirtualHost(ctx, importID)
	if err != nil {
		// Direct lookup failed — list all and match by name or hostname.
		hosts, listErr := r.client.ListObjectStoreVirtualHosts(ctx, client.ListObjectStoreVirtualHostsOpts{})
		if listErr != nil {
			resp.Diagnostics.AddError("Error listing object store virtual hosts for import", listErr.Error())
			return
		}
		for i := range hosts {
			if hosts[i].Name == importID || hosts[i].Hostname == importID {
				vh = &hosts[i]
				break
			}
		}
		if vh == nil {
			resp.Diagnostics.AddError(
				"Object store virtual host not found",
				fmt.Sprintf("No virtual host with name or hostname %q found on the FlashBlade array.", importID),
			)
			return
		}
	}

	var data objectStoreVirtualHostModel
	data.Timeouts = nullTimeoutsValue()
	mapVirtualHostToModel(ctx, vh, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetObjectStoreVirtualHost and maps the result into the provided model.
func (r *objectStoreVirtualHostResource) readIntoState(ctx context.Context, name string, data *objectStoreVirtualHostModel, diags interface {
	AddError(string, string)
	HasError() bool
	Append(...diag.Diagnostic)
}) {
	vh, err := r.client.GetObjectStoreVirtualHost(ctx, name)
	if err != nil {
		diags.AddError("Error reading object store virtual host after write", err.Error())
		return
	}
	mapVirtualHostToModel(ctx, vh, data, diags)
}

// mapVirtualHostToModel converts a client.ObjectStoreVirtualHost to the Terraform model.
// It preserves user-managed fields (Timeouts).
func mapVirtualHostToModel(ctx context.Context, vh *client.ObjectStoreVirtualHost, data *objectStoreVirtualHostModel, diags interface {
	Append(...diag.Diagnostic)
	HasError() bool
}) {
	data.ID = types.StringValue(vh.ID)
	data.Name = types.StringValue(vh.Name)
	data.Hostname = types.StringValue(vh.Hostname)

	// Convert []NamedReference to types.List of string (server names).
	if len(vh.AttachedServers) > 0 {
		serverNames := make([]string, len(vh.AttachedServers))
		for i, s := range vh.AttachedServers {
			serverNames[i] = s.Name
		}
		serverList, d := types.ListValueFrom(ctx, types.StringType, serverNames)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		data.AttachedServers = serverList
	} else {
		// Empty list, not null — avoids drift.
		data.AttachedServers = types.ListValueMust(types.StringType, []attr.Value{})
	}
}

// modelServersToNamedRefs extracts attached_servers from the Terraform model and converts to []NamedReference.
func modelServersToNamedRefs(ctx context.Context, data *objectStoreVirtualHostModel, diags interface {
	Append(...diag.Diagnostic)
	HasError() bool
}) []client.NamedReference {
	if data.AttachedServers.IsNull() || data.AttachedServers.IsUnknown() || len(data.AttachedServers.Elements()) == 0 {
		return nil
	}

	var serverNames []string
	d := data.AttachedServers.ElementsAs(ctx, &serverNames, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	refs := make([]client.NamedReference, len(serverNames))
	for i, name := range serverNames {
		refs[i] = client.NamedReference{Name: name}
	}
	return refs
}

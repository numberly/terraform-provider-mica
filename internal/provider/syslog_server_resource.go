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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure syslogServerResource satisfies the resource interfaces.
var _ resource.Resource = &syslogServerResource{}
var _ resource.ResourceWithConfigure = &syslogServerResource{}
var _ resource.ResourceWithImportState = &syslogServerResource{}
var _ resource.ResourceWithUpgradeState = &syslogServerResource{}

// syslogServerResource implements the flashblade_syslog_server resource.
type syslogServerResource struct {
	client *client.FlashBladeClient
}

// NewSyslogServerResource is the factory function registered in the provider.
func NewSyslogServerResource() resource.Resource {
	return &syslogServerResource{}
}

// ---------- model structs ----------------------------------------------------

// syslogServerModel is the top-level model for the flashblade_syslog_server resource.
type syslogServerModel struct {
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	URI      types.String   `tfsdk:"uri"`
	Services types.List     `tfsdk:"services"`
	Sources  types.List     `tfsdk:"sources"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *syslogServerResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_syslog_server"
}

// Schema defines the resource schema.
func (r *syslogServerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade syslog server that receives audit and management logs.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the syslog server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the syslog server. Not renameable; changing forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uri": schema.StringAttribute{
				Required:    true,
				Description: "Syslog server URI in format PROTOCOL://HOST:PORT (e.g. udp://syslog.example.com:514).",
			},
			"services": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of services to send to this syslog server. Valid values: data-audit, management.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"sources": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of sources to send to this syslog server.",
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
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

func (r *syslogServerResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *syslogServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new syslog server.
func (r *syslogServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data syslogServerModel
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
	post := client.SyslogServerPost{
		URI:      data.URI.ValueString(),
		Services: stringSliceFromList(data.Services),
		Sources:  stringSliceFromList(data.Sources),
	}

	srv, err := r.client.PostSyslogServer(ctx, name, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating syslog server", err.Error())
		return
	}

	mapSyslogServerToModel(srv, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *syslogServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data syslogServerModel
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
	srv, err := r.client.GetSyslogServer(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading syslog server", err.Error())
		return
	}

	// Drift detection on uri.
	if !data.URI.IsNull() && !data.URI.IsUnknown() {
		if data.URI.ValueString() != srv.URI {
			tflog.Info(ctx, "drift detected on syslog server", map[string]any{
				"resource":    name,
				"field":       "uri",
				"state_value": data.URI.ValueString(),
				"api_value":   srv.URI,
			})
		}
	}

	mapSyslogServerToModel(srv, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing syslog server.
func (r *syslogServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state syslogServerModel
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
	patch := client.SyslogServerPatch{}
	needsPatch := false

	if !plan.URI.Equal(state.URI) {
		v := plan.URI.ValueString()
		patch.URI = &v
		needsPatch = true
	}

	if !plan.Services.Equal(state.Services) {
		services := stringSliceFromList(plan.Services)
		patch.Services = &services
		needsPatch = true
	}

	if !plan.Sources.Equal(state.Sources) {
		sources := stringSliceFromList(plan.Sources)
		patch.Sources = &sources
		needsPatch = true
	}

	if needsPatch {
		_, err := r.client.PatchSyslogServer(ctx, name, patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating syslog server", err.Error())
			return
		}
	}

	r.readIntoState(ctx, name, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a syslog server.
func (r *syslogServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data syslogServerModel
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
	if err := r.client.DeleteSyslogServer(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting syslog server", err.Error())
		return
	}
}

// ImportState imports an existing syslog server by name.
func (r *syslogServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data syslogServerModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}
	data.Name = types.StringValue(name)

	r.readIntoState(ctx, name, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetSyslogServer and maps the result into the provided model.
func (r *syslogServerResource) readIntoState(ctx context.Context, name string, data *syslogServerModel, diags interface {
	AddError(string, string)
	HasError() bool
	Append(...diag.Diagnostic)
}) {
	srv, err := r.client.GetSyslogServer(ctx, name)
	if err != nil {
		diags.AddError("Error reading syslog server after write", err.Error())
		return
	}
	mapSyslogServerToModel(srv, data)
}

// mapSyslogServerToModel converts a client.SyslogServer to the Terraform model.
// It preserves user-managed fields (Timeouts).
func mapSyslogServerToModel(srv *client.SyslogServer, data *syslogServerModel) {
	data.ID = types.StringValue(srv.ID)
	data.Name = types.StringValue(srv.Name)
	data.URI = types.StringValue(srv.URI)

	// Map Services: nil -> empty list (not null) to avoid drift.
	if len(srv.Services) > 0 {
		vals := make([]attr.Value, len(srv.Services))
		for i, s := range srv.Services {
			vals[i] = types.StringValue(s)
		}
		data.Services = types.ListValueMust(types.StringType, vals)
	} else {
		data.Services = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Map Sources: nil -> empty list (not null) to avoid drift.
	if len(srv.Sources) > 0 {
		vals := make([]attr.Value, len(srv.Sources))
		for i, s := range srv.Sources {
			vals[i] = types.StringValue(s)
		}
		data.Sources = types.ListValueMust(types.StringType, vals)
	} else {
		data.Sources = types.ListValueMust(types.StringType, []attr.Value{})
	}
}

// stringSliceFromList extracts a Go string slice from a types.List.
// Returns nil if the list is null or unknown.
func stringSliceFromList(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elems := list.Elements()
	result := make([]string, len(elems))
	for i, e := range elems {
		result[i] = e.(types.String).ValueString()
	}
	return result
}

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure arrayConnectionResource satisfies the resource interfaces.
var _ resource.Resource = &arrayConnectionResource{}
var _ resource.ResourceWithConfigure = &arrayConnectionResource{}
var _ resource.ResourceWithImportState = &arrayConnectionResource{}
var _ resource.ResourceWithUpgradeState = &arrayConnectionResource{}

// arrayConnectionResource implements the flashblade_array_connection resource.
type arrayConnectionResource struct {
	client *client.FlashBladeClient
}

// NewArrayConnectionResource is the factory function registered in the provider.
func NewArrayConnectionResource() resource.Resource {
	return &arrayConnectionResource{}
}

// ---------- model structs ----------------------------------------------------

// arrayConnectionModel is the top-level model for the flashblade_array_connection resource.
type arrayConnectionModel struct {
	ID                   types.String   `tfsdk:"id"`
	RemoteName           types.String   `tfsdk:"remote_name"`
	ManagementAddress    types.String   `tfsdk:"management_address"`
	ConnectionKey        types.String   `tfsdk:"connection_key"`
	Encrypted            types.Bool     `tfsdk:"encrypted"`
	CACertificateGroup   types.String   `tfsdk:"ca_certificate_group"`
	ReplicationAddresses types.List     `tfsdk:"replication_addresses"`
	Throttle             types.Object   `tfsdk:"throttle"`
	Status               types.String   `tfsdk:"status"`
	Type                 types.String   `tfsdk:"type"`
	OS                   types.String   `tfsdk:"os"`
	Version              types.String   `tfsdk:"version"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

// throttleModel maps the throttle nested object.
type throttleModel struct {
	DefaultLimit types.Int64  `tfsdk:"default_limit"`
	WindowLimit  types.Int64  `tfsdk:"window_limit"`
	WindowStart  types.String `tfsdk:"window_start"`
	WindowEnd    types.String `tfsdk:"window_end"`
}

// throttleAttrTypes is the attribute type map for the throttle nested object.
var throttleAttrTypes = map[string]attr.Type{
	"default_limit": types.Int64Type,
	"window_limit":  types.Int64Type,
	"window_start":  types.StringType,
	"window_end":    types.StringType,
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *arrayConnectionResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_array_connection"
}

// Schema defines the resource schema.
func (r *arrayConnectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade array connection to a remote FlashBlade array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the array connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"remote_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the remote array. Used as the import identifier. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"management_address": schema.StringAttribute{
				Required:    true,
				Description: "Management IP or hostname of the remote array.",
			},
			"connection_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Connection key of the remote array. Write-only: set on creation only, not returned by GET.",
			},
			"encrypted": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether data is encrypted in transit.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ca_certificate_group": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the CA certificate group for TLS verification.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replication_addresses": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Replication IP addresses or FQDNs.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"throttle": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Bandwidth throttle configuration for the array connection.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"default_limit": schema.Int64Attribute{
						Optional:    true,
						Description: "Default bandwidth limit in bytes per second.",
					},
					"window_limit": schema.Int64Attribute{
						Optional:    true,
						Description: "Window bandwidth limit in bytes per second.",
					},
					"window_start": schema.StringAttribute{
						Optional:    true,
						Description: "Start time of the throttle window (HH:MM format).",
					},
					"window_end": schema.StringAttribute{
						Optional:    true,
						Description: "End time of the throttle window (HH:MM format).",
					},
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Connection status (connected, connecting, etc.).",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Connection type (async-replication, etc.).",
			},
			"os": schema.StringAttribute{
				Computed:    true,
				Description: "Operating system of the remote array.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the remote array.",
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
func (r *arrayConnectionResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *arrayConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new array connection.
func (r *arrayConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data arrayConnectionModel
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

	post := client.ArrayConnectionPost{
		ManagementAddress: data.ManagementAddress.ValueString(),
		ConnectionKey:     data.ConnectionKey.ValueString(),
		Encrypted:         data.Encrypted.ValueBool(),
	}

	if !data.CACertificateGroup.IsNull() && !data.CACertificateGroup.IsUnknown() && data.CACertificateGroup.ValueString() != "" {
		post.CACertificateGroup = &client.NamedReference{Name: data.CACertificateGroup.ValueString()}
	}

	if !data.ReplicationAddresses.IsNull() && !data.ReplicationAddresses.IsUnknown() {
		addrs, d := listToStrings(ctx, data.ReplicationAddresses)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(addrs) > 0 {
			post.ReplicationAddresses = addrs
		}
	}

	if !data.Throttle.IsNull() && !data.Throttle.IsUnknown() {
		throttle, d := throttleFromObject(ctx, data.Throttle)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		post.Throttle = throttle
	}

	// Preserve connection_key from plan — API never returns it.
	planConnKey := data.ConnectionKey

	conn, err := r.client.PostArrayConnection(ctx, data.RemoteName.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating array connection", err.Error())
		return
	}

	resp.Diagnostics.Append(mapArrayConnectionToModel(ctx, conn, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Restore connection_key from plan (not returned by API).
	data.ConnectionKey = planConnKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API, logging field-level drift.
func (r *arrayConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data arrayConnectionModel
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

	remoteName := data.RemoteName.ValueString()
	conn, err := r.client.GetArrayConnection(ctx, remoteName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading array connection", err.Error())
		return
	}

	// Drift detection: compare old state vs API response, log each changed field.
	if data.ManagementAddress.ValueString() != conn.ManagementAddress {
		tflog.Debug(ctx, "array_connection field changed outside Terraform", map[string]any{
			"resource": remoteName,
			"field":    "management_address",
			"was":      data.ManagementAddress.ValueString(),
			"now":      conn.ManagementAddress,
		})
	}

	if data.Encrypted.ValueBool() != conn.Encrypted {
		tflog.Debug(ctx, "array_connection field changed outside Terraform", map[string]any{
			"resource": remoteName,
			"field":    "encrypted",
			"was":      data.Encrypted.ValueBool(),
			"now":      conn.Encrypted,
		})
	}

	newCACertGroup := ""
	if conn.CACertificateGroup != nil {
		newCACertGroup = conn.CACertificateGroup.Name
	}
	if data.CACertificateGroup.ValueString() != newCACertGroup {
		tflog.Debug(ctx, "array_connection field changed outside Terraform", map[string]any{
			"resource": remoteName,
			"field":    "ca_certificate_group",
			"was":      data.CACertificateGroup.ValueString(),
			"now":      newCACertGroup,
		})
	}

	if data.Status.ValueString() != conn.Status {
		tflog.Debug(ctx, "array_connection field changed outside Terraform", map[string]any{
			"resource": remoteName,
			"field":    "status",
			"was":      data.Status.ValueString(),
			"now":      conn.Status,
		})
	}

	if data.Type.ValueString() != conn.Type {
		tflog.Debug(ctx, "array_connection field changed outside Terraform", map[string]any{
			"resource": remoteName,
			"field":    "type",
			"was":      data.Type.ValueString(),
			"now":      conn.Type,
		})
	}

	if data.OS.ValueString() != conn.OS {
		tflog.Debug(ctx, "array_connection field changed outside Terraform", map[string]any{
			"resource": remoteName,
			"field":    "os",
			"was":      data.OS.ValueString(),
			"now":      conn.OS,
		})
	}

	if data.Version.ValueString() != conn.Version {
		tflog.Debug(ctx, "array_connection field changed outside Terraform", map[string]any{
			"resource": remoteName,
			"field":    "version",
			"was":      data.Version.ValueString(),
			"now":      conn.Version,
		})
	}

	// Preserve connection_key from state — API never returns it.
	existingConnKey := data.ConnectionKey

	resp.Diagnostics.Append(mapArrayConnectionToModel(ctx, conn, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Restore connection_key from state (not returned by API).
	data.ConnectionKey = existingConnKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing array connection.
func (r *arrayConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state arrayConnectionModel
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

	patch := client.ArrayConnectionPatch{}
	hasChanges := false

	if !plan.ManagementAddress.Equal(state.ManagementAddress) {
		v := plan.ManagementAddress.ValueString()
		patch.ManagementAddress = &v
		hasChanges = true
	}

	if !plan.Encrypted.Equal(state.Encrypted) {
		v := plan.Encrypted.ValueBool()
		patch.Encrypted = &v
		hasChanges = true
	}

	if !plan.CACertificateGroup.Equal(state.CACertificateGroup) {
		planVal := plan.CACertificateGroup.ValueString()
		if planVal == "" || plan.CACertificateGroup.IsNull() {
			// Clear: outer non-nil, inner nil.
			var inner *client.NamedReference
			patch.CACertificateGroup = &inner
		} else {
			ref := &client.NamedReference{Name: planVal}
			patch.CACertificateGroup = &ref
		}
		hasChanges = true
	}

	if !plan.ReplicationAddresses.Equal(state.ReplicationAddresses) {
		addrs, d := listToStrings(ctx, plan.ReplicationAddresses)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.ReplicationAddresses = &addrs
		hasChanges = true
	}

	if !plan.Throttle.Equal(state.Throttle) {
		if !plan.Throttle.IsNull() && !plan.Throttle.IsUnknown() {
			throttle, d := throttleFromObject(ctx, plan.Throttle)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			patch.Throttle = throttle
		}
		hasChanges = true
	}

	// connection_key cannot be PATCHed — log warning if it changed.
	if !plan.ConnectionKey.Equal(state.ConnectionKey) {
		tflog.Warn(ctx, "connection_key change detected but cannot be applied via PATCH (API limitation); key change is ignored", map[string]any{
			"resource": state.RemoteName.ValueString(),
		})
	}

	remoteName := state.RemoteName.ValueString()

	if hasChanges {
		_, err := r.client.PatchArrayConnection(ctx, remoteName, patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating array connection", err.Error())
			return
		}
	}

	// Re-read to refresh computed fields.
	conn, err := r.client.GetArrayConnection(ctx, remoteName)
	if err != nil {
		resp.Diagnostics.AddError("Error reading array connection after update", err.Error())
		return
	}

	// Preserve connection_key from state.
	stateConnKey := state.ConnectionKey

	resp.Diagnostics.Append(mapArrayConnectionToModel(ctx, conn, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Restore connection_key from state (not returned by API).
	plan.ConnectionKey = stateConnKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an array connection.
func (r *arrayConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data arrayConnectionModel
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

	err := r.client.DeleteArrayConnection(ctx, data.RemoteName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting array connection", err.Error())
		return
	}
}

// ImportState imports an existing array connection by remote name.
func (r *arrayConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	remoteName := req.ID
	conn, err := r.client.GetArrayConnection(ctx, remoteName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing array connection", err.Error())
		return
	}

	var data arrayConnectionModel
	data.Timeouts = nullTimeoutsValue()

	resp.Diagnostics.Append(mapArrayConnectionToModel(ctx, conn, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// connection_key is write-only — not returned by GET; set to empty string on import.
	data.ConnectionKey = types.StringValue("")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapArrayConnectionToModel maps a client.ArrayConnection to an arrayConnectionModel.
// It does NOT touch ConnectionKey — the caller manages it.
func mapArrayConnectionToModel(ctx context.Context, conn *client.ArrayConnection, data *arrayConnectionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(conn.ID)
	data.RemoteName = types.StringValue(conn.Remote.Name)
	data.ManagementAddress = types.StringValue(conn.ManagementAddress)
	data.Encrypted = types.BoolValue(conn.Encrypted)
	data.Status = types.StringValue(conn.Status)
	data.Type = types.StringValue(conn.Type)
	data.OS = types.StringValue(conn.OS)
	data.Version = types.StringValue(conn.Version)

	if conn.CACertificateGroup != nil {
		data.CACertificateGroup = types.StringValue(conn.CACertificateGroup.Name)
	} else {
		data.CACertificateGroup = types.StringNull()
	}

	// replication_addresses: map to empty list (not null) when API returns empty, to avoid perpetual drift.
	if len(conn.ReplicationAddresses) == 0 {
		data.ReplicationAddresses = types.ListValueMust(types.StringType, []attr.Value{})
	} else {
		addrs, d := types.ListValueFrom(ctx, types.StringType, conn.ReplicationAddresses)
		diags.Append(d...)
		if !diags.HasError() {
			data.ReplicationAddresses = addrs
		}
	}

	// throttle: map to object or null.
	if conn.Throttle != nil {
		tm := throttleModel{}
		if conn.Throttle.DefaultLimit != nil {
			tm.DefaultLimit = types.Int64Value(*conn.Throttle.DefaultLimit)
		} else {
			tm.DefaultLimit = types.Int64Null()
		}
		if conn.Throttle.WindowLimit != nil {
			tm.WindowLimit = types.Int64Value(*conn.Throttle.WindowLimit)
		} else {
			tm.WindowLimit = types.Int64Null()
		}
		if conn.Throttle.WindowStart != nil {
			tm.WindowStart = types.StringValue(*conn.Throttle.WindowStart)
		} else {
			tm.WindowStart = types.StringNull()
		}
		if conn.Throttle.WindowEnd != nil {
			tm.WindowEnd = types.StringValue(*conn.Throttle.WindowEnd)
		} else {
			tm.WindowEnd = types.StringNull()
		}
		obj, d := types.ObjectValueFrom(ctx, throttleAttrTypes, tm)
		diags.Append(d...)
		if !diags.HasError() {
			data.Throttle = obj
		}
	} else {
		data.Throttle = types.ObjectNull(throttleAttrTypes)
	}

	return diags
}

// throttleFromObject extracts a client.ArrayConnectionThrottle from a types.Object.
func throttleFromObject(ctx context.Context, obj types.Object) (*client.ArrayConnectionThrottle, diag.Diagnostics) {
	var tm throttleModel
	diags := obj.As(ctx, &tm, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	throttle := &client.ArrayConnectionThrottle{}
	if !tm.DefaultLimit.IsNull() && !tm.DefaultLimit.IsUnknown() {
		v := tm.DefaultLimit.ValueInt64()
		throttle.DefaultLimit = &v
	}
	if !tm.WindowLimit.IsNull() && !tm.WindowLimit.IsUnknown() {
		v := tm.WindowLimit.ValueInt64()
		throttle.WindowLimit = &v
	}
	if !tm.WindowStart.IsNull() && !tm.WindowStart.IsUnknown() {
		v := tm.WindowStart.ValueString()
		throttle.WindowStart = &v
	}
	if !tm.WindowEnd.IsNull() && !tm.WindowEnd.IsUnknown() {
		v := tm.WindowEnd.ValueString()
		throttle.WindowEnd = &v
	}
	return throttle, diags
}

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure targetResource satisfies the resource interfaces.
var _ resource.Resource = &targetResource{}
var _ resource.ResourceWithConfigure = &targetResource{}
var _ resource.ResourceWithImportState = &targetResource{}
var _ resource.ResourceWithUpgradeState = &targetResource{}

// targetResource implements the flashblade_target resource.
type targetResource struct {
	client *client.FlashBladeClient
}

// NewTargetResource is the factory function registered in the provider.
func NewTargetResource() resource.Resource {
	return &targetResource{}
}

// ---------- model structs ----------------------------------------------------

// targetModel is the top-level model for the flashblade_target resource.
type targetModel struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Address            types.String   `tfsdk:"address"`
	CACertificateGroup types.String   `tfsdk:"ca_certificate_group"`
	Status             types.String   `tfsdk:"status"`
	StatusDetails      types.String   `tfsdk:"status_details"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *targetResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_target"
}

// Schema defines the resource schema.
func (r *targetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade replication target (external S3 endpoint).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the target. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "The hostname or IP address of the target S3 endpoint.",
			},
			"ca_certificate_group": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the CA certificate group used to validate the target's TLS certificate. Null when not set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The connection status of the target (e.g. connected, connecting, error).",
			},
			"status_details": schema.StringAttribute{
				Computed:    true,
				Description: "Additional details about the connection status.",
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
func (r *targetResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *targetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new replication target.
func (r *targetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data targetModel
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

	post := client.TargetPost{
		Address: data.Address.ValueString(),
	}

	tgt, err := r.client.PostTarget(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating target", err.Error())
		return
	}

	mapTargetToModel(tgt, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API, logging field-level drift.
func (r *targetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data targetModel
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
	tgt, err := r.client.GetTarget(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading target", err.Error())
		return
	}

	// Drift detection: compare old state vs API response and log each changed field.
	newAddress := tgt.Address
	if data.Address.ValueString() != newAddress {
		tflog.Debug(ctx, "target field changed outside Terraform", map[string]any{
			"field": "address",
			"was":   data.Address.ValueString(),
			"now":   newAddress,
		})
	}

	newCACertGroup := ""
	if tgt.CACertificateGroup != nil {
		newCACertGroup = tgt.CACertificateGroup.Name
	}
	oldCACertGroup := data.CACertificateGroup.ValueString()
	if oldCACertGroup != newCACertGroup {
		tflog.Debug(ctx, "target field changed outside Terraform", map[string]any{
			"field": "ca_certificate_group",
			"was":   oldCACertGroup,
			"now":   newCACertGroup,
		})
	}

	if data.Status.ValueString() != tgt.Status {
		tflog.Debug(ctx, "target field changed outside Terraform", map[string]any{
			"field": "status",
			"was":   data.Status.ValueString(),
			"now":   tgt.Status,
		})
	}

	if data.StatusDetails.ValueString() != tgt.StatusDetails {
		tflog.Debug(ctx, "target field changed outside Terraform", map[string]any{
			"field": "status_details",
			"was":   data.StatusDetails.ValueString(),
			"now":   tgt.StatusDetails,
		})
	}

	mapTargetToModel(tgt, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing target.
func (r *targetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state targetModel
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

	patch := client.TargetPatch{}

	if !plan.Address.Equal(state.Address) {
		v := plan.Address.ValueString()
		patch.Address = &v
	}

	if !plan.CACertificateGroup.Equal(state.CACertificateGroup) && !plan.CACertificateGroup.IsUnknown() {
		if plan.CACertificateGroup.IsNull() || plan.CACertificateGroup.ValueString() == "" {
			// Clear the cert group: outer ptr non-nil, inner ptr nil.
			var inner *client.NamedReference
			patch.CACertificateGroup = &inner
		} else {
			// Set to a specific group.
			nr := &client.NamedReference{Name: plan.CACertificateGroup.ValueString()}
			patch.CACertificateGroup = &nr
		}
	}

	_, err := r.client.PatchTarget(ctx, state.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating target", err.Error())
		return
	}

	// Re-read to refresh computed fields.
	tgt, err := r.client.GetTarget(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading target after update", err.Error())
		return
	}

	mapTargetToModel(tgt, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a replication target.
func (r *targetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data targetModel
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

	err := r.client.DeleteTarget(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting target", err.Error())
		return
	}
}

// ImportState imports an existing target by name.
func (r *targetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	tgt, err := r.client.GetTarget(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing target", err.Error())
		return
	}

	var data targetModel
	data.Timeouts = nullTimeoutsValue()

	mapTargetToModel(tgt, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapTargetToModel maps a client.Target to a targetModel.
func mapTargetToModel(tgt *client.Target, data *targetModel) {
	data.ID = types.StringValue(tgt.ID)
	data.Name = types.StringValue(tgt.Name)
	data.Address = types.StringValue(tgt.Address)
	data.Status = types.StringValue(tgt.Status)
	data.StatusDetails = types.StringValue(tgt.StatusDetails)

	if tgt.CACertificateGroup != nil {
		data.CACertificateGroup = types.StringValue(tgt.CACertificateGroup.Name)
	} else {
		data.CACertificateGroup = types.StringNull()
	}
}

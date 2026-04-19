package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure certificateGroupResource satisfies the resource interfaces.
var _ resource.Resource = &certificateGroupResource{}
var _ resource.ResourceWithConfigure = &certificateGroupResource{}
var _ resource.ResourceWithImportState = &certificateGroupResource{}
var _ resource.ResourceWithUpgradeState = &certificateGroupResource{}

// certificateGroupResource implements the flashblade_certificate_group resource.
type certificateGroupResource struct {
	client *client.FlashBladeClient
}

// NewCertificateGroupResource is the factory function registered in the provider.
func NewCertificateGroupResource() resource.Resource {
	return &certificateGroupResource{}
}

// ---------- model structs ----------------------------------------------------

// certificateGroupModel is the Terraform state model for the flashblade_certificate_group resource.
type certificateGroupModel struct {
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	Realms   types.List     `tfsdk:"realms"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *certificateGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_certificate_group"
}

// Schema defines the resource schema.
func (r *certificateGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade certificate group that bundles certificates for use in TLS policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the certificate group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the certificate group. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"realms": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The list of realms associated with this certificate group. Set by the array.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

// UpgradeState returns state upgraders for schema migrations.
func (r *certificateGroupResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *certificateGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ---------- CRD methods (no Update) -----------------------------------------

// Create creates a new certificate group.
func (r *certificateGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data certificateGroupModel
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

	group, err := r.client.PostCertificateGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate group", err.Error())
		return
	}

	mapCertificateGroupToModel(group, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API, logging field-level drift.
func (r *certificateGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data certificateGroupModel
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
	group, err := r.client.GetCertificateGroup(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate group", err.Error())
		return
	}

	// Drift detection on all mutable/computed fields.
	if data.Name.ValueString() != group.Name {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "name",
			"was":      data.Name.ValueString(),
			"now":      group.Name,
		})
	}

	wasRealmsList, dRealms := listToStrings(ctx, data.Realms)
	resp.Diagnostics.Append(dRealms...)
	wasRealms := strings.Join(wasRealmsList, ",")
	nowRealms := strings.Join(group.Realms, ",")
	if wasRealms != nowRealms {
		tflog.Debug(ctx, "drift detected", map[string]any{
			"resource": name,
			"field":    "realms",
			"was":      wasRealms,
			"now":      nowRealms,
		})
	}

	mapCertificateGroupToModel(group, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported for certificate groups (CRD-only resource).
func (r *certificateGroupResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Certificate groups cannot be updated — all attributes force a new resource.",
	)
}

// Delete removes a certificate group.
func (r *certificateGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data certificateGroupModel
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

	err := r.client.DeleteCertificateGroup(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting certificate group", err.Error())
		return
	}
}

// ImportState imports an existing certificate group by name.
func (r *certificateGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID
	group, err := r.client.GetCertificateGroup(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing certificate group", err.Error())
		return
	}

	var data certificateGroupModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	mapCertificateGroupToModel(group, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapCertificateGroupToModel maps a client.CertificateGroup to a certificateGroupModel.
func mapCertificateGroupToModel(group *client.CertificateGroup, data *certificateGroupModel) {
	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)

	if len(group.Realms) == 0 {
		data.Realms = types.ListValueMust(types.StringType, []attr.Value{})
	} else {
		elems := make([]attr.Value, len(group.Realms))
		for i, realm := range group.Realms {
			elems[i] = types.StringValue(realm)
		}
		data.Realms = types.ListValueMust(types.StringType, elems)
	}
}

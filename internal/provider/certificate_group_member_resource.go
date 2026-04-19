package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &certificateGroupMemberResource{}
var _ resource.ResourceWithConfigure = &certificateGroupMemberResource{}
var _ resource.ResourceWithImportState = &certificateGroupMemberResource{}
var _ resource.ResourceWithUpgradeState = &certificateGroupMemberResource{}

// certificateGroupMemberResource implements the flashblade_certificate_group_member resource.
type certificateGroupMemberResource struct {
	client *client.FlashBladeClient
}

func NewCertificateGroupMemberResource() resource.Resource {
	return &certificateGroupMemberResource{}
}

// ---------- model structs ----------------------------------------------------

// certificateGroupMemberModel is the Terraform state model for the flashblade_certificate_group_member resource.
type certificateGroupMemberModel struct {
	GroupName types.String   `tfsdk:"group_name"`
	CertName  types.String   `tfsdk:"certificate_name"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *certificateGroupMemberResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_certificate_group_member"
}

// Schema defines the resource schema.
func (r *certificateGroupMemberResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Assigns a certificate to a FlashBlade certificate group. This is a CRD-only resource — all fields force a new resource on change.",
		Attributes: map[string]schema.Attribute{
			"group_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the certificate group. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the certificate to add to the group. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
func (r *certificateGroupMemberResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *certificateGroupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ---------- CRD methods (no Update) ------------------------------------------

// Create adds a certificate to a certificate group.
func (r *certificateGroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data certificateGroupMemberModel
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

	_, err := r.client.PostCertificateGroupMember(ctx, data.GroupName.ValueString(), data.CertName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate group member", err.Error())
		return
	}

	// Preserve plan values — API response may not include names in NamedReference fields.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API by listing members and finding the match.
func (r *certificateGroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data certificateGroupMemberModel
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

	_, err := r.client.GetCertificateGroupMember(ctx, data.GroupName.ValueString(), data.CertName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate group member", err.Error())
		return
	}

	// Preserve state values — API membership response may not include names in NamedReference fields.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not supported for certificate group members (CRD-only resource).
func (r *certificateGroupMemberResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Certificate group members cannot be updated — changing group_name or certificate_name forces a new resource.",
	)
}

// Delete removes a certificate from a certificate group.
func (r *certificateGroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data certificateGroupMemberModel
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

	err := r.client.DeleteCertificateGroupMember(ctx, data.GroupName.ValueString(), data.CertName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting certificate group member", err.Error())
		return
	}
}

// ImportState imports an existing certificate group member by composite ID "groupName/certificateName".
func (r *certificateGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: groupName/certificateName. Error: %s", err))
		return
	}

	groupName := parts[0]
	certName := parts[1]

	if _, err := r.client.GetCertificateGroupMember(ctx, groupName, certName); err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Certificate group member not found",
				fmt.Sprintf("No certificate %q found in group %q.", certName, groupName),
			)
			return
		}
		resp.Diagnostics.AddError("Error importing certificate group member", err.Error())
		return
	}

	var data certificateGroupMemberModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	data.GroupName = types.StringValue(groupName)
	data.CertName = types.StringValue(certName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

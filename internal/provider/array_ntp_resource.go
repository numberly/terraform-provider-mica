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

var _ resource.Resource = &arrayNtpResource{}
var _ resource.ResourceWithConfigure = &arrayNtpResource{}
var _ resource.ResourceWithImportState = &arrayNtpResource{}
var _ resource.ResourceWithUpgradeState = &arrayNtpResource{}

// arrayNtpResource implements the flashblade_array_ntp singleton resource.
type arrayNtpResource struct {
	client *client.FlashBladeClient
}

func NewArrayNtpResource() resource.Resource {
	return &arrayNtpResource{}
}

// ---------- model structs ----------------------------------------------------

// arrayNtpModel is the top-level model for the flashblade_array_ntp resource.
type arrayNtpModel struct {
	ID         types.String   `tfsdk:"id"`
	NtpServers types.List     `tfsdk:"ntp_servers"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *arrayNtpResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_array_ntp"
}

// Schema defines the resource schema.
func (r *arrayNtpResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages the NTP server list of a FlashBlade array. This is a singleton resource that wraps the ntp_servers field of the /arrays endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the array.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ntp_servers": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of NTP server hostnames or IP addresses.",
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
func (r *arrayNtpResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *arrayNtpResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create patches the NTP servers on the array (singleton pattern).
func (r *arrayNtpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data arrayNtpModel
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

	servers, d := listToStrings(ctx, data.NtpServers)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.PatchArrayNtp(ctx, client.ArrayNtpPatch{NtpServers: &servers})
	if err != nil {
		resp.Diagnostics.AddError("Error setting NTP servers", err.Error())
		return
	}

	resp.Diagnostics.Append(mapArrayNtpToModel(ctx, result, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *arrayNtpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data arrayNtpModel
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

	arrayInfo, err := r.client.GetArrayNtp(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading NTP servers", err.Error())
		return
	}

	resp.Diagnostics.Append(mapArrayNtpToModel(ctx, arrayInfo, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *arrayNtpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan arrayNtpModel
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

	servers, d := listToStrings(ctx, plan.NtpServers)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.PatchArrayNtp(ctx, client.ArrayNtpPatch{NtpServers: &servers})
	if err != nil {
		resp.Diagnostics.AddError("Error updating NTP servers", err.Error())
		return
	}

	resp.Diagnostics.Append(mapArrayNtpToModel(ctx, result, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete clears the NTP server list (singleton — PATCH to empty).
func (r *arrayNtpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data arrayNtpModel
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

	empty := []string{}
	_, err := r.client.PatchArrayNtp(ctx, client.ArrayNtpPatch{NtpServers: &empty})
	if err != nil {
		resp.Diagnostics.AddError("Error clearing NTP servers", err.Error())
		return
	}

	tflog.Info(ctx, "NTP servers cleared")
}

// ImportState imports the singleton NTP config using "default" as the import ID.
func (r *arrayNtpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data arrayNtpModel
	data.Timeouts = nullTimeoutsValue()

	arrayInfo, err := r.client.GetArrayNtp(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing NTP configuration", err.Error())
		return
	}

	resp.Diagnostics.Append(mapArrayNtpToModel(ctx, arrayInfo, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapArrayNtpToModel maps a client.ArrayInfo to an arrayNtpModel.
func mapArrayNtpToModel(ctx context.Context, info *client.ArrayInfo, data *arrayNtpModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(info.ID)

	if len(info.NtpServers) > 0 {
		servers, d := types.ListValueFrom(ctx, types.StringType, info.NtpServers)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.NtpServers = servers
	} else {
		data.NtpServers = emptyStringList()
	}

	return diags
}

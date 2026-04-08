package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure arrayConnectionKeyResource satisfies the resource interfaces.
// No ImportState (key is ephemeral, no stable import identifier).
// No UpgradeState (schema version 0, no migrations).
var _ resource.Resource = &arrayConnectionKeyResource{}
var _ resource.ResourceWithConfigure = &arrayConnectionKeyResource{}

// arrayConnectionKeyResource implements the flashblade_array_connection_key resource.
type arrayConnectionKeyResource struct {
	client *client.FlashBladeClient
}

// NewArrayConnectionKeyResource is the factory function registered in the provider.
func NewArrayConnectionKeyResource() resource.Resource {
	return &arrayConnectionKeyResource{}
}

// ---------- model structs ----------------------------------------------------

// arrayConnectionKeyModel is the Terraform model for the flashblade_array_connection_key resource.
type arrayConnectionKeyModel struct {
	ID            types.String   `tfsdk:"id"`
	ConnectionKey types.String   `tfsdk:"connection_key"`
	Created       types.Int64    `tfsdk:"created"`
	Expires       types.Int64    `tfsdk:"expires"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *arrayConnectionKeyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_array_connection_key"
}

// Schema defines the resource schema.
func (r *arrayConnectionKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Generates a FlashBlade array connection key via POST. Each apply regenerates the key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic identifier set to the generated connection key value.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The generated connection key. Used by the remote array to establish a connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (ms) when the key was created.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expires": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (ms) when the key expires.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
			}),
		},
	}
}

// UpgradeState returns state upgraders (empty — schema version 0).
func (r *arrayConnectionKeyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *arrayConnectionKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create generates a new connection key via POST.
func (r *arrayConnectionKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data arrayConnectionKeyModel
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

	key, err := r.client.PostArrayConnectionKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error generating array connection key", err.Error())
		return
	}

	data.ID = types.StringValue(key.ConnectionKey)
	data.ConnectionKey = types.StringValue(key.ConnectionKey)
	data.Created = types.Int64Value(key.Created)
	data.Expires = types.Int64Value(key.Expires)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read preserves state as-is. Connection keys are ephemeral — once consumed by the
// remote array or expired, the API no longer returns them. Calling GET would remove
// the resource from state on every refresh, causing perpetual recreations.
// The key value in state remains valid for reference (e.g., by flashblade_array_connection).
func (r *arrayConnectionKeyResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// No-op: preserve existing state. Key is write-once, read from state only.
}

// Update is a stub — all attributes are Computed so Update is never called in practice.
func (r *arrayConnectionKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes are computed. Use -replace to regenerate the key.",
	)
}

// Delete is a no-op. Keys expire automatically — no API call is needed.
func (r *arrayConnectionKeyResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Key expires automatically. No API call needed.
}

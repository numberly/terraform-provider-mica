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
	"github.com/hashicorp/terraform-plugin-log/tflog"

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

// Read refreshes Terraform state from the API, logging field-level drift.
// If the API returns an empty key (expired or array reset), the resource is removed from state.
func (r *arrayConnectionKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data arrayConnectionKeyModel
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

	key, err := r.client.GetArrayConnectionKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading array connection key", err.Error())
		return
	}

	// If the key has expired or been reset, remove from state to force re-creation.
	if key.ConnectionKey == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Drift detection: log field-level changes.
	if data.ConnectionKey.ValueString() != key.ConnectionKey {
		tflog.Debug(ctx, "array_connection_key field changed outside Terraform", map[string]any{
			"field": "connection_key",
			"was":   data.ConnectionKey.ValueString(),
			"now":   key.ConnectionKey,
		})
	}
	if data.Created.ValueInt64() != key.Created {
		tflog.Debug(ctx, "array_connection_key field changed outside Terraform", map[string]any{
			"field": "created",
			"was":   data.Created.ValueInt64(),
			"now":   key.Created,
		})
	}
	if data.Expires.ValueInt64() != key.Expires {
		tflog.Debug(ctx, "array_connection_key field changed outside Terraform", map[string]any{
			"field": "expires",
			"was":   data.Expires.ValueInt64(),
			"now":   key.Expires,
		})
	}

	data.ID = types.StringValue(key.ConnectionKey)
	data.ConnectionKey = types.StringValue(key.ConnectionKey)
	data.Created = types.Int64Value(key.Created)
	data.Expires = types.Int64Value(key.Expires)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

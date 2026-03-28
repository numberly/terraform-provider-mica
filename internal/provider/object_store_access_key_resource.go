package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure objectStoreAccessKeyResource satisfies the resource interfaces.
// Intentionally does NOT implement ResourceWithImportState — secret is unavailable after creation.
var _ resource.Resource = &objectStoreAccessKeyResource{}
var _ resource.ResourceWithConfigure = &objectStoreAccessKeyResource{}

// objectStoreAccessKeyResource implements the flashblade_object_store_access_key resource.
type objectStoreAccessKeyResource struct {
	client *client.FlashBladeClient
}

// NewAccessKeyResource is the factory function registered in the provider.
func NewAccessKeyResource() resource.Resource {
	return &objectStoreAccessKeyResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccessKeyModel is the top-level model for the flashblade_object_store_access_key resource.
type objectStoreAccessKeyModel struct {
	Name               types.String   `tfsdk:"name"`
	ObjectStoreAccount types.String   `tfsdk:"object_store_account"`
	AccessKeyID        types.String   `tfsdk:"access_key_id"`
	SecretAccessKey    types.String   `tfsdk:"secret_access_key"`
	Created            types.Int64    `tfsdk:"created"`
	Enabled            types.Bool     `tfsdk:"enabled"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *objectStoreAccessKeyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_access_key"
}

// Schema defines the resource schema.
func (r *objectStoreAccessKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FlashBlade object store access key. Access keys are immutable — any attribute change forces replacement. The secret_access_key is returned only at creation and preserved in state; it cannot be imported.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The access key name (format: <account>/admin/<key-id>). Assigned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_store_account": schema.StringAttribute{
				Required:    true,
				Description: "The object store account this access key belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "The access key ID (public part of the credential pair).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_access_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret access key. Returned only at creation and preserved in state. Never returned by subsequent API reads.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the access key was created.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the access key is enabled. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *objectStoreAccessKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new object store access key.
// The secret_access_key is returned only here and stored in state immediately.
func (r *objectStoreAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreAccessKeyModel
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

	// Build user name in format "<account>/admin".
	userName := data.ObjectStoreAccount.ValueString() + "/admin"

	key, err := r.client.PostObjectStoreAccessKey(ctx, client.ObjectStoreAccessKeyPost{
		User: client.NamedReference{Name: userName},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store access key", err.Error())
		return
	}

	// Map all response fields — secret_access_key is only available here.
	data.Name = types.StringValue(key.Name)
	data.AccessKeyID = types.StringValue(key.AccessKeyID)
	data.SecretAccessKey = types.StringValue(key.SecretAccessKey)
	data.Created = types.Int64Value(key.Created)
	data.Enabled = types.BoolValue(key.Enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
// CRITICAL: Does NOT overwrite SecretAccessKey — it is not returned by GET.
// UseStateForUnknown preserves the secret from state.
func (r *objectStoreAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreAccessKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	key, err := r.client.GetObjectStoreAccessKey(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store access key", err.Error())
		return
	}

	// Map response fields. SecretAccessKey is intentionally NOT updated here —
	// the API never returns it on GET, and state already holds the value from Create.
	data.Name = types.StringValue(key.Name)
	data.AccessKeyID = types.StringValue(key.AccessKeyID)
	data.Created = types.Int64Value(key.Created)
	data.Enabled = types.BoolValue(key.Enabled)
	// data.SecretAccessKey is left as-is — preserved from prior state.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is not implemented — all attributes are ForceNew (RequiresReplace).
// Terraform will always destroy and recreate rather than calling Update.
func (r *objectStoreAccessKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"flashblade_object_store_access_key does not support in-place updates. All attribute changes force replacement.",
	)
}

// Delete removes an object store access key.
// Not-found is handled gracefully (already deleted).
func (r *objectStoreAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreAccessKeyModel
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

	if err := r.client.DeleteObjectStoreAccessKey(ctx, data.Name.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting object store access key", err.Error())
		return
	}
}

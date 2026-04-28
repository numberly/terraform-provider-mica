package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

// ImportState is implemented as a hard-reject shim: the secret_access_key is
// only returned at creation time and cannot be retrieved by the API afterwards,
// so reconstructing state via import is intentionally unsupported. Convention
// compliance is preserved (4th interface declared) while the runtime contract
// remains explicit. See CONVENTIONS.md §Resource Implementation.
var _ resource.Resource = &objectStoreAccessKeyResource{}
var _ resource.ResourceWithConfigure = &objectStoreAccessKeyResource{}
var _ resource.ResourceWithImportState = &objectStoreAccessKeyResource{}
var _ resource.ResourceWithUpgradeState = &objectStoreAccessKeyResource{}

// objectStoreAccessKeyResource implements the flashblade_object_store_access_key resource.
type objectStoreAccessKeyResource struct {
	client *client.FlashBladeClient
}

func NewObjectStoreAccessKeyResource() resource.Resource {
	return &objectStoreAccessKeyResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccessKeyModel is the top-level model for the flashblade_object_store_access_key resource.
type objectStoreAccessKeyModel struct {
	Name               types.String   `tfsdk:"name"`
	ObjectStoreAccount types.String   `tfsdk:"object_store_account"`
	User               types.String   `tfsdk:"user"`
	AccessKeyID        types.String   `tfsdk:"access_key_id"`
	SecretAccessKey    types.String   `tfsdk:"secret_access_key"`
	Created            types.Int64    `tfsdk:"created"`
	Enabled            types.Bool     `tfsdk:"enabled"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *objectStoreAccessKeyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_access_key"
}

// Schema defines the resource schema.
func (r *objectStoreAccessKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade object store access key. Access keys are immutable — any attribute change forces replacement. The secret_access_key can be optionally provided for cross-array replication (sharing the same credentials across arrays). When omitted, the API generates a random secret. The secret is stored in state (encrypted) and marked sensitive.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The access key name (format: <account>/admin/<key-id>). When providing a secret_access_key for cross-array replication, this must be set to the same name as the source key. When omitted, the API assigns it automatically.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_store_account": schema.StringAttribute{
				Required:    true,
				Description: "The object store account this access key belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The S3 user this access key belongs to (format: account/username). When omitted, defaults to account/admin. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "The secret access key. When provided, the key is created with this exact secret (for cross-array replication). When omitted, the API generates it. Returned only at creation time and stored in state (encrypted).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the access key was created.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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


// UpgradeState returns state upgraders for schema migrations.
func (r *objectStoreAccessKeyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// ImportState is intentionally rejected — see file header comment.
func (r *objectStoreAccessKeyResource) ImportState(_ context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Import not supported",
		"flashblade_object_store_access_key cannot be imported because secret_access_key is only returned at creation time and is never retrievable afterwards. Recreate the resource via `terraform apply` instead.",
	)
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

// The secret_access_key is returned only here — it is a write-only attribute and will
// not be persisted in Terraform state. Operators must capture it via a Terraform output.
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

	// Resolve user name: explicit user attribute takes precedence, default to account/admin.
	var userName string
	if !data.User.IsNull() && !data.User.IsUnknown() {
		userName = data.User.ValueString()
	} else {
		userName = data.ObjectStoreAccount.ValueString() + "/admin"
	}

	// Ensure the object store user exists before creating the access key.
	// FlashBlade requires the user to exist — it is not auto-created.
	if err := r.client.EnsureObjectStoreUser(ctx, userName); err != nil {
		resp.Diagnostics.AddError("Error ensuring object store user", err.Error())
		return
	}

	post := client.ObjectStoreAccessKeyPost{
		User: client.NamedReference{Name: userName},
	}
	if !data.SecretAccessKey.IsNull() && !data.SecretAccessKey.IsUnknown() {
		post.SecretAccessKey = data.SecretAccessKey.ValueString()
	}

	// When secret_access_key is provided, the API requires ?names= as well.
	// The user must set name to the full key name (e.g. "account/admin/PSFBxxxxxxxx").
	var names string
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		names = data.Name.ValueString()
	}

	key, err := r.client.PostObjectStoreAccessKey(ctx, names, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store access key", err.Error())
		return
	}

	// Map all response fields — secret_access_key is only available here.
	data.Name = types.StringValue(key.Name)
	data.AccessKeyID = types.StringValue(key.AccessKeyID)
	// secret_access_key handling:
	// - When user provided it (cross-array): keep the planned value exactly as-is.
	//   The API may not echo it back, and Terraform requires the state value to
	//   match the planned value for sensitive attributes.
	// - When user did NOT provide it (API-generated): use the API response.
	if key.SecretAccessKey != "" && (data.SecretAccessKey.IsNull() || data.SecretAccessKey.IsUnknown()) {
		data.SecretAccessKey = types.StringValue(key.SecretAccessKey)
	}
	// If user provided secret_access_key, data.SecretAccessKey already holds the
	// planned value — do not overwrite it.
	data.User = types.StringValue(key.User.Name)
	data.Created = types.Int64Value(key.Created)
	data.Enabled = types.BoolValue(key.Enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// SecretAccessKey is intentionally not set — it is a write-only attribute and is never
// returned by GET. The framework ensures write-only values are always null in state.
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
	data.User = types.StringValue(key.User.Name)
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

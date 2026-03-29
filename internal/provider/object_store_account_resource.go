package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure objectStoreAccountResource satisfies the resource interfaces.
var _ resource.Resource = &objectStoreAccountResource{}
var _ resource.ResourceWithConfigure = &objectStoreAccountResource{}
var _ resource.ResourceWithImportState = &objectStoreAccountResource{}
var _ resource.ResourceWithUpgradeState = &objectStoreAccountResource{}

// objectStoreAccountResource implements the flashblade_object_store_account resource.
type objectStoreAccountResource struct {
	client *client.FlashBladeClient
}

// NewObjectStoreAccountResource is the factory function registered in the provider.
func NewObjectStoreAccountResource() resource.Resource {
	return &objectStoreAccountResource{}
}

// ---------- model structs ----------------------------------------------------

// objectStoreAccountModel is the top-level model for the flashblade_object_store_account resource.
type objectStoreAccountModel struct {
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Created             types.Int64    `tfsdk:"created"`
	QuotaLimit          types.Int64    `tfsdk:"quota_limit"`
	HardLimitEnabled    types.Bool     `tfsdk:"hard_limit_enabled"`
	ObjectCount         types.Int64    `tfsdk:"object_count"`
	Space               types.Object   `tfsdk:"space"`
	SkipDefaultExport   types.Bool     `tfsdk:"skip_default_export"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *objectStoreAccountResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_object_store_account"
}

// Schema defines the resource schema.
func (r *objectStoreAccountResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade object store account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the object store account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the object store account. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp (milliseconds) when the account was created.",
				PlanModifiers: []planmodifier.Int64{
					int64UseStateForUnknown(),
				},
			},
			"quota_limit": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The effective quota limit applied against the size of the account, in bytes.",
			},
			"hard_limit_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the account's size cannot exceed the quota limit.",
			},
			"object_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The count of objects within the account.",
			},
			"space": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Storage space breakdown (read-only, API-managed).",
				Attributes: map[string]schema.Attribute{
					"data_reduction": schema.Float64Attribute{
						Computed:    true,
						Description: "Data reduction ratio.",
					},
					"snapshots": schema.Int64Attribute{
						Computed:    true,
						Description: "Physical space used by snapshots in bytes.",
					},
					"total_physical": schema.Int64Attribute{
						Computed:    true,
						Description: "Total physical space used in bytes.",
					},
					"unique": schema.Int64Attribute{
						Computed:    true,
						Description: "Unique physical space used in bytes.",
					},
					"virtual": schema.Int64Attribute{
						Computed:    true,
						Description: "Virtual (logical) space used in bytes.",
					},
					"snapshots_effective": schema.Int64Attribute{
						Computed:    true,
						Description: "Effective snapshot space used in bytes.",
					},
				},
			},
			"skip_default_export": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, suppresses the default account export to _array_server at creation time. Use this when you manage exports explicitly via flashblade_object_store_account_export.",
				Default:     booldefault.StaticBool(false),
				Computed:    true,
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
func (r *objectStoreAccountResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *objectStoreAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new object store account.
func (r *objectStoreAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data objectStoreAccountModel
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

	post := client.ObjectStoreAccountPost{}
	if !data.QuotaLimit.IsNull() && !data.QuotaLimit.IsUnknown() {
		post.QuotaLimit = strconv.FormatInt(data.QuotaLimit.ValueInt64(), 10)
	}
	if !data.HardLimitEnabled.IsNull() && !data.HardLimitEnabled.IsUnknown() {
		post.HardLimitEnabled = data.HardLimitEnabled.ValueBool()
	}
	if data.SkipDefaultExport.ValueBool() {
		// Send empty account_exports array to suppress the default _array_server export.
		emptyExports := json.RawMessage(`[]`)
		post.AccountExports = &emptyExports
	}

	_, err := r.client.PostObjectStoreAccount(ctx, data.Name.ValueString(), post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating object store account", err.Error())
		return
	}

	r.readIntoState(ctx, data.Name.ValueString(), &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *objectStoreAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data objectStoreAccountModel
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
	acct, err := r.client.GetObjectStoreAccount(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading object store account", err.Error())
		return
	}

	// Drift detection on mutable fields.
	if !data.QuotaLimit.IsNull() && !data.QuotaLimit.IsUnknown() {
		if data.QuotaLimit.ValueInt64() != acct.QuotaLimit {
			tflog.Info(ctx, "drift detected on object store account", map[string]any{
				"resource":    name,
				"field":       "quota_limit",
				"state_value": data.QuotaLimit.ValueInt64(),
				"api_value":   acct.QuotaLimit,
			})
		}
	}
	if !data.HardLimitEnabled.IsNull() && !data.HardLimitEnabled.IsUnknown() {
		if data.HardLimitEnabled.ValueBool() != acct.HardLimitEnabled {
			tflog.Info(ctx, "drift detected on object store account", map[string]any{
				"resource":    name,
				"field":       "hard_limit_enabled",
				"state_value": data.HardLimitEnabled.ValueBool(),
				"api_value":   acct.HardLimitEnabled,
			})
		}
	}

	resp.Diagnostics.Append(mapOSAToModel(acct, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing object store account.
func (r *objectStoreAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state objectStoreAccountModel
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

	patch := client.ObjectStoreAccountPatch{}

	if !plan.QuotaLimit.Equal(state.QuotaLimit) {
		v := strconv.FormatInt(plan.QuotaLimit.ValueInt64(), 10)
		patch.QuotaLimit = &v
	}
	if !plan.HardLimitEnabled.Equal(state.HardLimitEnabled) {
		v := plan.HardLimitEnabled.ValueBool()
		patch.HardLimitEnabled = &v
	}

	_, err := r.client.PatchObjectStoreAccount(ctx, state.Name.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating object store account", err.Error())
		return
	}

	r.readIntoState(ctx, plan.Name.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an object store account (single-phase, no soft-delete).
// Before deleting, checks for existing buckets and blocks if any are found.
func (r *objectStoreAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data objectStoreAccountModel
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

	// Guard: list active (non-destroyed) buckets in this account. Prevent delete if any exist.
	notDestroyed := false
	activeBuckets, err := r.client.ListBuckets(ctx, client.ListBucketsOpts{
		AccountNames: []string{name},
		Destroyed:    &notDestroyed,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error checking buckets before account deletion", err.Error())
		return
	}
	if len(activeBuckets) > 0 {
		resp.Diagnostics.AddError(
			"Cannot delete object store account",
			fmt.Sprintf("Account %q has %d active bucket(s). Destroy all buckets in the account before deleting the account.", name, len(activeBuckets)),
		)
		return
	}

	// Eradicate any soft-deleted buckets remaining in the account.
	// FlashBlade refuses to delete an account that has soft-deleted buckets.
	destroyed := true
	softDeleted, err := r.client.ListBuckets(ctx, client.ListBucketsOpts{
		AccountNames: []string{name},
		Destroyed:    &destroyed,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing soft-deleted buckets before account deletion", err.Error())
		return
	}
	for _, b := range softDeleted {
		if err := r.client.DeleteBucket(ctx, b.ID); err != nil {
			resp.Diagnostics.AddError(
				"Error eradicating soft-deleted bucket",
				fmt.Sprintf("Failed to eradicate bucket %q (ID: %s) before account deletion: %s", b.Name, b.ID, err.Error()),
			)
			return
		}
	}

	// Delete any object store users in the account.
	// FlashBlade refuses to delete an account that has users.
	if err := r.client.DeleteObjectStoreUser(ctx, name+"/admin"); err != nil {
		if !client.IsNotFound(err) {
			// Tolerate "user does not exist" (400) — may have been cleaned up already.
			var apiErr *client.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
				resp.Diagnostics.AddError("Error deleting object store user before account deletion", err.Error())
				return
			}
		}
	}

	if err := r.client.DeleteObjectStoreAccount(ctx, name); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting object store account", err.Error())
		return
	}
}

// ImportState imports an existing object store account by name.
func (r *objectStoreAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	var data objectStoreAccountModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	// Set Name so Read can look up the account.
	data.Name = types.StringValue(name)

	r.readIntoState(ctx, name, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetObjectStoreAccount and maps the result into the provided model.
func (r *objectStoreAccountResource) readIntoState(ctx context.Context, name string, data *objectStoreAccountModel, diags DiagnosticReporter) {
	acct, err := r.client.GetObjectStoreAccount(ctx, name)
	if err != nil {
		diags.AddError("Error reading object store account after write", err.Error())
		return
	}
	for _, d := range mapOSAToModel(acct, data) {
		diags.AddError(d.Summary(), d.Detail())
	}
}

// mapOSAToModel maps a client.ObjectStoreAccount to an objectStoreAccountModel.
// It preserves user-managed fields (Timeouts).
// Returns diagnostics instead of panicking on object construction errors.
func mapOSAToModel(acct *client.ObjectStoreAccount, data *objectStoreAccountModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(acct.ID)
	data.Name = types.StringValue(acct.Name)
	data.Created = types.Int64Value(acct.Created)
	data.QuotaLimit = types.Int64Value(acct.QuotaLimit)
	data.HardLimitEnabled = types.BoolValue(acct.HardLimitEnabled)
	data.ObjectCount = types.Int64Value(acct.ObjectCount)

	spaceObj, spaceDiags := mapSpaceToObject(acct.Space)
	diags.Append(spaceDiags...)
	if diags.HasError() {
		return diags
	}
	data.Space = spaceObj

	return diags
}

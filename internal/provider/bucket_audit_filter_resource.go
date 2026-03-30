package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// Ensure bucketAuditFilterResource satisfies the resource interfaces.
var _ resource.Resource = &bucketAuditFilterResource{}
var _ resource.ResourceWithConfigure = &bucketAuditFilterResource{}
var _ resource.ResourceWithImportState = &bucketAuditFilterResource{}

// bucketAuditFilterResource implements the flashblade_bucket_audit_filter resource.
type bucketAuditFilterResource struct {
	client *client.FlashBladeClient
}

// NewBucketAuditFilterResource is the factory function registered in the provider.
func NewBucketAuditFilterResource() resource.Resource {
	return &bucketAuditFilterResource{}
}

// ---------- model structs ----------------------------------------------------

// bucketAuditFilterModel is the Terraform state model for the flashblade_bucket_audit_filter resource.
type bucketAuditFilterModel struct {
	ID         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	BucketName types.String   `tfsdk:"bucket_name"`
	Actions    types.Set      `tfsdk:"actions"`
	S3Prefixes types.Set      `tfsdk:"s3_prefixes"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

// Metadata sets the Terraform type name.
func (r *bucketAuditFilterResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_audit_filter"
}

// Schema defines the resource schema.
func (r *bucketAuditFilterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade bucket audit filter for per-bucket S3 operation audit filtering.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the audit filter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the audit filter (1-63 alphanumeric characters, must start/end with letter or number).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket this audit filter belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"actions": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Set of S3 actions to audit (e.g. s3:GetObject, s3:PutObject). Order-independent.",
			},
			"s3_prefixes": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Description: "Set of S3 object key prefixes to filter audit events. Defaults to empty set (all prefixes).",
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

// Configure injects the FlashBladeClient into the resource.
func (r *bucketAuditFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new bucket audit filter.
func (r *bucketAuditFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketAuditFilterModel
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

	var actions []string
	resp.Diagnostics.Append(data.Actions.ElementsAs(ctx, &actions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var s3Prefixes []string
	resp.Diagnostics.Append(data.S3Prefixes.ElementsAs(ctx, &s3Prefixes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.BucketAuditFilterPost{
		Actions:    actions,
		S3Prefixes: s3Prefixes,
	}

	filter, err := r.client.PostBucketAuditFilter(ctx, data.Name.ValueString(), data.BucketName.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket audit filter", err.Error())
		return
	}

	mapBucketAuditFilterToModel(filter, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes Terraform state from the API.
func (r *bucketAuditFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketAuditFilterModel
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

	filter, err := r.client.GetBucketAuditFilter(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket audit filter", err.Error())
		return
	}

	mapBucketAuditFilterToModel(filter, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing bucket audit filter.
func (r *bucketAuditFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state bucketAuditFilterModel
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

	patch := client.BucketAuditFilterPatch{}
	needsPatch := false

	if !plan.Actions.Equal(state.Actions) {
		var actions []string
		resp.Diagnostics.Append(plan.Actions.ElementsAs(ctx, &actions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Actions = &actions
		needsPatch = true
	}

	if !plan.S3Prefixes.Equal(state.S3Prefixes) {
		var s3Prefixes []string
		resp.Diagnostics.Append(plan.S3Prefixes.ElementsAs(ctx, &s3Prefixes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.S3Prefixes = &s3Prefixes
		needsPatch = true
	}

	if needsPatch {
		_, err := r.client.PatchBucketAuditFilter(ctx, state.Name.ValueString(), patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating bucket audit filter", err.Error())
			return
		}
	}

	// Re-read to refresh computed fields.
	filter, err := r.client.GetBucketAuditFilter(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket audit filter after update", err.Error())
		return
	}

	mapBucketAuditFilterToModel(filter, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes a bucket audit filter.
func (r *bucketAuditFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketAuditFilterModel
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

	err := r.client.DeleteBucketAuditFilter(ctx, data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting bucket audit filter", err.Error())
		return
	}
}

// ImportState imports an existing bucket audit filter by filter name.
func (r *bucketAuditFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	filterName := req.ID

	filter, err := r.client.GetBucketAuditFilter(ctx, filterName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing bucket audit filter", err.Error())
		return
	}

	var data bucketAuditFilterModel
	data.Timeouts = nullTimeoutsValue()

	mapBucketAuditFilterToModel(filter, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// mapBucketAuditFilterToModel maps a client.BucketAuditFilter to the Terraform model.
func mapBucketAuditFilterToModel(filter *client.BucketAuditFilter, data *bucketAuditFilterModel) {
	data.ID = types.StringValue(filter.Name)
	data.Name = types.StringValue(filter.Name)
	data.BucketName = types.StringValue(filter.Bucket.Name)

	// Convert []string to types.Set for actions (order-independent).
	data.Actions = types.SetValueMust(types.StringType, stringSliceToAttrValues(filter.Actions))

	// Convert []string to types.Set for s3_prefixes.
	prefixes := filter.S3Prefixes
	if prefixes == nil {
		prefixes = []string{}
	}
	data.S3Prefixes = types.SetValueMust(types.StringType, stringSliceToAttrValues(prefixes))
}

// stringSliceToAttrValues converts a []string to []attr.Value for use with types.ListValueMust.
func stringSliceToAttrValues(ss []string) []attr.Value {
	vals := make([]attr.Value, len(ss))
	for i, s := range ss {
		vals[i] = types.StringValue(s)
	}
	return vals
}

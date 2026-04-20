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

var _ resource.Resource = &bucketAccessPolicyResource{}
var _ resource.ResourceWithConfigure = &bucketAccessPolicyResource{}
var _ resource.ResourceWithImportState = &bucketAccessPolicyResource{}
var _ resource.ResourceWithUpgradeState = &bucketAccessPolicyResource{}

// bucketAccessPolicyResource implements the flashblade_bucket_access_policy resource.
type bucketAccessPolicyResource struct {
	client *client.FlashBladeClient
}

func NewBucketAccessPolicyResource() resource.Resource {
	return &bucketAccessPolicyResource{}
}

// ---------- model structs ----------------------------------------------------

// bucketAccessPolicyModel is the Terraform state model for the flashblade_bucket_access_policy resource.
type bucketAccessPolicyModel struct {
	ID         types.String   `tfsdk:"id"`
	BucketName types.String   `tfsdk:"bucket_name"`
	Enabled    types.Bool     `tfsdk:"enabled"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *bucketAccessPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_access_policy"
}

// Schema defines the resource schema.
func (r *bucketAccessPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages a FlashBlade bucket access policy for IAM-style per-bucket authorization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the bucket access policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket this policy belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the bucket access policy is enabled. Read-only, managed by the array.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *bucketAccessPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketAccessPolicyModel
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

	body := client.BucketAccessPolicyPost{}

	policy, err := r.client.PostBucketAccessPolicy(ctx, data.BucketName.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket access policy", err.Error())
		return
	}

	mapBucketAccessPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bucketAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketAccessPolicyModel
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

	policy, err := r.client.GetBucketAccessPolicy(ctx, data.BucketName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket access policy", err.Error())
		return
	}

	mapBucketAccessPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op — bucket access policies have no PATCH endpoint.
// All mutable fields trigger RequiresReplace, so this method should never be called.
func (r *bucketAccessPolicyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Bucket access policies cannot be updated in place. All changes require replacement.",
	)
}

// Delete removes a bucket access policy.
func (r *bucketAccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketAccessPolicyModel
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

	err := r.client.DeleteBucketAccessPolicy(ctx, data.BucketName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting bucket access policy", err.Error())
		return
	}
}

// ImportState imports an existing bucket access policy by bucket name.
func (r *bucketAccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	bucketName := req.ID

	policy, err := r.client.GetBucketAccessPolicy(ctx, bucketName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing bucket access policy", err.Error())
		return
	}

	var data bucketAccessPolicyModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	mapBucketAccessPolicyToModel(policy, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// UpgradeState returns state upgraders for schema migrations.
func (r *bucketAccessPolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// ---------- helpers ---------------------------------------------------------

// mapBucketAccessPolicyToModel maps a client.BucketAccessPolicy to the Terraform model.
func mapBucketAccessPolicyToModel(policy *client.BucketAccessPolicy, data *bucketAccessPolicyModel) {
	data.ID = types.StringValue(policy.ID)
	data.BucketName = types.StringValue(policy.Bucket.Name)
	data.Enabled = types.BoolValue(policy.Enabled)
}

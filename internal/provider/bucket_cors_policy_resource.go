package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &bucketCorsPolicyResource{}
var _ resource.ResourceWithConfigure = &bucketCorsPolicyResource{}
var _ resource.ResourceWithImportState = &bucketCorsPolicyResource{}
var _ resource.ResourceWithUpgradeState = &bucketCorsPolicyResource{}

// FlashBlade only supports a fully permissive (wildcard) CORS rule today: origins
// and headers must be "*", and methods are all-or-nothing. This resource is therefore
// a per-bucket toggle — its presence applies the single wildcard rule below, letting
// browsers do cross-origin PUT/GET/POST against presigned URLs. When FlashBlade gains
// granular CORS, this resource can grow a configurable rules list.
const corsWildcardRuleName = "corsrule"

var corsWildcardMethods = []string{"GET", "PUT", "HEAD", "POST", "DELETE"}

func corsWildcardRule() client.CorsRulePost {
	return client.CorsRulePost{
		AllowedOrigins: []string{"*"},
		AllowedMethods: corsWildcardMethods,
		AllowedHeaders: []string{"*"},
	}
}

// bucketCorsPolicyResource implements the flashblade_bucket_cors_policy resource.
// The real CORS API creates the policy with an EMPTY body (auto-named) and manages
// rules on the /rules sub-endpoint (no PATCH). Because only a wildcard rule is
// supported, this resource ensures the policy plus one permissive rule per bucket.
type bucketCorsPolicyResource struct {
	client *client.FlashBladeClient
}

func NewBucketCorsPolicyResource() resource.Resource {
	return &bucketCorsPolicyResource{}
}

// bucketCorsPolicyModel is the Terraform state model for flashblade_bucket_cors_policy.
type bucketCorsPolicyModel struct {
	ID         types.String   `tfsdk:"id"`
	BucketName types.String   `tfsdk:"bucket_name"`
	IsLocal    types.Bool     `tfsdk:"is_local"`
	PolicyType types.String   `tfsdk:"policy_type"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

func (r *bucketCorsPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_cors_policy"
}

func (r *bucketCorsPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Enables a permissive (wildcard) CORS policy on a FlashBlade bucket so browsers can perform cross-origin requests against presigned URLs. FlashBlade only supports fully permissive CORS today (origins and headers are '*', all HTTP methods allowed), so this resource is a per-bucket toggle: its presence applies the wildcard rule, and destroying it removes the policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the CORS policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket this CORS policy belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the CORS policy is local to this array. Read-only.",
			},
			"policy_type": schema.StringAttribute{
				Computed:    true,
				Description: "The policy type. Read-only, managed by the array.",
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

func (r *bucketCorsPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketCorsPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketCorsPolicyModel
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

	if err := r.applyWildcardPolicy(ctx, data.BucketName.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating bucket CORS policy", err.Error())
		return
	}

	policy, err := r.client.GetCorsPolicy(ctx, data.BucketName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket CORS policy after create", err.Error())
		return
	}

	mapCorsPolicyToModel(policy, data.BucketName.ValueString(), &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bucketCorsPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketCorsPolicyModel
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

	name := data.BucketName.ValueString()
	policy, err := r.client.GetCorsPolicy(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket CORS policy", err.Error())
		return
	}

	logCorsDrift(ctx, name, &data, policy)

	mapCorsPolicyToModel(policy, name, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bucketCorsPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data bucketCorsPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// No mutable inputs (bucket_name forces replacement). Re-apply the wildcard rule
	// so the policy self-heals if it drifted, then read back.
	if err := r.applyWildcardPolicy(ctx, data.BucketName.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating bucket CORS policy", err.Error())
		return
	}

	policy, err := r.client.GetCorsPolicy(ctx, data.BucketName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket CORS policy after update", err.Error())
		return
	}

	mapCorsPolicyToModel(policy, data.BucketName.ValueString(), &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bucketCorsPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketCorsPolicyModel
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

	err := r.client.DeleteCorsPolicy(ctx, data.BucketName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting bucket CORS policy", err.Error())
		return
	}
}

func (r *bucketCorsPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policy, err := r.client.GetCorsPolicy(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing bucket CORS policy", err.Error())
		return
	}

	var data bucketCorsPolicyModel
	data.Timeouts = nullTimeoutsValue()
	mapCorsPolicyToModel(policy, req.ID, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bucketCorsPolicyResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// ---------- helpers ---------------------------------------------------------

// applyWildcardPolicy ensures the bucket's CORS policy exists and carries the single
// permissive wildcard rule. It is idempotent and self-healing: the policy POST tolerates
// an existing policy, and the rule is deleted before being (re)created so a retry — or a
// rule left over on the array by an earlier partial apply — does not fail with the array's
// non-idempotent "Rule already exists." (HTTP 400). This mirrors the proven dbauth flow
// (ensure policy -> delete rule -> post rule).
func (r *bucketCorsPolicyResource) applyWildcardPolicy(ctx context.Context, bucket string) error {
	if err := r.client.EnsureCorsPolicy(ctx, bucket); err != nil {
		return err
	}
	if err := r.client.DeleteCorsRule(ctx, bucket, corsWildcardRuleName); err != nil {
		return err
	}
	return r.client.PostCorsRule(ctx, bucket, corsWildcardRuleName, corsWildcardRule())
}

// mapCorsPolicyToModel maps a client.CrossOriginResourceSharingPolicy into the state model.
// The array does not return a usable policy id (the CORS policy name is auto-generated
// and the GET response's id comes back empty), so the bucket name is used as the resource
// id — it is unique (one CORS policy per bucket), stable, and matches import-by-bucket-name.
// A non-empty id is also required by the Pulumi bridge (Create must return a resource ID).
func mapCorsPolicyToModel(policy *client.CrossOriginResourceSharingPolicy, bucketName string, data *bucketCorsPolicyModel) {
	data.ID = types.StringValue(bucketName)
	data.BucketName = types.StringValue(bucketName)
	data.IsLocal = types.BoolValue(policy.IsLocal)
	data.PolicyType = types.StringValue(policy.PolicyType)
}

// logCorsDrift logs field-level drift between state and the API response.
func logCorsDrift(ctx context.Context, name string, data *bucketCorsPolicyModel, policy *client.CrossOriginResourceSharingPolicy) {
	if data.IsLocal.ValueBool() != policy.IsLocal {
		tflog.Debug(ctx, "drift detected", map[string]any{"resource": name, "field": "is_local", "was": data.IsLocal.ValueBool(), "now": policy.IsLocal})
	}
	if data.PolicyType.ValueString() != policy.PolicyType {
		tflog.Debug(ctx, "drift detected", map[string]any{"resource": name, "field": "policy_type", "was": data.PolicyType.ValueString(), "now": policy.PolicyType})
	}
}

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &bucketAccessPolicyRuleResource{}
var _ resource.ResourceWithConfigure = &bucketAccessPolicyRuleResource{}
var _ resource.ResourceWithImportState = &bucketAccessPolicyRuleResource{}
var _ resource.ResourceWithUpgradeState = &bucketAccessPolicyRuleResource{}

// bucketAccessPolicyRuleResource implements the flashblade_bucket_access_policy_rule resource.
type bucketAccessPolicyRuleResource struct {
	client *client.FlashBladeClient
}

func NewBucketAccessPolicyRuleResource() resource.Resource {
	return &bucketAccessPolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// bucketAccessPolicyRuleModel is the Terraform state model for the flashblade_bucket_access_policy_rule resource.
type bucketAccessPolicyRuleModel struct {
	Name       types.String   `tfsdk:"name"`
	BucketName types.String   `tfsdk:"bucket_name"`
	Actions    types.List     `tfsdk:"actions"`
	Effect     types.String   `tfsdk:"effect"`
	Principals types.List     `tfsdk:"principals"`
	Resources  types.List     `tfsdk:"resources"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *bucketAccessPolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_bucket_access_policy_rule"
}

// Schema defines the resource schema.
func (r *bucketAccessPolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages an individual rule within a FlashBlade bucket access policy.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The rule name. When provided, the rule is created with this name. When omitted, the API assigns one automatically.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the bucket this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"actions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of S3 actions this rule applies to (e.g. s3:GetObject).",
			},
			"effect": schema.StringAttribute{
				Computed:    true,
				Description: "The effect of the rule. Always 'allow' — set by the API.",
			},
			"principals": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of principals this rule applies to (mapped to principals.all in the API). Note: the accepted format depends on the FlashBlade firmware version — consult your array documentation for valid principal values.",
			},
			"resources": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of S3 resource ARNs this rule applies to.",
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
func (r *bucketAccessPolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bucketAccessPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bucketAccessPolicyRuleModel
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

	actions, diags := listToStrings(ctx, data.Actions)
	resp.Diagnostics.Append(diags...)
	principals, diags := listToStrings(ctx, data.Principals)
	resp.Diagnostics.Append(diags...)
	resources, diags := listToStrings(ctx, data.Resources)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.BucketAccessPolicyRulePost{
		Actions: actions,
		Principals: client.BucketAccessPolicyPrincipals{
			All: principals,
		},
		Resources: resources,
	}

	rule, err := r.client.PostBucketAccessPolicyRule(ctx, data.BucketName.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket access policy rule", err.Error())
		return
	}

	resp.Diagnostics.Append(mapBucketAccessPolicyRuleToModel(ctx, rule, data.BucketName.ValueString(), &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bucketAccessPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bucketAccessPolicyRuleModel
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

	rule, err := r.client.GetBucketAccessPolicyRule(ctx, data.BucketName.ValueString(), data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bucket access policy rule", err.Error())
		return
	}

	resp.Diagnostics.Append(mapBucketAccessPolicyRuleToModel(ctx, rule, data.BucketName.ValueString(), &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op — bucket access policy rules have no PATCH endpoint.
// All mutable fields trigger RequiresReplace, so this method should never be called.
func (r *bucketAccessPolicyRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Bucket access policy rules cannot be updated in place. All changes require replacement.",
	)
}

// Delete removes a bucket access policy rule.
func (r *bucketAccessPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bucketAccessPolicyRuleModel
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

	err := r.client.DeleteBucketAccessPolicyRule(ctx, data.BucketName.ValueString(), data.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting bucket access policy rule", err.Error())
		return
	}
}

// ImportState imports an existing bucket access policy rule by composite ID "bucketName/ruleName".
func (r *bucketAccessPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: bucketName/ruleName. Error: %s", err))
		return
	}

	rule, err := r.client.GetBucketAccessPolicyRule(ctx, parts[0], parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error importing bucket access policy rule", err.Error())
		return
	}

	var data bucketAccessPolicyRuleModel
	data.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"read":   types.StringType,
			"delete": types.StringType,
		}),
	}

	resp.Diagnostics.Append(mapBucketAccessPolicyRuleToModel(ctx, rule, parts[0], &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// UpgradeState returns state upgraders for schema migrations.
func (r *bucketAccessPolicyRuleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// ---------- helpers ---------------------------------------------------------

// mapBucketAccessPolicyRuleToModel maps a client.BucketAccessPolicyRule to the Terraform model.
func mapBucketAccessPolicyRuleToModel(ctx context.Context, rule *client.BucketAccessPolicyRule, bucketName string, data *bucketAccessPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Name = types.StringValue(rule.Name)
	data.BucketName = types.StringValue(bucketName)
	data.Effect = types.StringValue(rule.Effect)

	actionsList, d := types.ListValueFrom(ctx, types.StringType, rule.Actions)
	diags.Append(d...)
	data.Actions = actionsList

	principalsList, d := types.ListValueFrom(ctx, types.StringType, rule.Principals.All)
	diags.Append(d...)
	data.Principals = principalsList

	resourcesList, d := types.ListValueFrom(ctx, types.StringType, rule.Resources)
	diags.Append(d...)
	data.Resources = resourcesList

	return diags
}

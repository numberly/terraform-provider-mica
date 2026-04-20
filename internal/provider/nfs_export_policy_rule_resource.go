package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ resource.Resource = &nfsExportPolicyRuleResource{}
var _ resource.ResourceWithConfigure = &nfsExportPolicyRuleResource{}
var _ resource.ResourceWithImportState = &nfsExportPolicyRuleResource{}
var _ resource.ResourceWithUpgradeState = &nfsExportPolicyRuleResource{}

// nfsExportPolicyRuleResource implements the flashblade_nfs_export_policy_rule resource.
type nfsExportPolicyRuleResource struct {
	client *client.FlashBladeClient
}

func NewNfsExportPolicyRuleResource() resource.Resource {
	return &nfsExportPolicyRuleResource{}
}

// ---------- model structs ----------------------------------------------------

// nfsExportPolicyRuleModel is the top-level model for the flashblade_nfs_export_policy_rule resource.
type nfsExportPolicyRuleModel struct {
	ID                        types.String   `tfsdk:"id"`
	PolicyName                types.String   `tfsdk:"policy_name"`
	Name                      types.String   `tfsdk:"name"`
	Index                     types.Int64    `tfsdk:"index"`
	PolicyVersion             types.String   `tfsdk:"policy_version"`
	Access                    types.String   `tfsdk:"access"`
	Client                    types.String   `tfsdk:"client"`
	Permission                types.String   `tfsdk:"permission"`
	Anonuid                   types.Int64    `tfsdk:"anonuid"`
	Anongid                   types.Int64    `tfsdk:"anongid"`
	Atime                     types.Bool     `tfsdk:"atime"`
	Fileid32bit               types.Bool     `tfsdk:"fileid_32bit"`
	Secure                    types.Bool     `tfsdk:"secure"`
	Security                  types.List     `tfsdk:"security"`
	RequiredTransportSecurity types.String   `tfsdk:"required_transport_security"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
}

// ---------- resource interface methods --------------------------------------

func (r *nfsExportPolicyRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_nfs_export_policy_rule"
}

// Schema defines the resource schema.
func (r *nfsExportPolicyRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a rule within a FlashBlade NFS export policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the NFS export policy rule (server-assigned UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the NFS export policy this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The server-assigned rule identifier. Used internally for PATCH/DELETE API calls.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"index": schema.Int64Attribute{
				Computed:    true,
				Description: "The server-assigned ordering index for this rule within the policy. Used for import.",
			},
			"policy_version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of the parent policy at the time this rule was last read.",
			},
			"access": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The access control for NFS clients (e.g. 'root-squash', 'no-root-squash', 'all-squash').",
				Validators: []validator.String{
					stringvalidator.OneOf("root-squash", "no-root-squash", "all-squash"),
				},
			},
			"client": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A pattern matching the clients to which this rule applies (e.g. '*', '10.0.0.0/8').",
			},
			"permission": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The read/write permission for matching clients (e.g. 'rw', 'ro').",
				Validators: []validator.String{
					stringvalidator.OneOf("rw", "ro"),
				},
			},
			"anonuid": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The UID to use for anonymous (squashed) users.",
			},
			"anongid": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The GID to use for anonymous (squashed) users.",
			},
			"atime": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, access time updates are enabled.",
			},
			"fileid_32bit": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, use 32-bit file IDs.",
			},
			"secure": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, require clients to use a privileged port.",
			},
			"security": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Security flavors to enforce for this rule.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"required_transport_security": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Required transport security for this rule (e.g. 'krb5', 'krb5i', 'krb5p').",
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


// nfsExportPolicyRuleV0Model mirrors the schema as of Version 0. It is identical
// to the v1 model for this migration — the change was wire-level only (PATCH
// encoding of Security became pointer-aware), not model-level. The bump is
// mandated by CONVENTIONS.md §State Upgraders because a client model type changed.
type nfsExportPolicyRuleV0Model struct {
	ID                        types.String   `tfsdk:"id"`
	PolicyName                types.String   `tfsdk:"policy_name"`
	Name                      types.String   `tfsdk:"name"`
	Index                     types.Int64    `tfsdk:"index"`
	PolicyVersion             types.String   `tfsdk:"policy_version"`
	Access                    types.String   `tfsdk:"access"`
	Client                    types.String   `tfsdk:"client"`
	Permission                types.String   `tfsdk:"permission"`
	Anonuid                   types.Int64    `tfsdk:"anonuid"`
	Anongid                   types.Int64    `tfsdk:"anongid"`
	Atime                     types.Bool     `tfsdk:"atime"`
	Fileid32bit               types.Bool     `tfsdk:"fileid_32bit"`
	Secure                    types.Bool     `tfsdk:"secure"`
	Security                  types.List     `tfsdk:"security"`
	RequiredTransportSecurity types.String   `tfsdk:"required_transport_security"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
}

// UpgradeState returns state upgraders for schema migrations.
func (r *nfsExportPolicyRuleResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The unique identifier of the NFS export policy rule (server-assigned UUID).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"policy_name": schema.StringAttribute{
						Required:    true,
						Description: "The name of the NFS export policy this rule belongs to. Changing this forces a new resource.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The server-assigned rule identifier. Used internally for PATCH/DELETE API calls.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"index": schema.Int64Attribute{
						Computed:    true,
						Description: "The server-assigned ordering index for this rule within the policy. Used for import.",
					},
					"policy_version": schema.StringAttribute{
						Computed:    true,
						Description: "The version of the parent policy at the time this rule was last read.",
					},
					"access": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The access control for NFS clients (e.g. 'root-squash', 'no-root-squash', 'all-squash').",
						Validators: []validator.String{
							stringvalidator.OneOf("root-squash", "no-root-squash", "all-squash"),
						},
					},
					"client": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "A pattern matching the clients to which this rule applies (e.g. '*', '10.0.0.0/8').",
					},
					"permission": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The read/write permission for matching clients (e.g. 'rw', 'ro').",
						Validators: []validator.String{
							stringvalidator.OneOf("rw", "ro"),
						},
					},
					"anonuid": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The UID to use for anonymous (squashed) users.",
					},
					"anongid": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The GID to use for anonymous (squashed) users.",
					},
					"atime": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "If true, access time updates are enabled.",
					},
					"fileid_32bit": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "If true, use 32-bit file IDs.",
					},
					"secure": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "If true, require clients to use a privileged port.",
					},
					"security": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "Security flavors to enforce for this rule.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"required_transport_security": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Required transport security for this rule (e.g. 'krb5', 'krb5i', 'krb5p').",
					},
					"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
						Create: true,
						Read:   true,
						Update: true,
						Delete: true,
					}),
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var old nfsExportPolicyRuleV0Model
				resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
				if resp.Diagnostics.HasError() {
					return
				}
				// Identity migration: the v0 and v1 models are structurally identical;
				// the bump is convention-driven (client model type changed at wire level).
				newState := nfsExportPolicyRuleModel(old)
				resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
			},
		},
	}
}

// Configure injects the FlashBladeClient into the resource.
func (r *nfsExportPolicyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *nfsExportPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data nfsExportPolicyRuleModel
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

	policyName := data.PolicyName.ValueString()

	post := client.NfsExportPolicyRulePost{}
	if !data.Access.IsNull() && !data.Access.IsUnknown() {
		post.Access = data.Access.ValueString()
	}
	if !data.Client.IsNull() && !data.Client.IsUnknown() {
		post.Client = data.Client.ValueString()
	}
	if !data.Permission.IsNull() && !data.Permission.IsUnknown() {
		post.Permission = data.Permission.ValueString()
	}
	if !data.Anonuid.IsNull() && !data.Anonuid.IsUnknown() {
		post.Anonuid = int(data.Anonuid.ValueInt64())
	}
	if !data.Anongid.IsNull() && !data.Anongid.IsUnknown() {
		post.Anongid = int(data.Anongid.ValueInt64())
	}
	if !data.Atime.IsNull() && !data.Atime.IsUnknown() {
		v := data.Atime.ValueBool()
		post.Atime = &v
	}
	if !data.Fileid32bit.IsNull() && !data.Fileid32bit.IsUnknown() {
		v := data.Fileid32bit.ValueBool()
		post.Fileid32bit = &v
	}
	if !data.Secure.IsNull() && !data.Secure.IsUnknown() {
		v := data.Secure.ValueBool()
		post.Secure = &v
	}
	if !data.Security.IsNull() && !data.Security.IsUnknown() {
		var security []string
		resp.Diagnostics.Append(data.Security.ElementsAs(ctx, &security, false)...)
		if !resp.Diagnostics.HasError() {
			post.Security = security
		}
	}
	if !data.RequiredTransportSecurity.IsNull() && !data.RequiredTransportSecurity.IsUnknown() {
		post.RequiredTransportSecurity = data.RequiredTransportSecurity.ValueString()
	}

	created, err := r.client.PostNfsExportPolicyRule(ctx, policyName, post)
	if err != nil {
		resp.Diagnostics.AddError("Error creating NFS export policy rule", err.Error())
		return
	}

	// Read back full state by server-assigned rule name.
	readDiags := r.readIntoState(ctx, policyName, created.Name, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *nfsExportPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data nfsExportPolicyRuleModel
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

	policyName := data.PolicyName.ValueString()
	ruleName := data.Name.ValueString()

	rule, err := r.client.GetNfsExportPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading NFS export policy rule", err.Error())
		return
	}

	// Drift detection on mutable fields.
	if !data.Client.IsNull() && !data.Client.IsUnknown() {
		if data.Client.ValueString() != rule.Client {
			tflog.Debug(ctx, "drift detected on NFS export policy rule", map[string]any{
				"policy":      policyName,
				"rule":        ruleName,
				"field":       "client",
				"was":         data.Client.ValueString(),
				"now":           rule.Client,
			})
		}
	}

	mapDiags := mapNfsExportPolicyRuleToModel(ctx, rule, &data)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to an existing NFS export policy rule.
func (r *nfsExportPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state nfsExportPolicyRuleModel
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

	policyName := state.PolicyName.ValueString()
	ruleName := state.Name.ValueString()

	patch := client.NfsExportPolicyRulePatch{}

	if !plan.Access.Equal(state.Access) {
		v := plan.Access.ValueString()
		patch.Access = &v
	}
	if !plan.Client.Equal(state.Client) {
		v := plan.Client.ValueString()
		patch.Client = &v
	}
	if !plan.Permission.Equal(state.Permission) {
		v := plan.Permission.ValueString()
		patch.Permission = &v
	}
	if !plan.Anonuid.Equal(state.Anonuid) {
		v := strconv.FormatInt(plan.Anonuid.ValueInt64(), 10)
		patch.Anonuid = &v
	}
	if !plan.Anongid.Equal(state.Anongid) {
		v := strconv.FormatInt(plan.Anongid.ValueInt64(), 10)
		patch.Anongid = &v
	}
	if !plan.Atime.Equal(state.Atime) {
		v := plan.Atime.ValueBool()
		patch.Atime = &v
	}
	if !plan.Fileid32bit.Equal(state.Fileid32bit) {
		v := plan.Fileid32bit.ValueBool()
		patch.Fileid32bit = &v
	}
	if !plan.Secure.Equal(state.Secure) {
		v := plan.Secure.ValueBool()
		patch.Secure = &v
	}
	if !plan.Security.Equal(state.Security) {
		var security []string
		resp.Diagnostics.Append(plan.Security.ElementsAs(ctx, &security, false)...)
		if !resp.Diagnostics.HasError() {
			if security == nil {
				security = []string{}
			}
			patch.Security = &security
		}
	}
	if !plan.RequiredTransportSecurity.Equal(state.RequiredTransportSecurity) {
		v := plan.RequiredTransportSecurity.ValueString()
		patch.RequiredTransportSecurity = &v
	}

	_, err := r.client.PatchNfsExportPolicyRule(ctx, policyName, ruleName, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating NFS export policy rule", err.Error())
		return
	}

	readDiags := r.readIntoState(ctx, policyName, ruleName, &plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an NFS export policy rule.
func (r *nfsExportPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data nfsExportPolicyRuleModel
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

	policyName := data.PolicyName.ValueString()
	ruleName := data.Name.ValueString()

	if err := r.client.DeleteNfsExportPolicyRule(ctx, policyName, ruleName); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting NFS export policy rule", err.Error())
		return
	}
}

// ImportState imports an existing NFS export policy rule using composite ID "policy_name/rule_index".
func (r *nfsExportPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := parseCompositeID(req.ID, 2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected 'policy_name/rule_index', got: %q. Example: 'my-policy/0'. %s", req.ID, err),
		)
		return
	}

	policyName := parts[0]
	indexStr := parts[1]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid rule index in import ID",
			fmt.Sprintf("The rule index %q is not a valid integer: %s", indexStr, err),
		)
		return
	}

	rule, err := r.client.GetNfsExportPolicyRuleByIndex(ctx, policyName, index)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error finding NFS export policy rule by index",
			fmt.Sprintf("Could not find rule at index %d in policy %q: %s", index, policyName, err),
		)
		return
	}

	var data nfsExportPolicyRuleModel
	// Initialize timeouts with a null value so the framework can serialize it.
	data.Timeouts = nullTimeoutsValue()
	data.PolicyName = types.StringValue(policyName)
	data.Name = types.StringValue(rule.Name)

	readDiags := r.readIntoState(ctx, policyName, rule.Name, &data)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState calls GetNfsExportPolicyRuleByName and maps the result into the provided model.
// Returns diagnostics from the map operation.
func (r *nfsExportPolicyRuleResource) readIntoState(ctx context.Context, policyName, ruleName string, data *nfsExportPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	rule, err := r.client.GetNfsExportPolicyRuleByName(ctx, policyName, ruleName)
	if err != nil {
		diags.AddError("Error reading NFS export policy rule after write", err.Error())
		return diags
	}
	diags.Append(mapNfsExportPolicyRuleToModel(ctx, rule, data)...)
	return diags
}

// mapNfsExportPolicyRuleToModel maps a client.NfsExportPolicyRule to an nfsExportPolicyRuleModel.
func mapNfsExportPolicyRuleToModel(ctx context.Context, rule *client.NfsExportPolicyRule, data *nfsExportPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(rule.ID)
	data.Name = types.StringValue(rule.Name)
	data.Index = types.Int64Value(int64(rule.Index))
	data.PolicyVersion = types.StringValue(rule.PolicyVersion)
	data.PolicyName = types.StringValue(rule.Policy.Name)
	data.Access = types.StringValue(rule.Access)
	data.Client = types.StringValue(rule.Client)
	data.Permission = types.StringValue(rule.Permission)
	data.Anonuid = types.Int64Value(int64(rule.Anonuid))
	data.Anongid = types.Int64Value(int64(rule.Anongid))
	data.Atime = types.BoolValue(rule.Atime)
	data.Fileid32bit = types.BoolValue(rule.Fileid32bit)
	data.Secure = types.BoolValue(rule.Secure)
	data.RequiredTransportSecurity = types.StringValue(rule.RequiredTransportSecurity)

	if len(rule.Security) > 0 {
		securityList, secDiags := types.ListValueFrom(ctx, types.StringType, rule.Security)
		diags.Append(secDiags...)
		if !diags.HasError() {
			data.Security = securityList
		}
	} else {
		data.Security = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

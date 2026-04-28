package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/terraform-provider-mica/internal/client"
)

var _ resource.Resource = &arraySmtpResource{}
var _ resource.ResourceWithConfigure = &arraySmtpResource{}
var _ resource.ResourceWithImportState = &arraySmtpResource{}
var _ resource.ResourceWithUpgradeState = &arraySmtpResource{}

// arraySmtpResource implements the flashblade_array_smtp composite singleton resource.
// It manages both the SMTP relay config and alert watchers (email recipients) in a single resource.
type arraySmtpResource struct {
	client *client.FlashBladeClient
}

func NewArraySmtpResource() resource.Resource {
	return &arraySmtpResource{}
}

// ---------- model structs ----------------------------------------------------

// alertWatcherModel represents a single alert watcher entry in the schema.
type alertWatcherModel struct {
	Email                       types.String `tfsdk:"email"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	MinimumNotificationSeverity types.String `tfsdk:"minimum_notification_severity"`
}

// arraySmtpModel is the top-level model for the flashblade_array_smtp resource.
type arraySmtpModel struct {
	ID             types.String   `tfsdk:"id"`
	RelayHost      types.String   `tfsdk:"relay_host"`
	SenderDomain   types.String   `tfsdk:"sender_domain"`
	EncryptionMode types.String   `tfsdk:"encryption_mode"`
	AlertWatchers  types.Set      `tfsdk:"alert_watchers"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

// alertWatcherAttrTypes returns the attr.Type map for an alertWatcherModel.
func alertWatcherAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"email":                         types.StringType,
		"enabled":                       types.BoolType,
		"minimum_notification_severity": types.StringType,
	}
}

// ---------- resource interface methods --------------------------------------

func (r *arraySmtpResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "flashblade_array_smtp"
}

// Schema defines the resource schema.
func (r *arraySmtpResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     0,
		Description: "Manages the SMTP relay configuration and alert watchers of a FlashBlade array. This is a composite singleton resource — Delete resets SMTP config and removes all alert watchers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMTP server configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"relay_host": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Hostname or IP address of the SMTP relay server.",
			},
			"sender_domain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Domain appended to the sender email address.",
			},
			"encryption_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("none"),
				Description: "SMTP encryption mode: 'none', 'tls', or 'starttls'.",
				Validators: []validator.String{
					stringvalidator.OneOf("none", "tls", "starttls"),
				},
			},
			"alert_watchers": schema.SetNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set of alert watcher email recipients.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Required:    true,
							Description: "Email address of the alert recipient. This is the unique identifier for the watcher.",
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
							Description: "If true, this watcher receives alert notifications.",
						},
						"minimum_notification_severity": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("warning"),
							Description: "Minimum alert severity that triggers a notification: 'info', 'warning', 'error', or 'critical'.",
						},
					},
				},
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
func (r *arraySmtpResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// Configure injects the FlashBladeClient into the resource.
func (r *arraySmtpResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create patches SMTP config and creates all configured alert watchers.
func (r *arraySmtpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data arraySmtpModel
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

	// Patch SMTP relay config.
	relayHost := data.RelayHost.ValueString()
	senderDomain := data.SenderDomain.ValueString()
	encryptionMode := data.EncryptionMode.ValueString()
	patch := client.SmtpServerPatch{
		RelayHost:      &relayHost,
		SenderDomain:   &senderDomain,
		EncryptionMode: &encryptionMode,
	}
	_, err := r.client.PatchSmtpServer(ctx, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error configuring SMTP server", err.Error())
		return
	}

	// Create alert watchers.
	watchers := extractAlertWatchers(ctx, data.AlertWatchers, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, w := range watchers {
		email := w.Email.ValueString()
		severity := w.MinimumNotificationSeverity.ValueString()
		_, err := r.client.PostAlertWatcher(ctx, email, client.AlertWatcherPost{
			MinimumNotificationSeverity: severity,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating alert watcher",
				fmt.Sprintf("Failed to create alert watcher for %s: %s", email, err),
			)
			return
		}
		// AlertWatcherPost has no enabled field; the API defaults to enabled=true.
		// PATCH to apply the plan's enabled=false.
		if !w.Enabled.IsNull() && !w.Enabled.IsUnknown() && !w.Enabled.ValueBool() {
			disabled := false
			if _, err := r.client.PatchAlertWatcher(ctx, email, client.AlertWatcherPatch{Enabled: &disabled}); err != nil {
				resp.Diagnostics.AddError(
					"Error disabling alert watcher",
					fmt.Sprintf("Failed to set enabled=false on new alert watcher %s: %s", email, err),
				)
				return
			}
		}
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *arraySmtpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data arraySmtpModel
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

	resp.Diagnostics.Append(r.readIntoState(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to SMTP config and reconciles alert watchers.
func (r *arraySmtpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state arraySmtpModel
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

	// Patch SMTP relay config if changed.
	smtpPatch := client.SmtpServerPatch{}
	if !plan.RelayHost.Equal(state.RelayHost) {
		v := plan.RelayHost.ValueString()
		smtpPatch.RelayHost = &v
	}
	if !plan.SenderDomain.Equal(state.SenderDomain) {
		v := plan.SenderDomain.ValueString()
		smtpPatch.SenderDomain = &v
	}
	if !plan.EncryptionMode.Equal(state.EncryptionMode) {
		v := plan.EncryptionMode.ValueString()
		smtpPatch.EncryptionMode = &v
	}
	if smtpPatch.RelayHost != nil || smtpPatch.SenderDomain != nil || smtpPatch.EncryptionMode != nil {
		_, err := r.client.PatchSmtpServer(ctx, smtpPatch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating SMTP server configuration", err.Error())
			return
		}
	}

	// Reconcile alert watchers.
	planWatchers := extractAlertWatchers(ctx, plan.AlertWatchers, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	stateWatchers := extractAlertWatchers(ctx, state.AlertWatchers, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build maps by email for efficient lookup.
	planMap := make(map[string]alertWatcherModel, len(planWatchers))
	for _, w := range planWatchers {
		planMap[w.Email.ValueString()] = w
	}
	stateMap := make(map[string]alertWatcherModel, len(stateWatchers))
	for _, w := range stateWatchers {
		stateMap[w.Email.ValueString()] = w
	}

	// Add new watchers.
	for email, w := range planMap {
		if _, exists := stateMap[email]; !exists {
			severity := w.MinimumNotificationSeverity.ValueString()
			_, err := r.client.PostAlertWatcher(ctx, email, client.AlertWatcherPost{
				MinimumNotificationSeverity: severity,
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating alert watcher",
					fmt.Sprintf("Failed to create alert watcher for %s: %s", email, err),
				)
				return
			}
			// AlertWatcherPost has no enabled field; the API defaults to enabled=true.
			// PATCH to apply the plan's enabled=false.
			if !w.Enabled.IsNull() && !w.Enabled.IsUnknown() && !w.Enabled.ValueBool() {
				disabled := false
				if _, err := r.client.PatchAlertWatcher(ctx, email, client.AlertWatcherPatch{Enabled: &disabled}); err != nil {
					resp.Diagnostics.AddError(
						"Error disabling alert watcher",
						fmt.Sprintf("Failed to set enabled=false on new alert watcher %s: %s", email, err),
					)
					return
				}
			}
		}
	}

	// Remove deleted watchers.
	for email := range stateMap {
		if _, exists := planMap[email]; !exists {
			if err := r.client.DeleteAlertWatcher(ctx, email); err != nil {
				resp.Diagnostics.AddError(
					"Error deleting alert watcher",
					fmt.Sprintf("Failed to delete alert watcher for %s: %s", email, err),
				)
				return
			}
		}
	}

	// Update changed watchers (enabled or severity changed).
	for email, pw := range planMap {
		sw, exists := stateMap[email]
		if !exists {
			continue // already created above
		}
		watcherPatch := client.AlertWatcherPatch{}
		if !pw.Enabled.Equal(sw.Enabled) {
			v := pw.Enabled.ValueBool()
			watcherPatch.Enabled = &v
		}
		if !pw.MinimumNotificationSeverity.Equal(sw.MinimumNotificationSeverity) {
			v := pw.MinimumNotificationSeverity.ValueString()
			watcherPatch.MinimumNotificationSeverity = &v
		}
		if watcherPatch.Enabled != nil || watcherPatch.MinimumNotificationSeverity != nil {
			_, err := r.client.PatchAlertWatcher(ctx, email, watcherPatch)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating alert watcher",
					fmt.Sprintf("Failed to update alert watcher for %s: %s", email, err),
				)
				return
			}
		}
	}

	resp.Diagnostics.Append(r.readIntoState(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resets SMTP config and removes all alert watchers.
func (r *arraySmtpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data arraySmtpModel
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

	// Reset SMTP config to defaults.
	emptyStr := ""
	noneMode := "none"
	_, err := r.client.PatchSmtpServer(ctx, client.SmtpServerPatch{
		RelayHost:      &emptyStr,
		SenderDomain:   &emptyStr,
		EncryptionMode: &noneMode,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error resetting SMTP server configuration", err.Error())
		return
	}

	// Delete all alert watchers from state.
	watchers := extractAlertWatchers(ctx, data.AlertWatchers, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, w := range watchers {
		if err := r.client.DeleteAlertWatcher(ctx, w.Email.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error deleting alert watcher",
				fmt.Sprintf("Failed to delete alert watcher for %s: %s", w.Email.ValueString(), err),
			)
			return
		}
	}

	tflog.Info(ctx, "SMTP config and alert watchers cleared")
}

// ImportState imports the singleton SMTP config using "default" as the import ID.
func (r *arraySmtpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var data arraySmtpModel
	data.Timeouts = nullTimeoutsValue()

	resp.Diagnostics.Append(r.readIntoState(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ---------- helpers ---------------------------------------------------------

// readIntoState fetches SMTP config + alert watchers from the API and populates the model.
func (r *arraySmtpResource) readIntoState(ctx context.Context, data *arraySmtpModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smtp, err := r.client.GetSmtpServer(ctx)
	if err != nil {
		diags.AddError("Error reading SMTP server configuration", err.Error())
		return diags
	}

	watchers, err := r.client.ListAlertWatchers(ctx)
	if err != nil {
		diags.AddError("Error reading alert watchers", err.Error())
		return diags
	}

	data.ID = types.StringValue(smtp.ID)
	data.RelayHost = types.StringValue(smtp.RelayHost)
	data.SenderDomain = types.StringValue(smtp.SenderDomain)
	data.EncryptionMode = types.StringValue(smtp.EncryptionMode)

	// Map alert watchers to the set.
	watcherObjs := make([]attr.Value, 0, len(watchers))
	for _, w := range watchers {
		obj, d := types.ObjectValue(alertWatcherAttrTypes(), map[string]attr.Value{
			"email":                         types.StringValue(w.Name),
			"enabled":                       types.BoolValue(w.Enabled),
			"minimum_notification_severity": types.StringValue(w.MinimumNotificationSeverity),
		})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		watcherObjs = append(watcherObjs, obj)
	}

	watcherSet, d := types.SetValue(types.ObjectType{AttrTypes: alertWatcherAttrTypes()}, watcherObjs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.AlertWatchers = watcherSet

	return diags
}

// extractAlertWatchers deserializes the alert_watchers set from the model.
func extractAlertWatchers(ctx context.Context, set types.Set, diags interface {
	Append(...diag.Diagnostic)
	HasError() bool
}) []alertWatcherModel {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var watchers []alertWatcherModel
	d := set.ElementsAs(ctx, &watchers, false)
	diags.Append(d...)
	return watchers
}

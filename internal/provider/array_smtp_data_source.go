package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ datasource.DataSource = &arraySmtpDataSource{}
var _ datasource.DataSourceWithConfigure = &arraySmtpDataSource{}

// arraySmtpDataSource implements the flashblade_array_smtp data source.
type arraySmtpDataSource struct {
	client *client.FlashBladeClient
}

func NewArraySmtpDataSource() datasource.DataSource {
	return &arraySmtpDataSource{}
}

// ---------- model structs ----------------------------------------------------

// alertWatcherDataSourceModel represents a single alert watcher in the data source.
type alertWatcherDataSourceModel struct {
	Email                       types.String `tfsdk:"email"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	MinimumNotificationSeverity types.String `tfsdk:"minimum_notification_severity"`
}

// arraySmtpDataSourceModel is the top-level model for the flashblade_array_smtp data source.
type arraySmtpDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	RelayHost      types.String `tfsdk:"relay_host"`
	SenderDomain   types.String `tfsdk:"sender_domain"`
	EncryptionMode types.String `tfsdk:"encryption_mode"`
	AlertWatchers  types.Set    `tfsdk:"alert_watchers"`
}

// ---------- data source interface methods -----------------------------------

func (d *arraySmtpDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "flashblade_array_smtp"
}

// Schema defines the data source schema.
func (d *arraySmtpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the current SMTP relay configuration and alert watchers of a FlashBlade array.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SMTP server configuration.",
			},
			"relay_host": schema.StringAttribute{
				Computed:    true,
				Description: "Hostname or IP address of the SMTP relay server.",
			},
			"sender_domain": schema.StringAttribute{
				Computed:    true,
				Description: "Domain appended to the sender email address.",
			},
			"encryption_mode": schema.StringAttribute{
				Computed:    true,
				Description: "SMTP encryption mode: 'none', 'tls', or 'starttls'.",
			},
			"alert_watchers": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Set of alert watcher email recipients.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "Email address of the alert recipient.",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "If true, this watcher receives alert notifications.",
						},
						"minimum_notification_severity": schema.StringAttribute{
							Computed:    true,
							Description: "Minimum alert severity that triggers a notification.",
						},
					},
				},
			},
		},
	}
}

// Configure injects the FlashBladeClient into the data source.
func (d *arraySmtpDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = c
}

// Read fetches the SMTP configuration and alert watchers, then populates state.
func (d *arraySmtpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config arraySmtpDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smtp, err := d.client.GetSmtpServer(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SMTP server configuration", err.Error())
		return
	}

	watchers, err := d.client.ListAlertWatchers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert watchers", err.Error())
		return
	}

	config.ID = types.StringValue(smtp.ID)
	config.RelayHost = types.StringValue(smtp.RelayHost)
	config.SenderDomain = types.StringValue(smtp.SenderDomain)
	config.EncryptionMode = types.StringValue(smtp.EncryptionMode)

	watcherAttrTypes := alertWatcherAttrTypes()
	watcherObjs := make([]attr.Value, 0, len(watchers))
	for _, w := range watchers {
		obj, diags := types.ObjectValue(watcherAttrTypes, map[string]attr.Value{
			"email":                         types.StringValue(w.Name),
			"enabled":                       types.BoolValue(w.Enabled),
			"minimum_notification_severity": types.StringValue(w.MinimumNotificationSeverity),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		watcherObjs = append(watcherObjs, obj)
	}

	watcherSet, diags := types.SetValue(types.ObjectType{AttrTypes: watcherAttrTypes}, watcherObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.AlertWatchers = watcherSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

var _ provider.Provider = &FlashBladeProvider{}

// FlashBladeProvider is the root provider struct.
type FlashBladeProvider struct {
	version string
}

// flashBladeProviderModel maps the provider schema to Go types.
type flashBladeProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	CACertFile         types.String `tfsdk:"ca_cert_file"`
	CACert             types.String `tfsdk:"ca_cert"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	MaxRetries         types.Int64  `tfsdk:"max_retries"`
	Auth               *authModel   `tfsdk:"auth"`
}

// authModel maps the nested auth block.
type authModel struct {
	APIToken types.String `tfsdk:"api_token"`
	OAuth2   *oauth2Model `tfsdk:"oauth2"`
}

// oauth2Model maps the nested oauth2 sub-block.
type oauth2Model struct {
	ClientID types.String `tfsdk:"client_id"`
	KeyID    types.String `tfsdk:"key_id"`
	Issuer   types.String `tfsdk:"issuer"`
}

// New returns a factory function that creates a FlashBladeProvider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FlashBladeProvider{version: version}
	}
}

// Metadata sets the provider type name.
func (p *FlashBladeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "flashblade"
	resp.Version = p.version
}

// Schema returns the provider configuration schema.
func (p *FlashBladeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Pure Storage FlashBlade arrays via the REST API v2.22.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "FlashBlade management endpoint URL (e.g. https://flashblade.example.com). Falls back to FLASHBLADE_HOST environment variable.",
			},
			"ca_cert_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to a PEM-encoded CA certificate file used for TLS verification.",
			},
			"ca_cert": schema.StringAttribute{
				Optional:    true,
				Description: "Inline PEM-encoded CA certificate string used for TLS verification.",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLS certificate verification. For testing and development only.",
			},
			"max_retries": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of retry attempts for transient errors (429, 5xx). Default: 3.",
			},
			"auth": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Authentication configuration for the FlashBlade array.",
				Attributes: map[string]schema.Attribute{
					"api_token": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "API token for session-based authentication. Falls back to FLASHBLADE_API_TOKEN environment variable.",
					},
					"oauth2": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "OAuth2 token-exchange authentication configuration.",
						Attributes: map[string]schema.Attribute{
							"client_id": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "OAuth2 client ID. Falls back to FLASHBLADE_OAUTH2_CLIENT_ID environment variable.",
							},
							"key_id": schema.StringAttribute{
								Optional:    true,
								Sensitive:   true,
								Description: "OAuth2 key ID. Falls back to FLASHBLADE_OAUTH2_KEY_ID environment variable.",
							},
							"issuer": schema.StringAttribute{
								Optional:    true,
								Description: "OAuth2 issuer. Falls back to FLASHBLADE_OAUTH2_ISSUER environment variable.",
							},
						},
					},
				},
			},
		},
	}
}

// Configure initializes the FlashBladeClient from the provider configuration
// and injects it into ResourceData and DataSourceData for resource use.
func (p *FlashBladeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config flashBladeProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve endpoint: config value takes precedence, fall back to env var.
	endpoint := config.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("FLASHBLADE_HOST")
	}
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing FlashBlade Endpoint",
			"The provider requires an endpoint URL. Set the 'endpoint' attribute or the FLASHBLADE_HOST environment variable.",
		)
		return
	}

	// Resolve auth credentials.
	var apiToken, oauth2ClientID, oauth2KeyID, oauth2Issuer string
	if config.Auth != nil {
		apiToken = config.Auth.APIToken.ValueString()
		if config.Auth.OAuth2 != nil {
			oauth2ClientID = config.Auth.OAuth2.ClientID.ValueString()
			oauth2KeyID = config.Auth.OAuth2.KeyID.ValueString()
			oauth2Issuer = config.Auth.OAuth2.Issuer.ValueString()
		}
	}

	// Apply env var fallbacks for all auth fields.
	if apiToken == "" {
		apiToken = os.Getenv("FLASHBLADE_API_TOKEN")
	}
	if oauth2ClientID == "" {
		oauth2ClientID = os.Getenv("FLASHBLADE_OAUTH2_CLIENT_ID")
	}
	if oauth2KeyID == "" {
		oauth2KeyID = os.Getenv("FLASHBLADE_OAUTH2_KEY_ID")
	}
	if oauth2Issuer == "" {
		oauth2Issuer = os.Getenv("FLASHBLADE_OAUTH2_ISSUER")
	}

	// Validate that at least one auth method is configured.
	hasAPIToken := apiToken != ""
	hasOAuth2 := oauth2ClientID != "" || oauth2KeyID != ""
	if !hasAPIToken && !hasOAuth2 {
		resp.Diagnostics.AddError(
			"Missing FlashBlade Authentication",
			"The provider requires either an 'api_token' or an 'oauth2' block (client_id, key_id). "+
				"Set the credentials in the provider configuration or via FLASHBLADE_API_TOKEN / FLASHBLADE_OAUTH2_* environment variables.",
		)
		return
	}

	// Determine auth mode for logging.
	authMode := "token"
	if !hasAPIToken && hasOAuth2 {
		authMode = "oauth2"
	}

	// Log provider configuration (endpoint is logged, credentials are never logged).
	tflog.Info(ctx, "Configuring FlashBlade provider", map[string]any{
		"endpoint":  endpoint,
		"auth_mode": authMode,
	})

	// Resolve TLS config.
	caCertFile := config.CACertFile.ValueString()
	caCert := config.CACert.ValueString()
	insecureSkipVerify := config.InsecureSkipVerify.ValueBool()
	if insecureSkipVerify {
		tflog.Warn(ctx, "TLS certificate verification is disabled (insecure_skip_verify = true). This is unsafe for production use.")
	}

	// Resolve retry config.
	maxRetries := int(config.MaxRetries.ValueInt64())
	if maxRetries <= 0 {
		maxRetries = 3
	}

	// Build the client config.
	cfg := client.Config{
		Endpoint:           endpoint,
		APIToken:           apiToken,
		OAuth2ClientID:     oauth2ClientID,
		OAuth2KeyID:        oauth2KeyID,
		OAuth2Issuer:       oauth2Issuer,
		MaxRetries:         maxRetries,
		CACertFile:         caCertFile,
		CACert:             caCert,
		InsecureSkipVerify: insecureSkipVerify,
	}

	c, err := client.NewClient(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create FlashBlade client",
			fmt.Sprintf("Error initializing FlashBlade client for endpoint %q: %s", endpoint, err),
		)
		return
	}

	// Negotiate API version — fail early if v2.22 is not supported.
	if err := c.NegotiateVersion(ctx); err != nil {
		resp.Diagnostics.AddError(
			"FlashBlade API version not supported",
			fmt.Sprintf("FlashBlade at %s does not support API version v%s. Error: %s", endpoint, client.APIVersion, err),
		)
		return
	}

	tflog.Info(ctx, "FlashBlade provider configured successfully", map[string]any{
		"endpoint":  endpoint,
		"auth_mode": authMode,
	})

	// Inject the client into ResourceData and DataSourceData.
	resp.ResourceData = c
	resp.DataSourceData = c
}

// Resources returns the list of resource types provided by this provider.
func (p *FlashBladeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Storage — file systems
		NewFilesystemResource,

		// Storage — object store
		NewObjectStoreAccountResource,
		NewObjectStoreAccountExportResource,
		NewObjectStoreVirtualHostResource,
		NewObjectStoreUserResource,
		NewObjectStoreUserPolicyResource,
		NewAccessKeyResource,
		NewBucketResource,
		NewBucketAccessPolicyResource,
		NewBucketAccessPolicyRuleResource,
		NewBucketAuditFilterResource,
		NewLifecycleRuleResource,
		NewQosPolicyResource,
		NewQosPolicyMemberResource,

		// Policies — NFS, SMB, Snapshot, Network, S3
		NewNfsExportPolicyResource,
		NewNfsExportPolicyRuleResource,
		NewSmbSharePolicyResource,
		NewSmbSharePolicyRuleResource,
		NewSmbClientPolicyResource,
		NewSmbClientPolicyRuleResource,
		NewSnapshotPolicyResource,
		NewSnapshotPolicyRuleResource,
		NewNetworkAccessPolicyResource,
		NewNetworkAccessPolicyRuleResource,
		NewObjectStoreAccessPolicyResource,
		NewObjectStoreAccessPolicyRuleResource,
		NewS3ExportPolicyResource,
		NewS3ExportPolicyRuleResource,

		// Servers & exports
		NewServerResource,
		NewFileSystemExportResource,

		// Networking
		NewSubnetResource,
		NewNetworkInterfaceResource,

		// Replication
		NewRemoteCredentialsResource,
		NewBucketReplicaLinkResource,
		NewTargetResource,
		NewArrayConnectionResource,
		NewArrayConnectionKeyResource,

		// Security & TLS
		NewCertificateResource,
		NewTlsPolicyResource,
		NewTlsPolicyMemberResource,
		NewCertificateGroupResource,
		NewCertificateGroupMemberResource,

		// Quotas
		NewQuotaUserResource,
		NewQuotaGroupResource,

		// Array administration
		NewArrayDnsResource,
		NewArrayNtpResource,
		NewArraySmtpResource,
		NewSyslogServerResource,
		NewDirectoryServiceManagementResource,
		NewDirectoryServiceRoleResource,
		NewManagementAccessPolicyDirectoryServiceRoleMembershipResource,

		// Audit
		NewAuditObjectStorePolicyResource,
		NewAuditObjectStorePolicyMemberResource,
		NewLogTargetObjectStoreResource,
	}
}

// DataSources returns the list of data source types provided by this provider.
func (p *FlashBladeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Storage — file systems
		NewFilesystemDataSource,

		// Storage — object store
		NewObjectStoreAccountDataSource,
		NewObjectStoreAccountExportDataSource,
		NewObjectStoreVirtualHostDataSource,
		NewObjectStoreUserDataSource,
		NewAccessKeyDataSource,
		NewBucketDataSource,
		NewBucketAccessPolicyDataSource,
		NewBucketAuditFilterDataSource,
		NewLifecycleRuleDataSource,
		NewQosPolicyDataSource,

		// Policies — NFS, SMB, Snapshot, Network, S3
		NewNfsExportPolicyDataSource,
		NewSmbSharePolicyDataSource,
		NewSmbClientPolicyDataSource,
		NewSnapshotPolicyDataSource,
		NewNetworkAccessPolicyDataSource,
		NewObjectStoreAccessPolicyDataSource,
		NewS3ExportPolicyDataSource,

		// Servers & exports
		NewServerDataSource,
		NewFileSystemExportDataSource,

		// Networking
		NewSubnetDataSource,
		NewLinkAggregationGroupDataSource,
		NewNetworkInterfaceDataSource,

		// Replication
		NewArrayConnectionDataSource,
		NewRemoteCredentialsDataSource,
		NewBucketReplicaLinkDataSource,
		NewTargetDataSource,

		// Security & TLS
		NewCertificateDataSource,
		NewTlsPolicyDataSource,
		NewCertificateGroupDataSource,

		// Quotas
		NewQuotaUserDataSource,
		NewQuotaGroupDataSource,

		// Array administration
		NewArrayDnsDataSource,
		NewArrayNtpDataSource,
		NewArraySmtpDataSource,
		NewSyslogServerDataSource,
		NewDirectoryServiceManagementDataSource,
		NewDirectoryServiceRoleDataSource,

		// Audit
		NewAuditObjectStorePolicyDataSource,
		NewLogTargetObjectStoreDataSource,
	}
}

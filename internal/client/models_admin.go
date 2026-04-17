package client

// ArrayDns represents a FlashBlade DNS configuration from GET /dns.
type ArrayDns struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Domain      string   `json:"domain,omitempty"`
	Nameservers []string `json:"nameservers,omitempty"`
	Services    []string `json:"services,omitempty"`
	Sources     []string `json:"sources,omitempty"`
}

// ArrayDnsPost contains the fields accepted on POST /dns.
// Name is passed via ?names= query parameter, excluded from body.
type ArrayDnsPost struct {
	Name        string   `json:"-"`
	Domain      string   `json:"domain,omitempty"`
	Nameservers []string `json:"nameservers,omitempty"`
	Services    []string `json:"services,omitempty"`
	Sources     []string `json:"sources,omitempty"`
}

// ArrayDnsPatch contains pointer fields for PATCH /dns.
type ArrayDnsPatch struct {
	Domain      *string   `json:"domain,omitempty"`
	Nameservers *[]string `json:"nameservers,omitempty"`
	Services    *[]string `json:"services,omitempty"`
	Sources     *[]string `json:"sources,omitempty"`
}

// ArrayInfo represents the NTP-relevant fields from GET /arrays.
type ArrayInfo struct {
	ID         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	NtpServers []string `json:"ntp_servers,omitempty"`
}

// ArrayNtpPatch contains only the NTP servers field for PATCH /arrays.
// Only ntp_servers is sent to avoid unintentional modification of other array settings.
type ArrayNtpPatch struct {
	NtpServers *[]string `json:"ntp_servers,omitempty"`
}

// SmtpServer represents a FlashBlade SMTP server configuration from GET /smtp-servers.
type SmtpServer struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	RelayHost      string `json:"relay_host,omitempty"`
	SenderDomain   string `json:"sender_domain,omitempty"`
	EncryptionMode string `json:"encryption_mode,omitempty"`
}

// SmtpServerPatch contains pointer fields for PATCH /smtp-servers.
type SmtpServerPatch struct {
	RelayHost      *string `json:"relay_host,omitempty"`
	SenderDomain   *string `json:"sender_domain,omitempty"`
	EncryptionMode *string `json:"encryption_mode,omitempty"`
}

// AlertWatcher represents a FlashBlade alert watcher (email recipient) from GET /alert-watchers.
// Name is the email address.
type AlertWatcher struct {
	ID                          string `json:"id,omitempty"`
	Name                        string `json:"name"`
	Enabled                     bool   `json:"enabled"`
	MinimumNotificationSeverity string `json:"minimum_notification_severity,omitempty"`
}

// AlertWatcherPost contains the fields accepted on POST /alert-watchers.
// Name (email) is passed as a query parameter (?names=).
type AlertWatcherPost struct {
	MinimumNotificationSeverity string `json:"minimum_notification_severity,omitempty"`
}

// AlertWatcherPatch contains pointer fields for PATCH /alert-watchers.
type AlertWatcherPatch struct {
	Enabled                     *bool   `json:"enabled,omitempty"`
	MinimumNotificationSeverity *string `json:"minimum_notification_severity,omitempty"`
}

// SyslogServer represents a FlashBlade syslog server from GET responses.
type SyslogServer struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	URI      string   `json:"uri,omitempty"`
	Services []string `json:"services,omitempty"`
	Sources  []string `json:"sources,omitempty"`
}

// SyslogServerPost contains the fields accepted on POST /syslog-servers.
type SyslogServerPost struct {
	URI      string   `json:"uri,omitempty"`
	Services []string `json:"services,omitempty"`
	Sources  []string `json:"sources,omitempty"`
}

// SyslogServerPatch contains pointer fields for PATCH /syslog-servers.
type SyslogServerPatch struct {
	URI      *string   `json:"uri,omitempty"`
	Services *[]string `json:"services,omitempty"`
	Sources  *[]string `json:"sources,omitempty"`
}

// ArrayConnection represents a FlashBlade array connection from GET /array-connections.
type ArrayConnection struct {
	ID                   string                  `json:"id"`
	Status               string                  `json:"status,omitempty"`
	Remote               NamedReference          `json:"remote"`
	ManagementAddress    string                  `json:"management_address,omitempty"`
	ReplicationAddresses []string                `json:"replication_addresses,omitempty"`
	Encrypted            bool                    `json:"encrypted"`
	Type                 string                  `json:"type,omitempty"`
	Version              string                  `json:"version,omitempty"`
	OS                   string                  `json:"os,omitempty"`
	CACertificateGroup   *NamedReference         `json:"ca_certificate_group,omitempty"`
	Throttle             *ArrayConnectionThrottle `json:"throttle,omitempty"`
}

// ArrayConnectionThrottle configures bandwidth throttling for an array connection.
type ArrayConnectionThrottle struct {
	DefaultLimit *int64  `json:"default_limit,omitempty"`
	WindowLimit  *int64  `json:"window_limit,omitempty"`
	WindowStart  *string `json:"window_start,omitempty"`
	WindowEnd    *string `json:"window_end,omitempty"`
}

// ArrayConnectionPost contains the fields for POST /array-connections.
// The remote name is passed via ?remote_names= query parameter, not in the body.
type ArrayConnectionPost struct {
	ManagementAddress    string                   `json:"management_address"`
	ConnectionKey        string                   `json:"connection_key"`
	Encrypted            bool                     `json:"encrypted"`
	ReplicationAddresses []string                 `json:"replication_addresses,omitempty"`
	Throttle             *ArrayConnectionThrottle `json:"throttle,omitempty"`
	Remote               *NamedReference          `json:"remote,omitempty"`
	// ca_certificate_group is NOT accepted on POST — must be set via PATCH after creation.
}

// ArrayConnectionPatch contains pointer fields for PATCH /array-connections.
// Nil outer pointer means omit the field. Non-nil outer + nil inner = set to null (for **NamedReference).
// connection_key is absent — write-only on POST only.
type ArrayConnectionPatch struct {
	ManagementAddress    *string                  `json:"management_address,omitempty"`
	Encrypted            *bool                    `json:"encrypted,omitempty"`
	ReplicationAddresses *[]string                `json:"replication_addresses,omitempty"`
	Throttle             *ArrayConnectionThrottle `json:"throttle,omitempty"`
}

// ArrayConnectionKey represents the response from GET/POST /array-connections/connection-key.
// There is only one connection key per array at a time. All fields are read-only.
type ArrayConnectionKey struct {
	ConnectionKey string `json:"connection_key"`
	Created       int64  `json:"created"`
	Expires       int64  `json:"expires"`
}

// DirectoryServiceManagement represents the nested management sub-object in a DirectoryService.
// Holds LDAP attribute names used when authenticating FlashBlade admin users.
type DirectoryServiceManagement struct {
	UserLoginAttribute    string `json:"user_login_attribute,omitempty"`
	UserObjectClass       string `json:"user_object_class,omitempty"`
	SSHPublicKeyAttribute string `json:"ssh_public_key_attribute,omitempty"`
}

// DirectoryServiceManagementPatch contains pointer fields for the management sub-object of DirectoryServicePatch.
// Nil field = omit; non-nil = send (empty string clears on the array).
type DirectoryServiceManagementPatch struct {
	UserLoginAttribute    *string `json:"user_login_attribute,omitempty"`
	UserObjectClass       *string `json:"user_object_class,omitempty"`
	SSHPublicKeyAttribute *string `json:"ssh_public_key_attribute,omitempty"`
}

// DirectoryService represents a FlashBlade directory service configuration from GET /directory-services.
// The management singleton is identified by Name == "management".
// NOTE: the `smb` sub-object is DEPRECATED in v2.22 and intentionally NOT modelled here.
type DirectoryService struct {
	ID                 string                     `json:"id,omitempty"`
	Name               string                     `json:"name,omitempty"`
	Enabled            bool                       `json:"enabled"`
	URIs               []string                   `json:"uris,omitempty"`
	BaseDN             string                     `json:"base_dn,omitempty"`
	BindUser           string                     `json:"bind_user,omitempty"`
	CACertificate      *NamedReference            `json:"ca_certificate,omitempty"`
	CACertificateGroup *NamedReference            `json:"ca_certificate_group,omitempty"`
	Management         DirectoryServiceManagement `json:"management"`
	Services           []string                   `json:"services,omitempty"`
}

// DirectoryServicePatch contains pointer fields for PATCH /directory-services.
// **NamedReference semantics: nil outer = omit; non-nil outer + nil inner = set null;
// non-nil outer + non-nil inner = set to reference.
// bind_password is write-only — API never returns it; only sent when operator supplies it.
type DirectoryServicePatch struct {
	Enabled            *bool                            `json:"enabled,omitempty"`
	URIs               *[]string                        `json:"uris,omitempty"`
	BaseDN             *string                          `json:"base_dn,omitempty"`
	BindUser           *string                          `json:"bind_user,omitempty"`
	BindPassword       *string                          `json:"bind_password,omitempty"`
	CACertificate      **NamedReference                 `json:"ca_certificate,omitempty"`
	CACertificateGroup **NamedReference                 `json:"ca_certificate_group,omitempty"`
	Management         *DirectoryServiceManagementPatch `json:"management,omitempty"`
}

// DirectoryServiceRole represents a FlashBlade directory-service role mapping
// from GET /directory-services/roles. Reference: swagger schema DirectoryServiceRole.
// NOTE: management_access_policies is readonly per swagger on PATCH; role is deprecated
// but remains for backwards compatibility (surface it computed-only on the resource).
type DirectoryServiceRole struct {
	ID                       string           `json:"id"`
	Name                     string           `json:"name"`
	Group                    string           `json:"group"`
	GroupBase                string           `json:"group_base"`
	ManagementAccessPolicies []NamedReference `json:"management_access_policies"`
	Role                     *NamedReference  `json:"role,omitempty"`
}

// DirectoryServiceRolePost is the POST /directory-services/roles body.
// Name is user-supplied via the ?names= query param (see PostDirectoryServiceRole).
// management_access_policies is writable on POST only (readonly on PATCH).
type DirectoryServiceRolePost struct {
	Group                    string           `json:"group"`
	GroupBase                string           `json:"group_base"`
	ManagementAccessPolicies []NamedReference `json:"management_access_policies"`
	// Role is the deprecated field — omit unless explicitly set.
	Role *NamedReference `json:"role,omitempty"`
}

// DirectoryServiceRolePatch is the PATCH /directory-services/roles?names=<n> body.
// Every field is a pointer so nil = omit. management_access_policies is NOT present
// (readonly on PATCH per api_references/2.22.md line 434).
type DirectoryServiceRolePatch struct {
	Group     *string          `json:"group,omitempty"`
	GroupBase *string          `json:"group_base,omitempty"`
	Role      **NamedReference `json:"role,omitempty"`
}

// ManagementAccessPolicyDirectoryServiceRoleMembership represents a single
// association from GET /management-access-policies/directory-services/roles.
// The endpoint returns items keyed by (policy, role) pairs — no standalone identity.
type ManagementAccessPolicyDirectoryServiceRoleMembership struct {
	Policy NamedReference `json:"policy"`
	Role   NamedReference `json:"role"`
}

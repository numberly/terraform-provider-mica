package client

import "encoding/json"

// Space holds the storage space breakdown for a file system.
type Space struct {
	DataReduction       float64 `json:"data_reduction,omitempty"`
	Snapshots           int64   `json:"snapshots,omitempty"`
	TotalPhysical       int64   `json:"total_physical,omitempty"`
	Unique              int64   `json:"unique,omitempty"`
	Virtual             int64   `json:"virtual,omitempty"`
	SnapshotsEffective  int64   `json:"snapshots_effective,omitempty"`
}

// NFSConfig represents the NFS protocol configuration on a file system.
type NFSConfig struct {
	Enabled     bool   `json:"enabled"`
	V3Enabled   bool   `json:"v3_enabled"`
	V41Enabled  bool   `json:"v4_1_enabled"`
	Rules       string `json:"rules,omitempty"`
	Transport   string `json:"transport,omitempty"`
}

// SMBConfig represents the SMB protocol configuration on a file system.
type SMBConfig struct {
	Enabled                        bool `json:"enabled"`
	AccessBasedEnumerationEnabled  bool `json:"access_based_enumeration_enabled,omitempty"`
	ContinuousAvailabilityEnabled  bool `json:"continuous_availability_enabled,omitempty"`
	SMBEncryptionEnabled           bool `json:"smb_encryption_enabled,omitempty"`
}

// HTTPConfig represents the HTTP protocol configuration on a file system.
type HTTPConfig struct {
	Enabled bool `json:"enabled"`
}

// DefaultQuotas holds default quota configuration for a file system.
type DefaultQuotas struct {
	GroupQuota int64 `json:"group_quota,omitempty"`
	UserQuota  int64 `json:"user_quota,omitempty"`
}

// MultiProtocol holds multi-protocol configuration for a file system.
type MultiProtocol struct {
	AccessControlStyle string `json:"access_control_style,omitempty"`
	SafeguardACLsOnDestroy bool `json:"safeguard_acls_on_destroy,omitempty"`
}

// SourceReference refers to a source file system (for snapshots/clones).
type SourceReference struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// FileSystem represents a FlashBlade file system object from GET responses.
type FileSystem struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Provisioned      int64         `json:"provisioned"`
	Destroyed        bool          `json:"destroyed"`
	TimeRemaining    int64         `json:"time_remaining,omitempty"`
	Created          int64         `json:"created,omitempty"`
	Space            Space         `json:"space,omitempty"`
	NFS              NFSConfig     `json:"nfs,omitempty"`
	SMB              SMBConfig     `json:"smb,omitempty"`
	HTTP             HTTPConfig    `json:"http,omitempty"`
	DefaultQuotas    DefaultQuotas `json:"default_quotas,omitempty"`
	MultiProtocol    MultiProtocol `json:"multi_protocol,omitempty"`
	PromotionStatus  string        `json:"promotion_status,omitempty"`
	RequestNumber    int64         `json:"request_number,omitempty"`
	Source           *SourceReference `json:"source,omitempty"`
	Writable         bool          `json:"writable"`
}

// FileSystemPost contains the fields accepted on POST /api/2.22/file-systems.
type FileSystemPost struct {
	Name        string    `json:"name"`
	Provisioned int64     `json:"provisioned,omitempty"`
	NFS         NFSConfig `json:"nfs,omitempty"`
	SMB         SMBConfig `json:"smb,omitempty"`
	Writable    bool      `json:"writable,omitempty"`
}

// FileSystemPatch contains all mutable fields for PATCH /api/2.22/file-systems.
// Pointer types allow distinguishing "absent" from "false".
type FileSystemPatch struct {
	Name        *string    `json:"name,omitempty"`
	Provisioned *int64     `json:"provisioned,omitempty"`
	Destroyed   *bool      `json:"destroyed,omitempty"`
	Writable    *bool      `json:"writable,omitempty"`
	NFS         *NFSConfig `json:"nfs,omitempty"`
	SMB         *SMBConfig `json:"smb,omitempty"`
}

// ListResponse is a generic paginated list wrapper for FlashBlade API list endpoints.
type ListResponse[T any] struct {
	Items             []T    `json:"items"`
	ContinuationToken string `json:"continuation_token,omitempty"`
	TotalItemCount    int    `json:"total_item_count,omitempty"`
}

// VersionResponse represents the GET /api/api_version response.
type VersionResponse struct {
	Versions []string `json:"versions"`
}

// ---------- Phase 2 model structs -------------------------------------------

// NamedReference is a lightweight reference to another object by name and ID.
type NamedReference struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// ObjectStoreAccount represents a FlashBlade object store account from GET responses.
type ObjectStoreAccount struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Created          int64  `json:"created,omitempty"`
	QuotaLimit       string `json:"quota_limit,omitempty"`
	HardLimitEnabled bool   `json:"hard_limit_enabled"`
	ObjectCount      int64  `json:"object_count,omitempty"`
	Space            Space  `json:"space,omitempty"`
}

// ObjectStoreAccountPost contains the mutable fields for POST /object-store-accounts.
// Name is passed as a query parameter (?names=), not in the body.
type ObjectStoreAccountPost struct {
	QuotaLimit       string `json:"quota_limit,omitempty"`
	HardLimitEnabled bool   `json:"hard_limit_enabled,omitempty"`
}

// ObjectStoreAccountPatch contains pointer fields for PATCH semantics.
// Only non-nil fields are included in the request body.
type ObjectStoreAccountPatch struct {
	QuotaLimit       *string `json:"quota_limit,omitempty"`
	HardLimitEnabled *bool   `json:"hard_limit_enabled,omitempty"`
}

// Bucket represents a FlashBlade object store bucket from GET responses.
type Bucket struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Account          NamedReference `json:"account,omitempty"`
	Created          int64          `json:"created,omitempty"`
	Destroyed        bool           `json:"destroyed"`
	TimeRemaining    int64          `json:"time_remaining,omitempty"`
	Versioning       string         `json:"versioning,omitempty"`
	QuotaLimit       string         `json:"quota_limit,omitempty"`
	HardLimitEnabled bool           `json:"hard_limit_enabled"`
	ObjectCount      int64          `json:"object_count,omitempty"`
	BucketType       string         `json:"bucket_type,omitempty"`
	RetentionLock    string         `json:"retention_lock,omitempty"`
	Space            Space          `json:"space,omitempty"`
}

// BucketPost contains the fields accepted on POST /buckets.
type BucketPost struct {
	Account          NamedReference `json:"account"`
	Versioning       string         `json:"versioning,omitempty"`
	QuotaLimit       string         `json:"quota_limit,omitempty"`
	HardLimitEnabled bool           `json:"hard_limit_enabled,omitempty"`
	RetentionLock    string         `json:"retention_lock,omitempty"`
}

// BucketPatch contains pointer fields for PATCH semantics on /buckets.
type BucketPatch struct {
	Destroyed        *bool   `json:"destroyed,omitempty"`
	Versioning       *string `json:"versioning,omitempty"`
	QuotaLimit       *string `json:"quota_limit,omitempty"`
	HardLimitEnabled *bool   `json:"hard_limit_enabled,omitempty"`
	RetentionLock    *string `json:"retention_lock,omitempty"`
}

// ObjectStoreAccessKey represents a FlashBlade object store access key.
type ObjectStoreAccessKey struct {
	Name            string         `json:"name"`
	AccessKeyID     string         `json:"access_key_id"`
	SecretAccessKey string         `json:"secret_access_key,omitempty"`
	Created         int64          `json:"created,omitempty"`
	Enabled         bool           `json:"enabled"`
	User            NamedReference `json:"user,omitempty"`
}

// ObjectStoreAccessKeyPost contains the fields for POST /object-store-access-keys.
type ObjectStoreAccessKeyPost struct {
	User NamedReference `json:"user"`
}

// ---------- Phase 3 model structs -------------------------------------------

// PolicyMember represents a file system that is a member of a policy.
// Used for delete-guard checks across all policy families.
type PolicyMember struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// NfsExportPolicy represents a FlashBlade NFS export policy from GET responses.
type NfsExportPolicy struct {
	ID         string                      `json:"id,omitempty"`
	Name       string                      `json:"name"`
	Enabled    bool                        `json:"enabled"`
	IsLocal    bool                        `json:"is_local,omitempty"`
	PolicyType string                      `json:"policy_type,omitempty"`
	Version    string                      `json:"version,omitempty"`
	Rules      []NfsExportPolicyRuleInPolicy `json:"rules,omitempty"`
}

// NfsExportPolicyPost contains the fields accepted on POST /nfs-export-policies.
type NfsExportPolicyPost struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// NfsExportPolicyPatch contains pointer fields for PATCH /nfs-export-policies.
type NfsExportPolicyPatch struct {
	Name    *string `json:"name,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

// NfsExportPolicyRule represents a rule from GET /nfs-export-policies/rules.
// Note: anongid/anonuid are integers in GET responses.
type NfsExportPolicyRule struct {
	ID                        string         `json:"id,omitempty"`
	Name                      string         `json:"name,omitempty"`
	Index                     int            `json:"index"`
	Policy                    NamedReference `json:"policy,omitempty"`
	PolicyVersion             string         `json:"policy_version,omitempty"`
	Access                    string         `json:"access,omitempty"`
	Client                    string         `json:"client,omitempty"`
	Permission                string         `json:"permission,omitempty"`
	Anonuid                   int            `json:"anonuid,omitempty"`
	Anongid                   int            `json:"anongid,omitempty"`
	Atime                     bool           `json:"atime"`
	Fileid32bit               bool           `json:"fileid_32bit"`
	Secure                    bool           `json:"secure"`
	Security                  []string       `json:"security,omitempty"`
	RequiredTransportSecurity string         `json:"required_transport_security,omitempty"`
}

// NfsExportPolicyRuleInPolicy is an NFS rule embedded inside a policy GET response.
type NfsExportPolicyRuleInPolicy struct {
	Index                     int      `json:"index"`
	Access                    string   `json:"access,omitempty"`
	Client                    string   `json:"client,omitempty"`
	Permission                string   `json:"permission,omitempty"`
	Anonuid                   int      `json:"anonuid,omitempty"`
	Anongid                   int      `json:"anongid,omitempty"`
	Atime                     bool     `json:"atime"`
	Fileid32bit               bool     `json:"fileid_32bit"`
	Secure                    bool     `json:"secure"`
	Security                  []string `json:"security,omitempty"`
	RequiredTransportSecurity string   `json:"required_transport_security,omitempty"`
}

// NfsExportPolicyRulePost contains the writable fields for POST /nfs-export-policies/rules.
type NfsExportPolicyRulePost struct {
	Access                    string         `json:"access,omitempty"`
	Client                    string         `json:"client,omitempty"`
	Permission                string         `json:"permission,omitempty"`
	Anonuid                   int            `json:"anonuid,omitempty"`
	Anongid                   int            `json:"anongid,omitempty"`
	Atime                     *bool          `json:"atime,omitempty"`
	Fileid32bit               *bool          `json:"fileid_32bit,omitempty"`
	Secure                    *bool          `json:"secure,omitempty"`
	Security                  []string       `json:"security,omitempty"`
	RequiredTransportSecurity string         `json:"required_transport_security,omitempty"`
	Policy                    *NamedReference `json:"policy,omitempty"`
}

// NfsExportPolicyRulePatch contains pointer fields for PATCH /nfs-export-policies/rules.
// Note: anonuid/anongid are strings in PATCH requests (API schema difference from GET).
type NfsExportPolicyRulePatch struct {
	Index                     *int     `json:"index,omitempty"`
	Access                    *string  `json:"access,omitempty"`
	Client                    *string  `json:"client,omitempty"`
	Permission                *string  `json:"permission,omitempty"`
	Anonuid                   *string  `json:"anonuid,omitempty"`
	Anongid                   *string  `json:"anongid,omitempty"`
	Atime                     *bool    `json:"atime,omitempty"`
	Fileid32bit               *bool    `json:"fileid_32bit,omitempty"`
	Secure                    *bool    `json:"secure,omitempty"`
	Security                  []string `json:"security,omitempty"`
	RequiredTransportSecurity *string  `json:"required_transport_security,omitempty"`
}

// SmbSharePolicy represents a FlashBlade SMB share policy from GET responses.
type SmbSharePolicy struct {
	ID         string                      `json:"id,omitempty"`
	Name       string                      `json:"name"`
	Enabled    bool                        `json:"enabled"`
	IsLocal    bool                        `json:"is_local,omitempty"`
	PolicyType string                      `json:"policy_type,omitempty"`
	Rules      []SmbSharePolicyRuleInPolicy `json:"rules,omitempty"`
}

// SmbSharePolicyPost contains the fields accepted on POST /smb-share-policies.
type SmbSharePolicyPost struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// SmbSharePolicyPatch contains pointer fields for PATCH /smb-share-policies.
type SmbSharePolicyPatch struct {
	Name    *string `json:"name,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

// SmbSharePolicyRule represents a rule from GET /smb-share-policies/rules.
type SmbSharePolicyRule struct {
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Policy      NamedReference `json:"policy,omitempty"`
	Principal   string         `json:"principal,omitempty"`
	Change      string         `json:"change,omitempty"`
	FullControl string         `json:"full_control,omitempty"`
	Read        string         `json:"read,omitempty"`
}

// SmbSharePolicyRuleInPolicy is an SMB rule embedded inside a policy GET response.
type SmbSharePolicyRuleInPolicy struct {
	Name        string `json:"name,omitempty"`
	Principal   string `json:"principal,omitempty"`
	Change      string `json:"change,omitempty"`
	FullControl string `json:"full_control,omitempty"`
	Read        string `json:"read,omitempty"`
}

// SmbSharePolicyRulePost contains the writable fields for POST /smb-share-policies/rules.
type SmbSharePolicyRulePost struct {
	Principal   string `json:"principal,omitempty"`
	Change      string `json:"change,omitempty"`
	FullControl string `json:"full_control,omitempty"`
	Read        string `json:"read,omitempty"`
}

// SmbSharePolicyRulePatch contains pointer fields for PATCH /smb-share-policies/rules.
type SmbSharePolicyRulePatch struct {
	Principal   *string `json:"principal,omitempty"`
	Change      *string `json:"change,omitempty"`
	FullControl *string `json:"full_control,omitempty"`
	Read        *string `json:"read,omitempty"`
}

// SnapshotPolicy represents a FlashBlade snapshot policy from GET /policies.
type SnapshotPolicy struct {
	ID            string                       `json:"id,omitempty"`
	Name          string                       `json:"name"`
	Enabled       bool                         `json:"enabled"`
	IsLocal       bool                         `json:"is_local,omitempty"`
	PolicyType    string                       `json:"policy_type,omitempty"`
	RetentionLock string                       `json:"retention_lock,omitempty"`
	Rules         []SnapshotPolicyRuleInPolicy  `json:"rules,omitempty"`
}

// SnapshotPolicyPost contains the fields accepted on POST /policies.
type SnapshotPolicyPost struct {
	Enabled *bool                    `json:"enabled,omitempty"`
	Rules   []SnapshotPolicyRulePost `json:"rules,omitempty"`
}

// SnapshotPolicyPatch contains the fields for PATCH /policies.
// Name is read-only and must NOT be sent. Rules are managed via add_rules/remove_rules.
type SnapshotPolicyPatch struct {
	Enabled     *bool                      `json:"enabled,omitempty"`
	AddRules    []SnapshotPolicyRulePost    `json:"add_rules,omitempty"`
	RemoveRules []SnapshotPolicyRuleRemove  `json:"remove_rules,omitempty"`
}

// SnapshotPolicyRuleInPolicy represents a rule embedded in a snapshot policy GET response.
type SnapshotPolicyRuleInPolicy struct {
	Name       string  `json:"name,omitempty"`
	AtTime     *int64  `json:"at,omitempty"`
	Every      *int64  `json:"every,omitempty"`
	KeepFor    *int64  `json:"keep_for,omitempty"`
	Suffix     string  `json:"suffix,omitempty"`
	ClientName string  `json:"client_name,omitempty"`
}

// SnapshotPolicyRulePost contains the fields for adding a rule via add_rules.
type SnapshotPolicyRulePost struct {
	AtTime     *int64 `json:"at,omitempty"`
	Every      *int64 `json:"every,omitempty"`
	KeepFor    *int64 `json:"keep_for,omitempty"`
	Suffix     string `json:"suffix,omitempty"`
	ClientName string `json:"client_name,omitempty"`
}

// SnapshotPolicyRuleRemove identifies a rule to remove via remove_rules.
type SnapshotPolicyRuleRemove struct {
	Name string `json:"name"`
}

// ---------- Phase 4 model structs -------------------------------------------

// ObjectStoreAccessPolicy represents a FlashBlade object store access policy from GET responses.
type ObjectStoreAccessPolicy struct {
	ID          string                        `json:"id,omitempty"`
	Name        string                        `json:"name"`
	Description string                        `json:"description,omitempty"`
	ARN         string                        `json:"arn,omitempty"`
	Enabled     bool                          `json:"enabled"`
	IsLocal     bool                          `json:"is_local,omitempty"`
	PolicyType  string                        `json:"policy_type,omitempty"`
	Rules       []ObjectStoreAccessPolicyRule `json:"rules,omitempty"`
	Created     int64                         `json:"created,omitempty"`
	Updated     int64                         `json:"updated,omitempty"`
}

// ObjectStoreAccessPolicyPost contains the fields accepted on POST /object-store-access-policies.
// Name is passed as a query parameter (?names=), not in the body.
// Rules are created separately via the /rules endpoint.
type ObjectStoreAccessPolicyPost struct {
	Description string `json:"description,omitempty"`
}

// ObjectStoreAccessPolicyPatch contains pointer fields for PATCH /object-store-access-policies.
// Only name is writable via PATCH; description is POST-only per API spec.
type ObjectStoreAccessPolicyPatch struct {
	Name *string `json:"name,omitempty"`
}

// ObjectStoreAccessPolicyRule represents a rule from GET /object-store-access-policies/rules.
type ObjectStoreAccessPolicyRule struct {
	Name       string             `json:"name,omitempty"`
	Effect     string             `json:"effect,omitempty"`
	Actions    []string           `json:"actions,omitempty"`
	Conditions json.RawMessage    `json:"conditions,omitempty"`
	Resources  []string           `json:"resources,omitempty"`
	Policy     *NamedReference    `json:"policy,omitempty"`
}

// ObjectStoreAccessPolicyRulePost contains the writable fields for POST /object-store-access-policies/rules.
type ObjectStoreAccessPolicyRulePost struct {
	Effect     string          `json:"effect"`
	Actions    []string        `json:"actions"`
	Conditions json.RawMessage `json:"conditions,omitempty"`
	Resources  []string        `json:"resources"`
}

// ObjectStoreAccessPolicyRulePatch contains pointer fields for PATCH /object-store-access-policies/rules.
// Effect is read-only after creation (RequiresReplace in TF schema).
type ObjectStoreAccessPolicyRulePatch struct {
	Actions    []string        `json:"actions,omitempty"`
	Conditions json.RawMessage `json:"conditions,omitempty"`
	Resources  []string        `json:"resources,omitempty"`
}

// NetworkAccessPolicy represents a FlashBlade network access policy from GET responses.
type NetworkAccessPolicy struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	IsLocal    bool   `json:"is_local,omitempty"`
	PolicyType string `json:"policy_type,omitempty"`
	Version    string `json:"version,omitempty"`
}

// NetworkAccessPolicyPatch contains pointer fields for PATCH /network-access-policies.
// No POST model exists — network access policies are singletons.
type NetworkAccessPolicyPatch struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Name    *string `json:"name,omitempty"`
}

// NetworkAccessPolicyRule represents a rule from GET /network-access-policies/rules.
type NetworkAccessPolicyRule struct {
	ID            string         `json:"id,omitempty"`
	Name          string         `json:"name,omitempty"`
	Client        string         `json:"client,omitempty"`
	Effect        string         `json:"effect,omitempty"`
	Index         int            `json:"index"`
	Interfaces    []string       `json:"interfaces,omitempty"`
	Policy        *NamedReference `json:"policy,omitempty"`
	PolicyVersion string         `json:"policy_version,omitempty"`
}

// NetworkAccessPolicyRulePost contains the writable fields for POST /network-access-policies/rules.
type NetworkAccessPolicyRulePost struct {
	Client     string   `json:"client"`
	Effect     string   `json:"effect"`
	Index      int      `json:"index,omitempty"`
	Interfaces []string `json:"interfaces"`
}

// NetworkAccessPolicyRulePatch contains pointer fields for PATCH /network-access-policies/rules.
type NetworkAccessPolicyRulePatch struct {
	Client     *string  `json:"client,omitempty"`
	Effect     *string  `json:"effect,omitempty"`
	Index      *int     `json:"index,omitempty"`
	Interfaces []string `json:"interfaces,omitempty"`
}

// QuotaUser represents a per-filesystem user quota from GET /quotas/users.
type QuotaUser struct {
	FileSystem *NamedReference `json:"file_system,omitempty"`
	User       *NamedReference `json:"user,omitempty"`
	Quota      int64           `json:"quota"`
	Usage      int64           `json:"usage,omitempty"`
}

// QuotaUserPost contains the writable fields for POST /quotas/users.
// file_system_names and uids are passed as query parameters.
type QuotaUserPost struct {
	Quota int64 `json:"quota"`
}

// QuotaUserPatch contains pointer fields for PATCH /quotas/users.
type QuotaUserPatch struct {
	Quota *int64 `json:"quota,omitempty"`
}

// QuotaGroup represents a per-filesystem group quota from GET /quotas/groups.
type QuotaGroup struct {
	FileSystem *NamedReference `json:"file_system,omitempty"`
	Group      *NamedReference `json:"group,omitempty"`
	Quota      int64           `json:"quota"`
	Usage      int64           `json:"usage,omitempty"`
}

// QuotaGroupPost contains the writable fields for POST /quotas/groups.
// file_system_names and gids are passed as query parameters.
type QuotaGroupPost struct {
	Quota int64 `json:"quota"`
}

// QuotaGroupPatch contains pointer fields for PATCH /quotas/groups.
type QuotaGroupPatch struct {
	Quota *int64 `json:"quota,omitempty"`
}

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
type ArrayDnsPost struct {
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

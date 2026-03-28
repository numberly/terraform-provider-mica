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
// Name is NOT serialised to JSON — the API requires it as a ?names= query parameter.
type FileSystemPost struct {
	Name        string     `json:"-"`
	Provisioned int64      `json:"provisioned,omitempty"`
	NFS         *NFSConfig `json:"nfs,omitempty"`
	SMB         *SMBConfig `json:"smb,omitempty"`
	Writable    bool       `json:"writable,omitempty"`
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
	QuotaLimit       int64 `json:"quota_limit,omitempty"`
	HardLimitEnabled bool   `json:"hard_limit_enabled"`
	ObjectCount      int64  `json:"object_count,omitempty"`
	Space            Space  `json:"space,omitempty"`
}

// ObjectStoreAccountPost contains the mutable fields for POST /object-store-accounts.
// Name is passed as a query parameter (?names=), not in the body.
// NOTE: quota_limit must be serialized as a string per FlashBlade API.
type ObjectStoreAccountPost struct {
	QuotaLimit       string `json:"quota_limit,omitempty"`
	HardLimitEnabled bool   `json:"hard_limit_enabled,omitempty"`
	// AccountExports controls default exports at creation time.
	// Set to non-nil empty slice to suppress the default _array_server export.
	// Leave nil to let FlashBlade create the default export.
	// Note: uses json.RawMessage wrapper to distinguish nil (omit) from [] (send empty array).
	AccountExports *json.RawMessage `json:"account_exports,omitempty"`
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
	QuotaLimit       int64          `json:"quota_limit,omitempty"`
	HardLimitEnabled bool           `json:"hard_limit_enabled"`
	ObjectCount      int64          `json:"object_count,omitempty"`
	BucketType       string         `json:"bucket_type,omitempty"`
	RetentionLock    string         `json:"retention_lock,omitempty"`
	Space            Space          `json:"space,omitempty"`
}

// BucketPost contains the fields accepted on POST /buckets.
// NOTE: quota_limit must be serialized as a string per FlashBlade API.
// NOTE: versioning is NOT a valid POST parameter — use PATCH after creation.
type BucketPost struct {
	Account          NamedReference `json:"account"`
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
	// Policy is passed as a query parameter (policy_names=), NOT in the JSON body.
	// The API rejects any "policy" field in the POST body (HTTP 400).
	Policy *NamedReference `json:"-"`
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

// SmbClientPolicy represents a FlashBlade SMB client policy from GET responses.
type SmbClientPolicy struct {
	ID                           string                        `json:"id,omitempty"`
	Name                         string                        `json:"name"`
	Enabled                      bool                          `json:"enabled"`
	IsLocal                      bool                          `json:"is_local,omitempty"`
	PolicyType                   string                        `json:"policy_type,omitempty"`
	Version                      string                        `json:"version,omitempty"`
	AccessBasedEnumerationEnabled bool                         `json:"access_based_enumeration_enabled,omitempty"`
	Rules                        []SmbClientPolicyRuleInPolicy `json:"rules,omitempty"`
}

// SmbClientPolicyPost contains the fields accepted on POST /smb-client-policies.
type SmbClientPolicyPost struct {
	Enabled                      *bool `json:"enabled,omitempty"`
	AccessBasedEnumerationEnabled *bool `json:"access_based_enumeration_enabled,omitempty"`
}

// SmbClientPolicyPatch contains pointer fields for PATCH /smb-client-policies.
// Version is read-only and must NOT be included.
type SmbClientPolicyPatch struct {
	Name                         *string `json:"name,omitempty"`
	Enabled                      *bool   `json:"enabled,omitempty"`
	AccessBasedEnumerationEnabled *bool   `json:"access_based_enumeration_enabled,omitempty"`
}

// SmbClientPolicyRule represents a rule from GET /smb-client-policies/rules.
type SmbClientPolicyRule struct {
	ID         string         `json:"id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Index      int            `json:"index,omitempty"`
	Policy     NamedReference `json:"policy,omitempty"`
	Client     string         `json:"client,omitempty"`
	Encryption string         `json:"encryption,omitempty"`
	Permission string         `json:"permission,omitempty"`
}

// SmbClientPolicyRuleInPolicy is an SMB client rule embedded inside a policy GET response.
type SmbClientPolicyRuleInPolicy struct {
	Name       string `json:"name,omitempty"`
	Index      int    `json:"index,omitempty"`
	Client     string `json:"client,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	Permission string `json:"permission,omitempty"`
}

// SmbClientPolicyRulePost contains the writable fields for POST /smb-client-policies/rules.
type SmbClientPolicyRulePost struct {
	Client     string `json:"client,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	Permission string `json:"permission,omitempty"`
	Index      *int   `json:"index,omitempty"`
}

// SmbClientPolicyRulePatch contains pointer fields for PATCH /smb-client-policies/rules.
type SmbClientPolicyRulePatch struct {
	Client     *string `json:"client,omitempty"`
	Encryption *string `json:"encryption,omitempty"`
	Permission *string `json:"permission,omitempty"`
	Index      *int    `json:"index,omitempty"`
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
// NOTE: suffix is NOT a valid field in add_rules — the API returns HTTP 400 if it is sent.
// suffix is read-only and appears only in GET responses (SnapshotPolicyRuleInPolicy).
type SnapshotPolicyRulePost struct {
	AtTime     *int64 `json:"at,omitempty"`
	Every      *int64 `json:"every,omitempty"`
	KeepFor    *int64 `json:"keep_for,omitempty"`
	ClientName string `json:"client_name,omitempty"`
}

// SnapshotPolicyRuleRemove identifies a rule to remove via remove_rules.
// FlashBlade identifies rules by their scheduling fields, not by name.
type SnapshotPolicyRuleRemove struct {
	Every   int64  `json:"every,omitempty"`
	At      int64  `json:"at,omitempty"`
	KeepFor int64  `json:"keep_for,omitempty"`
	Suffix  string `json:"suffix,omitempty"`
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

// NumericIDReference is a reference where the ID is a number (used for user/group references).
type NumericIDReference struct {
	Name string `json:"name,omitempty"`
	ID   int64  `json:"id,omitempty"`
}

// QuotaUser represents a per-filesystem user quota from GET /quotas/users.
type QuotaUser struct {
	FileSystem *NamedReference     `json:"file_system,omitempty"`
	User       *NumericIDReference `json:"user,omitempty"`
	Quota      int64               `json:"quota"`
	Usage      int64               `json:"usage,omitempty"`
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
	FileSystem *NamedReference     `json:"file_system,omitempty"`
	Group      *NumericIDReference `json:"group,omitempty"`
	Quota      int64               `json:"quota"`
	Usage      int64               `json:"usage,omitempty"`
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

// ServerDNS represents the DNS configuration block for a FlashBlade server.
type ServerDNS struct {
	Domain      string   `json:"domain,omitempty"`
	Nameservers []string `json:"nameservers,omitempty"`
	Services    []string `json:"services,omitempty"`
}

// Server represents a FlashBlade server object from GET /servers responses.
type Server struct {
	Name    string      `json:"name"`
	ID      string      `json:"id"`
	Created int64       `json:"created,omitempty"`
	DNS     []ServerDNS `json:"dns,omitempty"`
}

// ServerPost contains the fields accepted on POST /servers.
// The server name is passed via the ?create_ds= query parameter.
type ServerPost struct {
	DNS []ServerDNS `json:"dns,omitempty"`
}

// ServerPatch contains the fields accepted on PATCH /servers.
type ServerPatch struct {
	DNS []ServerDNS `json:"dns,omitempty"`
}

// FileSystemExport represents a FlashBlade file system export from GET responses.
type FileSystemExport struct {
	Name        string          `json:"name"`
	ID          string          `json:"id"`
	ExportName  string          `json:"export_name"`
	Enabled     bool            `json:"enabled"`
	Member      *NamedReference `json:"member,omitempty"`
	Server      *NamedReference `json:"server,omitempty"`
	Policy      *NamedReference `json:"policy,omitempty"`
	PolicyType  string          `json:"policy_type,omitempty"`
	SharePolicy *NamedReference `json:"share_policy,omitempty"`
	Status      string          `json:"status,omitempty"`
}

// FileSystemExportPost contains the fields accepted on POST /file-system-exports.
// The member_names query parameter specifies which file system to export.
type FileSystemExportPost struct {
	ExportName  string          `json:"export_name,omitempty"`
	Server      *NamedReference `json:"server,omitempty"`
	SharePolicy *NamedReference `json:"share_policy,omitempty"`
}

// FileSystemExportPatch contains pointer fields for PATCH /file-system-exports.
type FileSystemExportPatch struct {
	ExportName  *string         `json:"export_name,omitempty"`
	Server      *NamedReference `json:"server,omitempty"`
	SharePolicy *NamedReference `json:"share_policy,omitempty"`
}

// ObjectStoreAccountExport represents a FlashBlade object store account export from GET responses.
type ObjectStoreAccountExport struct {
	Name    string          `json:"name"`
	ID      string          `json:"id"`
	Enabled bool            `json:"enabled"`
	Member  *NamedReference `json:"member,omitempty"`
	Server  *NamedReference `json:"server,omitempty"`
	Policy  *NamedReference `json:"policy,omitempty"`
}

// ObjectStoreAccountExportPost contains the fields accepted on POST /object-store-account-exports.
// The member_names query parameter specifies which account to export.
type ObjectStoreAccountExportPost struct {
	ExportEnabled bool            `json:"export_enabled"`
	Server        *NamedReference `json:"server,omitempty"`
}

// ObjectStoreAccountExportPatch contains pointer fields for PATCH /object-store-account-exports.
type ObjectStoreAccountExportPatch struct {
	ExportEnabled *bool           `json:"export_enabled,omitempty"`
	Policy        *NamedReference `json:"policy,omitempty"`
}

// ---------- Phase 7 model structs -------------------------------------------

// S3ExportPolicy represents a FlashBlade S3 export policy from GET responses.
type S3ExportPolicy struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	IsLocal    bool   `json:"is_local,omitempty"`
	PolicyType string `json:"policy_type,omitempty"`
	Version    string `json:"version,omitempty"`
}

// S3ExportPolicyPost contains the fields accepted on POST /s3-export-policies.
type S3ExportPolicyPost struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// S3ExportPolicyPatch contains pointer fields for PATCH /s3-export-policies.
type S3ExportPolicyPatch struct {
	Name    *string `json:"name,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

// S3ExportPolicyRule represents a rule from GET /s3-export-policies/rules.
type S3ExportPolicyRule struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Index     int            `json:"index"`
	Policy    NamedReference `json:"policy,omitempty"`
	Effect    string         `json:"effect,omitempty"`
	Actions   []string       `json:"actions,omitempty"`
	Resources []string       `json:"resources,omitempty"`
}

// S3ExportPolicyRulePost contains the writable fields for POST /s3-export-policies/rules.
type S3ExportPolicyRulePost struct {
	Effect    string   `json:"effect"`
	Actions   []string `json:"actions"`
	Resources []string `json:"resources"`
}

// S3ExportPolicyRulePatch contains pointer fields for PATCH /s3-export-policies/rules.
type S3ExportPolicyRulePatch struct {
	Effect    *string  `json:"effect,omitempty"`
	Actions   []string `json:"actions,omitempty"`
	Resources []string `json:"resources,omitempty"`
}

// ObjectStoreVirtualHost represents a FlashBlade virtual host from GET responses.
type ObjectStoreVirtualHost struct {
	ID              string           `json:"id,omitempty"`
	Name            string           `json:"name"`
	Hostname        string           `json:"hostname,omitempty"`
	AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}

// ObjectStoreVirtualHostPost contains the fields accepted on POST /object-store-virtual-hosts.
type ObjectStoreVirtualHostPost struct {
	Hostname        string           `json:"hostname"`
	AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}

// ObjectStoreVirtualHostPatch contains pointer fields for PATCH /object-store-virtual-hosts.
type ObjectStoreVirtualHostPatch struct {
	Name            *string          `json:"name,omitempty"`
	Hostname        *string          `json:"hostname,omitempty"`
	AttachedServers []NamedReference `json:"attached_servers,omitempty"`
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

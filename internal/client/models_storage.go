package client

import "encoding/json"

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

// FileSystem represents a FlashBlade file system object from GET responses.
type FileSystem struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Provisioned      int64         `json:"provisioned"`
	Destroyed        bool          `json:"destroyed"`
	TimeRemaining    int64         `json:"time_remaining,omitempty"`
	Created          int64         `json:"created,omitempty"`
	Space            Space         `json:"space"`
	NFS              NFSConfig     `json:"nfs"`
	SMB              SMBConfig     `json:"smb"`
	HTTP             HTTPConfig    `json:"http"`
	DefaultQuotas    DefaultQuotas `json:"default_quotas"`
	MultiProtocol    MultiProtocol `json:"multi_protocol"`
	PromotionStatus  string        `json:"promotion_status,omitempty"`
	RequestNumber    int64         `json:"request_number,omitempty"`
	Source           *NamedReference `json:"source,omitempty"`
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

// ObjectStoreAccount represents a FlashBlade object store account from GET responses.
type ObjectStoreAccount struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Created          int64  `json:"created,omitempty"`
	QuotaLimit       int64 `json:"quota_limit,omitempty"`
	HardLimitEnabled bool   `json:"hard_limit_enabled"`
	ObjectCount      int64  `json:"object_count,omitempty"`
	Space            Space  `json:"space"`
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

// EradicationConfig represents the eradication configuration for a bucket.
type EradicationConfig struct {
	EradicationDelay int64  `json:"eradication_delay,omitempty"`
	EradicationMode  string `json:"eradication_mode,omitempty"`
	ManualEradication string `json:"manual_eradication,omitempty"`
}

// ObjectLockConfig represents the S3 object lock configuration for a bucket.
type ObjectLockConfig struct {
	FreezeLockedObjects bool   `json:"freeze_locked_objects,omitempty"`
	DefaultRetention     int64  `json:"default_retention,omitempty"`
	DefaultRetentionMode string `json:"default_retention_mode,omitempty"`
	ObjectLockEnabled    bool   `json:"object_lock_enabled,omitempty"`
}

// PublicAccessConfig represents the public access configuration for a bucket.
type PublicAccessConfig struct {
	BlockNewPublicPolicies bool `json:"block_new_public_policies,omitempty"`
	BlockPublicAccess      bool `json:"block_public_access,omitempty"`
}

// Bucket represents a FlashBlade object store bucket from GET responses.
type Bucket struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Account          NamedReference     `json:"account"`
	Created          int64              `json:"created,omitempty"`
	Destroyed        bool               `json:"destroyed"`
	TimeRemaining    int64              `json:"time_remaining,omitempty"`
	Versioning       string             `json:"versioning,omitempty"`
	QuotaLimit       int64              `json:"quota_limit,omitempty"`
	HardLimitEnabled bool               `json:"hard_limit_enabled"`
	ObjectCount      int64              `json:"object_count,omitempty"`
	BucketType       string             `json:"bucket_type,omitempty"`
	RetentionLock    string             `json:"retention_lock,omitempty"`
	EradicationConfig  EradicationConfig  `json:"eradication_config"`
	ObjectLockConfig   ObjectLockConfig   `json:"object_lock_config"`
	PublicAccessConfig PublicAccessConfig  `json:"public_access_config"`
	PublicStatus       string             `json:"public_status,omitempty"`
	Space            Space              `json:"space"`
}

// BucketPost contains the fields accepted on POST /buckets.
// NOTE: quota_limit must be serialized as a string per FlashBlade API.
// NOTE: versioning is NOT a valid POST parameter — use PATCH after creation.
// NOTE: public_access_config is NOT valid on POST — PATCH only.
type BucketPost struct {
	Account           NamedReference     `json:"account"`
	QuotaLimit        string             `json:"quota_limit,omitempty"`
	HardLimitEnabled  bool               `json:"hard_limit_enabled,omitempty"`
	RetentionLock     string             `json:"retention_lock,omitempty"`
	EradicationConfig *EradicationConfig `json:"eradication_config,omitempty"`
	ObjectLockConfig  *ObjectLockConfig  `json:"object_lock_config,omitempty"`
}

// BucketPatch contains pointer fields for PATCH semantics on /buckets.
type BucketPatch struct {
	Destroyed          *bool              `json:"destroyed,omitempty"`
	Versioning         *string            `json:"versioning,omitempty"`
	QuotaLimit         *string            `json:"quota_limit,omitempty"`
	HardLimitEnabled   *bool              `json:"hard_limit_enabled,omitempty"`
	RetentionLock      *string            `json:"retention_lock,omitempty"`
	EradicationConfig  *EradicationConfig `json:"eradication_config,omitempty"`
	ObjectLockConfig   *ObjectLockConfig  `json:"object_lock_config,omitempty"`
	PublicAccessConfig *PublicAccessConfig `json:"public_access_config,omitempty"`
}

// ObjectStoreAccessKey represents a FlashBlade object store access key.
type ObjectStoreAccessKey struct {
	Name            string         `json:"name"`
	AccessKeyID     string         `json:"access_key_id"`
	SecretAccessKey string         `json:"secret_access_key,omitempty"`
	Created         int64          `json:"created,omitempty"`
	Enabled         bool           `json:"enabled"`
	User            NamedReference `json:"user"`
}

// ObjectStoreAccessKeyPost contains the fields for POST /object-store-access-keys.
type ObjectStoreAccessKeyPost struct {
	User            NamedReference `json:"user"`
	SecretAccessKey string         `json:"secret_access_key,omitempty"`
}

// ObjectStoreRemoteCredentials represents a FlashBlade remote credentials object from GET responses.
type ObjectStoreRemoteCredentials struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	AccessKeyID     string           `json:"access_key_id"`
	SecretAccessKey string           `json:"secret_access_key,omitempty"`
	Remote          NamedReference   `json:"remote"`
	Realms          []NamedReference `json:"realms,omitempty"`
}

// ObjectStoreRemoteCredentialsPost contains the fields for POST /object-store-remote-credentials.
// Name is passed via ?names= query param, remote via ?remote_names= query param.
type ObjectStoreRemoteCredentialsPost struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

// ObjectStoreRemoteCredentialsPatch contains pointer fields for PATCH semantics on /object-store-remote-credentials.
type ObjectStoreRemoteCredentialsPatch struct {
	AccessKeyID     *string `json:"access_key_id,omitempty"`
	SecretAccessKey *string `json:"secret_access_key,omitempty"`
}

// ObjectBacklog holds object backlog metrics for a bucket replica link.
type ObjectBacklog struct {
	Count     int64 `json:"count,omitempty"`
	TotalSize int64 `json:"total_size,omitempty"`
}

// BucketReplicaLink represents a FlashBlade bucket replica link from GET responses.
type BucketReplicaLink struct {
	ID                string          `json:"id"`
	LocalBucket       NamedReference  `json:"local_bucket"`
	RemoteBucket      NamedReference  `json:"remote_bucket"`
	Remote            NamedReference  `json:"remote"`
	RemoteCredentials *NamedReference `json:"remote_credentials,omitempty"`
	Paused            bool            `json:"paused"`
	CascadingEnabled  bool            `json:"cascading_enabled"`
	Direction         string          `json:"direction,omitempty"`
	Status            string          `json:"status,omitempty"`
	StatusDetails     string          `json:"status_details,omitempty"`
	Lag               int64           `json:"lag,omitempty"`
	RecoveryPoint     int64           `json:"recovery_point,omitempty"`
	ObjectBacklog     *ObjectBacklog  `json:"object_backlog,omitempty"`
}

// BucketReplicaLinkPost contains the fields for POST /bucket-replica-links.
// Local bucket, remote bucket, and remote credentials are all query params.
type BucketReplicaLinkPost struct {
	Paused           bool `json:"paused,omitempty"`
	CascadingEnabled bool `json:"cascading_enabled,omitempty"`
}

// BucketReplicaLinkPatch contains pointer fields for PATCH semantics on /bucket-replica-links.
type BucketReplicaLinkPatch struct {
	Paused *bool `json:"paused,omitempty"`
}

// LifecycleRule represents a FlashBlade lifecycle rule from GET responses.
type LifecycleRule struct {
	ID                                   string         `json:"id"`
	Name                                 string         `json:"name"`
	Bucket                               NamedReference `json:"bucket"`
	RuleID                               string         `json:"rule_id"`
	Prefix                               string         `json:"prefix"`
	Enabled                              bool           `json:"enabled"`
	AbortIncompleteMultipartUploadsAfter int64          `json:"abort_incomplete_multipart_uploads_after,omitempty"`
	KeepCurrentVersionFor                int64          `json:"keep_current_version_for,omitempty"`
	KeepCurrentVersionUntil              int64          `json:"keep_current_version_until,omitempty"`
	KeepPreviousVersionFor               int64          `json:"keep_previous_version_for,omitempty"`
	CleanupExpiredObjectDeleteMarker     bool           `json:"cleanup_expired_object_delete_marker,omitempty"`
}

// LifecycleRulePost contains the fields for POST /lifecycle-rules.
type LifecycleRulePost struct {
	Bucket                               NamedReference `json:"bucket"`
	RuleID                               string         `json:"rule_id"`
	Prefix                               string         `json:"prefix,omitempty"`
	AbortIncompleteMultipartUploadsAfter int64          `json:"abort_incomplete_multipart_uploads_after,omitempty"`
	KeepCurrentVersionFor                int64          `json:"keep_current_version_for,omitempty"`
	KeepCurrentVersionUntil              int64          `json:"keep_current_version_until,omitempty"`
	KeepPreviousVersionFor               int64          `json:"keep_previous_version_for,omitempty"`
}

// LifecycleRulePatch contains pointer fields for PATCH semantics on /lifecycle-rules.
type LifecycleRulePatch struct {
	Enabled                              *bool   `json:"enabled,omitempty"`
	Prefix                               *string `json:"prefix,omitempty"`
	AbortIncompleteMultipartUploadsAfter *int64  `json:"abort_incomplete_multipart_uploads_after,omitempty"`
	KeepCurrentVersionFor                *int64  `json:"keep_current_version_for,omitempty"`
	KeepCurrentVersionUntil              *int64  `json:"keep_current_version_until,omitempty"`
	KeepPreviousVersionFor               *int64  `json:"keep_previous_version_for,omitempty"`
}

// BucketAccessPolicy represents a FlashBlade bucket access policy from GET responses.
type BucketAccessPolicy struct {
	ID         string                   `json:"id,omitempty"`
	Name       string                   `json:"name,omitempty"`
	Bucket     NamedReference           `json:"bucket"`
	Enabled    bool                     `json:"enabled"`
	IsLocal    bool                     `json:"is_local,omitempty"`
	PolicyType string                   `json:"policy_type,omitempty"`
	Rules      []BucketAccessPolicyRule `json:"rules,omitempty"`
}

// BucketAccessPolicyRule represents a single rule within a bucket access policy.
type BucketAccessPolicyRule struct {
	Name       string                       `json:"name,omitempty"`
	Actions    []string                     `json:"actions"`
	Effect     string                       `json:"effect"`
	Principals BucketAccessPolicyPrincipals `json:"principals"`
	Resources  []string                     `json:"resources"`
	Policy     *NamedReference              `json:"policy,omitempty"`
}

// BucketAccessPolicyPrincipals represents the principals object within a bucket access policy rule.
type BucketAccessPolicyPrincipals struct {
	All []string `json:"all,omitempty"`
}

// BucketAccessPolicyPost contains the fields for POST /buckets/bucket-access-policies.
type BucketAccessPolicyPost struct {
	Rules []BucketAccessPolicyRulePost `json:"rules,omitempty"`
}

// BucketAccessPolicyRulePost contains the fields for POST /buckets/bucket-access-policies/rules.
type BucketAccessPolicyRulePost struct {
	Actions    []string                     `json:"actions"`
	Principals BucketAccessPolicyPrincipals `json:"principals"`
	Resources  []string                     `json:"resources"`
}

// BucketAuditFilter represents a FlashBlade bucket audit filter from GET responses.
type BucketAuditFilter struct {
	Actions    []string       `json:"actions"`
	Bucket     NamedReference `json:"bucket"`
	Name       string         `json:"name"`
	S3Prefixes []string       `json:"s3_prefixes"`
}

// BucketAuditFilterPost contains the fields for POST /buckets/audit-filters.
type BucketAuditFilterPost struct {
	Actions    []string `json:"actions"`
	S3Prefixes []string `json:"s3_prefixes"`
}

// BucketAuditFilterPatch contains pointer fields for PATCH /buckets/audit-filters.
type BucketAuditFilterPatch struct {
	Actions    *[]string `json:"actions,omitempty"`
	S3Prefixes *[]string `json:"s3_prefixes,omitempty"`
}

// QosPolicy represents a FlashBlade QoS policy from GET responses.
type QosPolicy struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Enabled              bool   `json:"enabled"`
	IsLocal              bool   `json:"is_local"`
	MaxTotalBytesPerSec  int64  `json:"max_total_bytes_per_sec"`
	MaxTotalOpsPerSec    int64  `json:"max_total_ops_per_sec"`
	PolicyType           string `json:"policy_type"`
}

// QosPolicyPost contains the fields for POST /qos-policies.
type QosPolicyPost struct {
	Name                string `json:"name"`
	Enabled             *bool  `json:"enabled,omitempty"`
	MaxTotalBytesPerSec int64  `json:"max_total_bytes_per_sec,omitempty"`
	MaxTotalOpsPerSec   int64  `json:"max_total_ops_per_sec,omitempty"`
}

// QosPolicyPatch contains pointer fields for PATCH /qos-policies.
type QosPolicyPatch struct {
	Enabled             *bool   `json:"enabled,omitempty"`
	MaxTotalBytesPerSec *int64  `json:"max_total_bytes_per_sec,omitempty"`
	MaxTotalOpsPerSec   *int64  `json:"max_total_ops_per_sec,omitempty"`
	Name                *string `json:"name,omitempty"`
}

// QosPolicyMember represents a QoS policy member assignment from GET responses.
type QosPolicyMember struct {
	Member NamedReference `json:"member"`
	Policy NamedReference `json:"policy"`
}

// ObjectStoreUser represents a FlashBlade object store user from GET responses.
type ObjectStoreUser struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	FullAccess bool   `json:"full_access"`
}

// ObjectStoreUserPost contains the fields for POST /object-store-users.
type ObjectStoreUserPost struct {
	FullAccess *bool `json:"full_access,omitempty"`
}

// ObjectStoreUserPolicyMember represents a user-to-policy association from GET responses.
type ObjectStoreUserPolicyMember struct {
	Member NamedReference `json:"member"`
	Policy NamedReference `json:"policy"`
}

// QosPolicyMemberPost contains the fields for POST /qos-policies/members.
type QosPolicyMemberPost struct {
	Member NamedReference `json:"member"`
}

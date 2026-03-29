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
	Space            Space         `json:"space"`
	NFS              NFSConfig     `json:"nfs"`
	SMB              SMBConfig     `json:"smb"`
	HTTP             HTTPConfig    `json:"http"`
	DefaultQuotas    DefaultQuotas `json:"default_quotas"`
	MultiProtocol    MultiProtocol `json:"multi_protocol"`
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

// Bucket represents a FlashBlade object store bucket from GET responses.
type Bucket struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Account          NamedReference `json:"account"`
	Created          int64          `json:"created,omitempty"`
	Destroyed        bool           `json:"destroyed"`
	TimeRemaining    int64          `json:"time_remaining,omitempty"`
	Versioning       string         `json:"versioning,omitempty"`
	QuotaLimit       int64          `json:"quota_limit,omitempty"`
	HardLimitEnabled bool           `json:"hard_limit_enabled"`
	ObjectCount      int64          `json:"object_count,omitempty"`
	BucketType       string         `json:"bucket_type,omitempty"`
	RetentionLock    string         `json:"retention_lock,omitempty"`
	Space            Space          `json:"space"`
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
	User            NamedReference `json:"user"`
}

// ObjectStoreAccessKeyPost contains the fields for POST /object-store-access-keys.
type ObjectStoreAccessKeyPost struct {
	User            NamedReference `json:"user"`
	SecretAccessKey string         `json:"secret_access_key,omitempty"`
}

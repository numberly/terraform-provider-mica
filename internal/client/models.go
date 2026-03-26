package client

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

package client

// AuditObjectStorePolicy represents a FlashBlade audit object store policy from GET responses.
type AuditObjectStorePolicy struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Enabled    bool             `json:"enabled"`
	IsLocal    bool             `json:"is_local"`
	PolicyType string           `json:"policy_type"`
	Location   *NamedReference  `json:"location,omitempty"`
	Realms     []NamedReference `json:"realms,omitempty"`
	LogTargets []NamedReference `json:"log_targets"`
}

// AuditObjectStorePolicyMember represents a member of an audit object store policy from GET responses.
type AuditObjectStorePolicyMember struct {
	Member NamedReference `json:"member"`
	Policy NamedReference `json:"policy"`
}

// AuditObjectStorePolicyPost contains the fields for POST /audit-object-store-policies.
type AuditObjectStorePolicyPost struct {
	Enabled    *bool            `json:"enabled,omitempty"`
	LogTargets []NamedReference `json:"log_targets,omitempty"`
}

// AuditObjectStorePolicyPatch contains pointer fields for PATCH /audit-object-store-policies.
type AuditObjectStorePolicyPatch struct {
	Enabled    *bool             `json:"enabled,omitempty"`
	LogTargets *[]NamedReference `json:"log_targets,omitempty"`
}

// AuditLogNamePrefix holds the prefix string for audit log object names.
type AuditLogNamePrefix struct {
	Prefix string `json:"prefix,omitempty"`
}

// AuditLogRotate holds the rotation configuration for audit logs.
type AuditLogRotate struct {
	// Duration is the rotation interval in milliseconds.
	Duration int64 `json:"duration,omitempty"`
}

// LogTargetObjectStore represents a FlashBlade log target object store from GET responses.
type LogTargetObjectStore struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Bucket        NamedReference     `json:"bucket"`
	LogNamePrefix AuditLogNamePrefix `json:"log_name_prefix"`
	LogRotate     AuditLogRotate     `json:"log_rotate"`
}

// LogTargetObjectStorePost contains the fields accepted on POST /log-targets/object-store.
// Name is passed via ?names= query parameter, not in the body.
type LogTargetObjectStorePost struct {
	Bucket        NamedReference     `json:"bucket"`
	LogNamePrefix AuditLogNamePrefix `json:"log_name_prefix"`
	LogRotate     AuditLogRotate     `json:"log_rotate"`
}

// LogTargetObjectStorePatch contains pointer fields for PATCH /log-targets/object-store.
// Nil pointer means omit the field; non-nil means send the value.
type LogTargetObjectStorePatch struct {
	Bucket        *NamedReference     `json:"bucket,omitempty"`
	LogNamePrefix *AuditLogNamePrefix `json:"log_name_prefix,omitempty"`
	LogRotate     *AuditLogRotate     `json:"log_rotate,omitempty"`
}

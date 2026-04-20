package client

import "encoding/json"

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
	Policy                    NamedReference `json:"policy"`
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
	Security                  *[]string `json:"security,omitempty"`
	RequiredTransportSecurity *string   `json:"required_transport_security,omitempty"`
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
	Policy      NamedReference `json:"policy"`
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
	Policy     NamedReference `json:"policy"`
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
// List fields use *[]string per CONVENTIONS.md §Pointer rules: nil = omit, &[]string{} = clear, &[...] = set.
type ObjectStoreAccessPolicyRulePatch struct {
	Actions    *[]string       `json:"actions,omitempty"`
	Conditions json.RawMessage `json:"conditions,omitempty"`
	Resources  *[]string       `json:"resources,omitempty"`
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
// List fields use *[]string per CONVENTIONS.md §Pointer rules.
type NetworkAccessPolicyRulePatch struct {
	Client     *string   `json:"client,omitempty"`
	Effect     *string   `json:"effect,omitempty"`
	Index      *int      `json:"index,omitempty"`
	Interfaces *[]string `json:"interfaces,omitempty"`
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
	Policy    NamedReference `json:"policy"`
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
// List fields use *[]string per CONVENTIONS.md §Pointer rules.
type S3ExportPolicyRulePatch struct {
	Effect    *string   `json:"effect,omitempty"`
	Actions   *[]string `json:"actions,omitempty"`
	Resources *[]string `json:"resources,omitempty"`
}

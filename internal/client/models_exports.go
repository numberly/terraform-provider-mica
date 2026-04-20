package client

// Server represents a FlashBlade server object from GET /servers responses.
type Server struct {
	Name              string           `json:"name"`
	ID                string           `json:"id"`
	Created           int64            `json:"created,omitempty"`
	DNS               []NamedReference `json:"dns,omitempty"`
	DirectoryServices []NamedReference `json:"directory_services,omitempty"`
}

// ServerPost contains the fields accepted on POST /servers.
// The server name is passed via the ?create_ds= query parameter.
type ServerPost struct {
	DNS []NamedReference `json:"dns,omitempty"`
}

// ServerPatch contains the fields accepted on PATCH /servers.
type ServerPatch struct {
	DNS []NamedReference `json:"dns,omitempty"`
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
// Server and SharePolicy are **NamedReference (CONVENTIONS.md §Pointer rules, PATCH struct):
//   - nil outer                    → omit
//   - non-nil outer, nil inner     → clear (send JSON null)
//   - non-nil outer, non-nil inner → set value
type FileSystemExportPatch struct {
	ExportName  *string          `json:"export_name,omitempty"`
	Server      **NamedReference `json:"server,omitempty"`
	SharePolicy **NamedReference `json:"share_policy,omitempty"`
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

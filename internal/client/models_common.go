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

// NamedReference is a lightweight reference to another object by name and ID.
type NamedReference struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// NumericIDReference is a reference where the ID is a number (used for user/group references).
type NumericIDReference struct {
	Name string `json:"name,omitempty"`
	ID   int64  `json:"id,omitempty"`
}

// PolicyMember represents a file system that is a member of a policy.
// Used for delete-guard checks across all policy families.
type PolicyMember struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

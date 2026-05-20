package client

// ResiliencyGroup represents a FlashBlade resiliency group from
// GET /api/2.23/resiliency-groups.
//
// Resiliency groups are hardware-managed high-availability groupings reported
// by the array. They are read-only: no POST/PATCH/DELETE endpoints exist, so
// there is no ResiliencyGroupPost or ResiliencyGroupPatch counterpart.
type ResiliencyGroup struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	Status        string `json:"status,omitempty"`
	StatusDetails string `json:"status_details,omitempty"`
}

// ResiliencyReference is a typed reference with id, name and resource_type as
// returned by /api/2.23/resiliency-groups/members. It mirrors the API _reference
// schema; we keep it local rather than extending NamedReference because the
// resource_type field is meaningful here (the referent can be a resiliency
// group on one side and an arbitrary member resource on the other).
type ResiliencyReference struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
}

// ResiliencyGroupMember represents a single membership row from
// GET /api/2.23/resiliency-groups/members.
//
// Each row links a resiliency group to one of its member resources.
// Members are hardware-managed; no POST/PATCH/DELETE endpoints exist.
type ResiliencyGroupMember struct {
	Group  ResiliencyReference `json:"group"`
	Member ResiliencyReference `json:"member"`
}

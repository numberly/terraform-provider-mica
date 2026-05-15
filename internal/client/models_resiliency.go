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

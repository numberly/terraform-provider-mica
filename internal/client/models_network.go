package client

// Subnet represents a FlashBlade subnet from GET /api/2.22/subnets.
// All fields except Name use omitempty; Name is always present in GET responses.
type Subnet struct {
	ID                   string           `json:"id,omitempty"`
	Name                 string           `json:"name"`
	Enabled              bool             `json:"enabled,omitempty"`
	Gateway              string           `json:"gateway,omitempty"`
	Interfaces           []NamedReference `json:"interfaces,omitempty"`
	LinkAggregationGroup *NamedReference  `json:"link_aggregation_group,omitempty"`
	MTU                  int64            `json:"mtu,omitempty"`
	Prefix               string           `json:"prefix,omitempty"`
	Services             []string         `json:"services,omitempty"`
	VLAN                 int64            `json:"vlan,omitempty"`
}

// SubnetPost contains writable fields for POST /api/2.22/subnets?names=<name>.
// Name is NOT included — it is passed as the ?names= query parameter.
// Enabled, ID, Interfaces, and Services are read-only (ro) in the API spec and must not be sent.
type SubnetPost struct {
	Gateway              string          `json:"gateway,omitempty"`
	LinkAggregationGroup *NamedReference `json:"link_aggregation_group,omitempty"`
	MTU                  int64           `json:"mtu,omitempty"`
	Prefix               string          `json:"prefix,omitempty"`
	VLAN                 int64           `json:"vlan,omitempty"`
}

// SubnetPatch contains writable fields for PATCH /api/2.22/subnets?names=<name>.
// Pointer types allow true omission of unchanged fields.
// *int64 is used for MTU and VLAN so that zero values (e.g., VLAN=0 for untagged) are serializable.
type SubnetPatch struct {
	Gateway              *string         `json:"gateway,omitempty"`
	LinkAggregationGroup *NamedReference `json:"link_aggregation_group,omitempty"`
	MTU                  *int64          `json:"mtu,omitempty"`
	Prefix               *string         `json:"prefix,omitempty"`
	VLAN                 *int64          `json:"vlan,omitempty"`
}

// LinkAggregationGroup represents a FlashBlade LAG from GET /api/2.22/link-aggregation-groups.
// All fields are read-only — the struct is used only in GET responses.
// LAGs are hardware-managed and cannot be created, updated, or deleted via the API.
type LinkAggregationGroup struct {
	ID         string   `json:"id,omitempty"`
	Name       string   `json:"name"`
	LagSpeed   int64    `json:"lag_speed,omitempty"`
	MacAddress string   `json:"mac_address,omitempty"`
	PortSpeed  int64    `json:"port_speed,omitempty"`
	Ports      []string `json:"ports,omitempty"`
	Status     string   `json:"status,omitempty"`
}

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
	Ports      []NamedReference `json:"ports,omitempty"`
	Status     string   `json:"status,omitempty"`
}

// NetworkInterface represents a FlashBlade network interface from GET /api/2.22/network-interfaces.
// All fields except Name use omitempty; Name is always present in GET responses.
type NetworkInterface struct {
	ID              string           `json:"id,omitempty"`
	Name            string           `json:"name"`
	Address         string           `json:"address,omitempty"`
	Enabled         bool             `json:"enabled,omitempty"`
	Gateway         string           `json:"gateway,omitempty"`
	MTU             int64            `json:"mtu,omitempty"`
	Netmask         string           `json:"netmask,omitempty"`
	Services        []string         `json:"services,omitempty"`
	Subnet          *NamedReference  `json:"subnet,omitempty"`
	Type            string           `json:"type,omitempty"`
	VLAN            int64            `json:"vlan,omitempty"`
	AttachedServers []NamedReference `json:"attached_servers,omitempty"`
	Realms          []string         `json:"realms,omitempty"`
}

// NetworkInterfacePost contains writable fields for POST /api/2.22/network-interfaces?names=<name>.
// Name is NOT included — it is passed as the ?names= query parameter.
// Subnet is passed via the ?subnet_names= query parameter, not in the body.
type NetworkInterfacePost struct {
	Address         string           `json:"address,omitempty"`
	Services        []string         `json:"services,omitempty"`
	Type            string           `json:"type,omitempty"`
	AttachedServers []NamedReference `json:"attached_servers,omitempty"`
}

// NetworkInterfacePatch contains mutable fields for PATCH /api/2.22/network-interfaces?names=<name>.
// Address uses *string + omitempty for true PATCH semantics (only sent when changed).
// Services and AttachedServers do NOT use omitempty — clearing them requires sending [] in JSON.
type NetworkInterfacePatch struct {
	Address         *string          `json:"address,omitempty"`
	Services        []string         `json:"services"`
	AttachedServers []NamedReference `json:"attached_servers"`
}

// Certificate represents a FlashBlade certificate from GET responses.
type Certificate struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	Certificate             string   `json:"certificate"`
	CertificateType         string   `json:"certificate_type"`
	CommonName              string   `json:"common_name"`
	Country                 string   `json:"country"`
	Email                   string   `json:"email"`
	IntermediateCertificate string   `json:"intermediate_certificate"`
	IssuedBy                string   `json:"issued_by"`
	IssuedTo                string   `json:"issued_to"`
	KeyAlgorithm            string   `json:"key_algorithm"`
	KeySize                 int      `json:"key_size"`
	Locality                string   `json:"locality"`
	Organization            string   `json:"organization"`
	OrganizationalUnit      string   `json:"organizational_unit"`
	State                   string   `json:"state"`
	Status                  string   `json:"status"`
	SubjectAlternativeNames []string `json:"subject_alternative_names"`
	ValidFrom               int64    `json:"valid_from"`
	ValidTo                 int64    `json:"valid_to"`
}

// CertificatePost contains the fields for POST /certificates.
// Name is passed via ?names= query param.
// For import mode: certificate + private_key required, passphrase + intermediate_certificate optional.
type CertificatePost struct {
	Certificate             string `json:"certificate"`
	CertificateType         string `json:"certificate_type,omitempty"`
	IntermediateCertificate string `json:"intermediate_certificate,omitempty"`
	Passphrase              string `json:"passphrase,omitempty"`
	PrivateKey              string `json:"private_key,omitempty"`
}

// CertificatePatch contains pointer fields for PATCH semantics on /certificates.
// A nil pointer means omit the field from the JSON body.
type CertificatePatch struct {
	Certificate             *string `json:"certificate,omitempty"`
	IntermediateCertificate *string `json:"intermediate_certificate,omitempty"`
	Passphrase              *string `json:"passphrase,omitempty"`
	PrivateKey              *string `json:"private_key,omitempty"`
}

// TlsPolicy represents a FlashBlade TLS policy from GET /api/2.22/tls-policies.
type TlsPolicy struct {
	ID                               string          `json:"id"`
	Name                             string          `json:"name"`
	ApplianceCertificate             *NamedReference `json:"appliance_certificate"`
	ClientCertificatesRequired       bool            `json:"client_certificates_required"`
	DisabledTlsCiphers               []string        `json:"disabled_tls_ciphers"`
	Enabled                          bool            `json:"enabled"`
	EnabledTlsCiphers                []string        `json:"enabled_tls_ciphers"`
	IsLocal                          bool            `json:"is_local"`
	MinTlsVersion                    string          `json:"min_tls_version"`
	PolicyType                       string          `json:"policy_type"`
	TrustedClientCertificateAuthority *NamedReference `json:"trusted_client_certificate_authority"`
	VerifyClientCertificateTrust     bool            `json:"verify_client_certificate_trust"`
}

// TlsPolicyPost contains writable fields for POST /api/2.22/tls-policies.
// Name is passed via ?names= query parameter, not in body.
type TlsPolicyPost struct {
	ApplianceCertificate             *NamedReference `json:"appliance_certificate,omitempty"`
	ClientCertificatesRequired       bool            `json:"client_certificates_required,omitempty"`
	DisabledTlsCiphers               []string        `json:"disabled_tls_ciphers,omitempty"`
	Enabled                          bool            `json:"enabled,omitempty"`
	EnabledTlsCiphers                []string        `json:"enabled_tls_ciphers,omitempty"`
	MinTlsVersion                    string          `json:"min_tls_version,omitempty"`
	TrustedClientCertificateAuthority *NamedReference `json:"trusted_client_certificate_authority,omitempty"`
	VerifyClientCertificateTrust     bool            `json:"verify_client_certificate_trust,omitempty"`
}

// TlsPolicyPatch contains pointer fields for PATCH semantics on /tls-policies.
// nil = omit from JSON. Non-nil = send.
// **NamedReference is used for ref fields: outer nil = omit, outer non-nil + inner nil = set to null,
// outer non-nil + inner non-nil = set value.
type TlsPolicyPatch struct {
	ApplianceCertificate             **NamedReference `json:"appliance_certificate,omitempty"`
	ClientCertificatesRequired       *bool            `json:"client_certificates_required,omitempty"`
	DisabledTlsCiphers               *[]string        `json:"disabled_tls_ciphers,omitempty"`
	Enabled                          *bool            `json:"enabled,omitempty"`
	EnabledTlsCiphers                *[]string        `json:"enabled_tls_ciphers,omitempty"`
	MinTlsVersion                    *string          `json:"min_tls_version,omitempty"`
	TrustedClientCertificateAuthority **NamedReference `json:"trusted_client_certificate_authority,omitempty"`
	VerifyClientCertificateTrust     *bool            `json:"verify_client_certificate_trust,omitempty"`
}

// TlsPolicyMember represents an association between a TLS policy and a network interface.
// Returned by GET /api/2.22/tls-policies/members and GET /api/2.22/network-interfaces/tls-policies.
type TlsPolicyMember struct {
	Policy NamedReference `json:"policy"`
	Member NamedReference `json:"member"`
}

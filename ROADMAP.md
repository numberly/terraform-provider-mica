# API Coverage Roadmap

FlashBlade® REST API v2.22 (Purity//FB 4.6.7) coverage status for terraform-provider-mica.

**Last updated:** 2026-04-17
**Provider version:** v2.22.2
**Total API sections:** 84 | **Covered:** ~44 | **Coverage of IaC-relevant CRUD:** ~75%

## Coverage Legend

| Status | Meaning |
|--------|---------|
| Done | Resource + Data Source implemented, tested, documented |
| DS-only | Data Source only (read-only) |
| Partial | Some endpoints covered, gaps remain |
| Planned | In a future milestone |
| Candidate | Useful but not yet scheduled |
| Deferred | Low priority or niche |
| N/A | Not applicable for Terraform (read-only metrics, hardware, etc.) |

## Implemented

### Storage

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| File Systems | `flashblade_file_system` | Yes | Done | CRUD + soft-delete + eradication |
| Buckets | `flashblade_bucket` | Yes | Done | Versioning, quota, eradication, object lock, public access |
| Object Store Accounts | `flashblade_object_store_account` | Yes | Done | S3 namespace |
| Object Store Access Keys | `flashblade_object_store_access_key` | Yes | Done | Cross-array secret sharing |
| Object Store Users | `flashblade_object_store_user` | Yes | Done | CRD only (no PATCH); full_access optional; import supported |
| Object Store User Policy | `flashblade_object_store_user_policy` | No | Done | user to policy association; import format: account/username/policyname |

### Bucket Sub-resources

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| Lifecycle Rules | `flashblade_lifecycle_rule` | Yes | Done | Version retention, multipart cleanup |
| Bucket Access Policies | `flashblade_bucket_access_policy` | Yes | Done | IAM-style per-bucket |
| Bucket Access Policy Rules | `flashblade_bucket_access_policy_rule` | No | Done | |
| Bucket Audit Filters | `flashblade_bucket_audit_filter` | Yes | Done | Actions + prefix filtering |
| QoS Policies | `flashblade_qos_policy` | Yes | Done | Bandwidth + IOPS limits |
| QoS Policy Members | `flashblade_qos_policy_member` | No | Done | FS assignment (buckets not supported on v2.22) |

### Policies

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| NFS Export Policy + Rules | Yes + Yes | Yes | Done | Full CRUD |
| SMB Share Policy + Rules | Yes + Yes | Yes | Done | Full CRUD |
| SMB Client Policy + Rules | Yes + Yes | Yes | Done | Full CRUD |
| Snapshot Policy + Rules | Yes + Yes | Yes | Done | Full CRUD |
| Network Access Policy + Rules | Yes + Yes | Yes | Done | Singleton + rules |
| Object Store Access Policy + Rules | Yes + Yes | Yes | Done | Full CRUD |
| S3 Export Policy + Rules | Yes + Yes | Yes | Done | Full CRUD |

### Servers & Exports

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| Servers | `flashblade_server` | Yes | Done | DNS, directory_services, network_interfaces |
| File System Exports | `flashblade_file_system_export` | Yes | Done | NFS export to server |
| Account Exports | `flashblade_object_store_account_export` | Yes | Done | S3 export to server |
| Virtual Hosts | `flashblade_object_store_virtual_host` | Yes | Done | S3 virtual-hosted endpoints |

### Networking

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| Subnets | `flashblade_subnet` | Yes | Done | Prefix, gateway, MTU, VLAN, LAG |
| Network Interfaces | `flashblade_network_interface` | Yes | Done | VIP: data, sts, egress-only, replication |
| Link Aggregation Groups | No | Yes | DS-only | Hardware-managed, read-only |

### Replication

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| Bucket Replica Links | `flashblade_bucket_replica_link` | Yes | Done | Bidirectional, pause/resume; Read/Import by UUID (supports multiple links per bucket pair) |
| Remote Credentials | `flashblade_remote_credentials` | Yes | Done | S3 cross-array + S3 target auth; target_name attr (v2.2) |
| Targets | `flashblade_target` | Yes | Done | External S3 replication endpoints (address, CA cert group) |
| Array Connections | `flashblade_array_connection` | Yes | Done | Full CRUD, connection_key sensitive/write-only, throttle, replication_addresses |

### Security & TLS

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| Certificates | `flashblade_certificate` | Yes | Done | Import PEM + private key; drift detection; import→apply skips PATCH (state-only update) |
| TLS Policies | `flashblade_tls_policy` | Yes | Done | Cipher suites, min version, mTLS; API-default handling; + `flashblade_tls_policy_member` for NI assignment |
| Certificate Groups | `flashblade_certificate_group` | Yes | Done | CRD-only (no PATCH); realms volatile; + `flashblade_certificate_group_member` for cert association |

### Array Administration

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| DNS | `flashblade_array_dns` | Yes | Done | Singleton |
| NTP | `flashblade_array_ntp` | Yes | Done | Singleton |
| SMTP | `flashblade_array_smtp` | Yes | Done | Singleton |
| Syslog Servers | `flashblade_syslog_server` | Yes | Done | Full CRUD |
| Directory Services (Management) | `flashblade_directory_service_management` | Yes | Done | Singleton; LDAP admin auth; write-only bind_password |
| Directory Services (Roles) | `flashblade_directory_service_role` | Yes | Done | LDAP group → management access policy mapping; user-supplied name via ?names= (50.1); v2.22.2 |
| Management Access Policy DS Role Membership | `flashblade_management_access_policy_directory_service_role_membership` | No | Done | Additive policy-to-role association; composite ID role_name/policy_name; v2.22.2 |

### Audit

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| Audit Policies (Object Store) | `flashblade_audit_object_store_policy` | Yes | Done | Object store audit logging policies with log targets |
| Audit Policies (Object Store) Members | `flashblade_audit_object_store_policy_member` | No | Done | Bucket-to-policy assignment (CRD) |
| Log Targets (Object Store) | `flashblade_log_target_object_store` | Yes | Done | Audit log target to bucket with prefix + rotation |

### Quotas

| API Section | Resource | Data Source | Status | Notes |
|-------------|----------|:----------:|--------|-------|
| Quota Users | `flashblade_quota_user` | Yes | Done | Per-filesystem |
| Quota Groups | `flashblade_quota_group` | Yes | Done | Per-filesystem |

---

## Not Implemented

### High Priority -- Direct ops workflow impact

| API Section | Type | Endpoints | Use Case | Status |
|-------------|------|-----------|----------|--------|
| File System Replica Links | Resource | Full CRUD | FS-level replication between arrays (parity with bucket replica links) | Candidate |
| File System Snapshots | Resource | Full CRUD | Snapshot management + policy member association for backup/DR | Candidate |
| File System Policy Members | Resource | POST, DELETE `/file-systems/policies` | Associate NFS/SMB/snapshot policies to filesystems via TF | Candidate |
| Snapshot Policy FS Members | Resource | POST, DELETE `/snapshot-policies/file-systems` | Link snapshot policies to filesystems | Candidate |
| Object Store Roles | Resource | Full CRUD + trust policies | IAM-style roles for S3 fine-grained access | Candidate |
| Active Directory | Resource | Full CRUD | AD integration for SMB/NFS authentication | Candidate |

### Medium Priority -- Admin and security

| API Section | Type | Endpoints | Use Case | Status |
|-------------|------|-----------|----------|--------|
| Certificate Groups | Resource | POST, DELETE + members | Certificate grouping and rotation | Done |
| KMIP | Resource | Full CRUD | External encryption key management | Candidate |
| SAML2 SSO | Resource | Full CRUD | SAML-based single sign-on for admin console | Candidate |
| OIDC SSO | Resource | Full CRUD | OpenID Connect authentication | Candidate |
| SNMP Managers | Resource | Full CRUD | SNMP trap destinations for monitoring | Candidate |
| Administrators | Resource | Full CRUD + API tokens | Admin account management | Candidate |
| Alert Watchers | Resource | Full CRUD | Email alerting configuration | Candidate |
| Public Keys | Resource | GET, POST, DELETE | SSH/API public key management | Candidate |
| Password Policies | Resource | GET, PATCH | Admin password policy enforcement | Candidate |
| Keytabs | Resource | GET, POST, DELETE + upload | Kerberos keytab management | Candidate |
| Legal Holds | Resource | Full CRUD + held entities | Compliance / legal data retention | Candidate |
| CORS Policies | Resource | Full CRUD (bucket sub-resource) | S3 CORS header configuration | Candidate |
| Syslog Settings | Resource | GET, PATCH | Global syslog settings (separate from syslog servers) | Candidate |

### Low Priority -- Niche or rarely IaC-managed

| API Section | Type | Endpoints | Use Case | Status |
|-------------|------|-----------|----------|--------|
| Fleets | Resource | Full CRUD + members | Multi-array fleet management | Deferred |
| Realms | Resource | Full CRUD | Multi-tenancy domain isolation | Deferred |
| Node Groups | Resource | Full CRUD + nodes | FS placement on specific nodes | Deferred |
| Maintenance Windows | Resource | POST, DELETE | Scheduled maintenance (operational, not persistent) | Deferred |
| Software | Resource | GET, POST | Array upgrades -- dangerous via Terraform | Deferred |
| RDL (Rapid Data Locking) | Resource | GET, POST, PATCH | Crypto/compliance key rotation | Deferred |
| Storage Class Tiering Policies | Resource | Full CRUD | NVMe/SSD tiering -- rare use case | Deferred |
| WORM Data Policies | Resource | Full CRUD | Write-Once-Read-Many compliance | Deferred |

| Management Access Policies | Resource | Full CRUD | Admin console access control | Deferred |
| Management Auth Policies | Resource | Full CRUD | Admin authentication policies | Deferred |
| SSH CA Policies | Resource | Full CRUD | SSH certificate authority management | Deferred |
| Data Eviction Policies | Resource | Full CRUD + FS members | Automatic data eviction | Deferred |
| Audit Policies (FS) | Resource | Full CRUD + members | File system audit logging policies | Deferred |
| Audit Policies (FS) members | Resource | Members endpoint | File system audit policy-to-FS association | Deferred |
| Log Targets (File Systems) | Resource | Full CRUD | File system audit log target configuration | Deferred |

### Not Applicable for Terraform

| API Section | Reason |
|-------------|--------|
| Arrays (GET, PATCH) | Global array config -- too dangerous for IaC |
| Blades, Drives, Hardware, Hardware Connectors | Physical hardware -- no logical CRUD |
| Clients, Sessions | Monitoring/observability, not state management |
| Audits, Logs | Read-only telemetry |
| Usage/Performance metrics | Telemetry -- not IaC |
| Support, Support Diagnostics | Operational one-off actions |
| Roles | Built-in roles, read-only |
| Remote Arrays | Read-only (created via array connections) |
| Verification Keys | Crypto internals |
| SNMP Agents | Singleton GET/PATCH only |
| Nodes | Hardware topology, GET/PATCH only |

---

## How to Update This File

When implementing a new resource or data source:

1. Move the row from the "Not Implemented" section to the appropriate "Implemented" subsection
2. Fill in the Resource name, Data Source (Yes/No), and set Status to `Done`
3. Update the header counters (covered sections count, coverage percentage)
4. Update `Last updated` date and `Provider version`
5. Run `make docs` to regenerate Terraform documentation

When a new FlashBlade API version adds endpoints:

1. Add new rows to the appropriate priority section in "Not Implemented"
2. Update `Total API sections` count in the header
3. Note the minimum Purity//FB version required in the Notes column

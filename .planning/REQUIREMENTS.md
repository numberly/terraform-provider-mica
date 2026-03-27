# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-26
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Provider Foundation

- [x] **PROV-01**: Provider accepts endpoint URL, API token, and TLS CA certificate via config block
- [x] **PROV-02**: Provider accepts OAuth2 client_id, key_id, and issuer for client_credentials auth
- [x] **PROV-03**: Provider falls back to FLASHBLADE_ENDPOINT, FLASHBLADE_API_TOKEN, FLASHBLADE_OAUTH2_* environment variables when config block values are absent
- [x] **PROV-04**: Provider negotiates API version on startup via GET /api/api_version and targets v2.22
- [x] **PROV-05**: Provider marks api_token, oauth2 private key, and access key secrets as Sensitive in schema
- [x] **PROV-06**: Provider logs all operations with structured tflog output (resource name, operation, API path)
- [x] **PROV-07**: Provider supports custom CA certificate for TLS verification (ca_cert_file or inline ca_cert)

### File System

- [x] **FS-01**: User can create a file system with name, provisioned size, and optional policy attachments
- [x] **FS-02**: User can update file system attributes (size, policies, NFS settings, SMB settings)
- [x] **FS-03**: User can destroy a file system (two-phase: mark destroyed, then eradicate)
- [x] **FS-04**: User can read file system state including all computed attributes (space, created timestamp)
- [x] **FS-05**: User can import an existing file system into Terraform state by name
- [x] **FS-06**: Data source returns file system attributes by name or filter
- [x] **FS-07**: Drift detection logs field-level diffs via tflog when Read finds state divergence

### Object Store Account

- [x] **OSA-01**: User can create an object store account with name
- [x] **OSA-02**: User can update object store account attributes
- [x] **OSA-03**: User can destroy an object store account (two-phase soft-delete)
- [x] **OSA-04**: User can import an existing object store account into Terraform state by name
- [x] **OSA-05**: Data source returns object store account attributes by name or filter

### Bucket

- [x] **BKT-01**: User can create a bucket with name, account reference, and optional settings
- [x] **BKT-02**: User can update bucket attributes (quotas, versioning, policies)
- [x] **BKT-03**: User can destroy a bucket (two-phase: mark destroyed, then eradicate)
- [x] **BKT-04**: User can import an existing bucket into Terraform state by name
- [x] **BKT-05**: Data source returns bucket attributes by name or filter
- [x] **BKT-06**: Drift detection logs field-level diffs via tflog when Read finds state divergence

### Object Store Access Key

- [x] **OAK-01**: User can create an object store access key for a given account
- [x] **OAK-02**: User can delete an object store access key
- [x] **OAK-03**: Secret access key is marked Sensitive and only available at creation time
- [x] **OAK-04**: User can import an existing access key into Terraform state by name
- [x] **OAK-05**: Data source returns access key attributes by name or filter

### NFS Export Policy

- [x] **NFP-01**: User can create an NFS export policy with name and optional settings
- [x] **NFP-02**: User can update NFS export policy attributes
- [x] **NFP-03**: User can destroy an NFS export policy
- [ ] **NFP-04**: User can import an existing NFS export policy into Terraform state by name
- [ ] **NFP-05**: Data source returns NFS export policy attributes by name or filter
- [x] **NFR-01**: User can create NFS export policy rules (client, access, permissions)
- [x] **NFR-02**: User can update NFS export policy rules
- [x] **NFR-03**: User can destroy NFS export policy rules
- [ ] **NFR-04**: User can import NFS export policy rules using composite ID (policy_name:rule_index)

### SMB Share Policy

- [x] **SMP-01**: User can create an SMB share policy with name and optional settings
- [x] **SMP-02**: User can update SMB share policy attributes
- [x] **SMP-03**: User can destroy an SMB share policy
- [ ] **SMP-04**: User can import an existing SMB share policy into Terraform state by name
- [ ] **SMP-05**: Data source returns SMB share policy attributes by name or filter
- [x] **SMR-01**: User can create SMB share policy rules
- [x] **SMR-02**: User can update SMB share policy rules
- [x] **SMR-03**: User can destroy SMB share policy rules
- [ ] **SMR-04**: User can import SMB share policy rules using composite ID

### Snapshot Policy

- [x] **SNP-01**: User can create a snapshot policy with name and optional settings
- [x] **SNP-02**: User can update snapshot policy attributes
- [x] **SNP-03**: User can destroy a snapshot policy
- [ ] **SNP-04**: User can import an existing snapshot policy into Terraform state by name
- [ ] **SNP-05**: Data source returns snapshot policy attributes by name or filter
- [x] **SNR-01**: User can create snapshot policy rules (schedule, retention)
- [x] **SNR-02**: User can update snapshot policy rules
- [x] **SNR-03**: User can destroy snapshot policy rules
- [ ] **SNR-04**: User can import snapshot policy rules using composite ID

### Object Store Access Policy

- [ ] **OAP-01**: User can create an object store access policy with name and rules
- [ ] **OAP-02**: User can update object store access policy attributes
- [ ] **OAP-03**: User can destroy an object store access policy
- [ ] **OAP-04**: User can import an existing object store access policy into Terraform state by name
- [ ] **OAP-05**: Data source returns object store access policy attributes by name or filter
- [ ] **OAR-01**: User can create object store access policy rules (effect, action, resource, condition)
- [ ] **OAR-02**: User can update object store access policy rules
- [ ] **OAR-03**: User can destroy object store access policy rules
- [ ] **OAR-04**: User can import object store access policy rules using composite ID

### Network Access Policy

- [ ] **NAP-01**: User can create a network access policy with name
- [ ] **NAP-02**: User can update network access policy attributes
- [ ] **NAP-03**: User can destroy a network access policy
- [ ] **NAP-04**: User can import an existing network access policy into Terraform state by name
- [ ] **NAP-05**: Data source returns network access policy attributes by name or filter
- [ ] **NAR-01**: User can create network access policy rules (client, interfaces, effect)
- [ ] **NAR-02**: User can update network access policy rules
- [ ] **NAR-03**: User can destroy network access policy rules
- [ ] **NAR-04**: User can import network access policy rules using composite ID

### Quota Policy

- [ ] **QTP-01**: User can create a quota policy with name
- [ ] **QTP-02**: User can update quota policy attributes
- [ ] **QTP-03**: User can destroy a quota policy
- [ ] **QTP-04**: User can import an existing quota policy into Terraform state by name
- [ ] **QTP-05**: Data source returns quota policy attributes by name or filter
- [ ] **QTR-01**: User can create quota policy rules (quota_limit, enforced)
- [ ] **QTR-02**: User can update quota policy rules
- [ ] **QTR-03**: User can destroy quota policy rules
- [ ] **QTR-04**: User can import quota policy rules using composite ID

### Array Administration

- [ ] **ADM-01**: User can manage array DNS configuration (nameservers, domain, search)
- [ ] **ADM-02**: User can manage array NTP configuration (servers)
- [ ] **ADM-03**: User can manage array SMTP configuration (relay host, sender)
- [ ] **ADM-04**: Data sources for DNS, NTP, SMTP read-only access
- [ ] **ADM-05**: User can import existing array admin configuration into Terraform state

### Cross-Cutting Quality

- [ ] **QUA-01**: All resources implement correct plan modifiers (UseStateForUnknown for stable computed, RequiresReplace for immutable)
- [ ] **QUA-02**: All resources validate input at plan time (invalid quota values, enum fields, required references)
- [ ] **QUA-03**: Unit tests cover all schema definitions, validators, and plan modifiers
- [ ] **QUA-04**: Mocked API integration tests cover CRUD lifecycle for all resource families
- [ ] **QUA-05**: HTTP client implements retry with exponential backoff for transient API errors
- [ ] **QUA-06**: Provider documentation generated via terraform-plugin-docs for all resources and data sources

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Enhanced Storage

- **ESR-01**: Object Lock and WORM configuration on buckets
- **ESR-02**: QoS policy attachment on file systems and buckets
- **ESR-03**: Eradication config management (custom eradication delay)
- **ESR-04**: Bucket replica links for DR automation
- **ESR-05**: File system replica links for DR automation

### Extended Admin

- **EAD-01**: Array connection management (multi-array connectivity)
- **EAD-02**: API client management (service accounts)
- **EAD-03**: Active Directory integration
- **EAD-04**: Alert watchers and routing configuration

### Testing & Distribution

- **TDR-01**: Acceptance tests against real FlashBlade
- **TDR-02**: Terraform Registry publication (public)
- **TDR-03**: Pulumi bridge

## Out of Scope

| Feature | Reason |
|---------|--------|
| Performance metrics resources | Time-series data, not configuration state — use Prometheus/Datadog |
| Snapshot management as resources | Operational artifacts, not declarative infra — use snapshot policies |
| Multi-array in one provider block | Breaks state isolation — use provider aliases |
| Hardware management (blades, drives) | Read-only hardware state, cannot be declared |
| Audit log target resources | Circular dependencies with storage resources — manage separately |
| Session/client management | Ephemeral runtime state — operational runbooks, not IaC |
| Automatic resource name generation | Unstable state — user supplies names, use random_id if needed |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| PROV-01 | Phase 1 | Complete |
| PROV-02 | Phase 1 | Complete |
| PROV-03 | Phase 1 | Complete |
| PROV-04 | Phase 1 | Complete |
| PROV-05 | Phase 1 | Complete |
| PROV-06 | Phase 1 | Complete |
| PROV-07 | Phase 1 | Complete |
| FS-01 | Phase 1 | Complete |
| FS-02 | Phase 1 | Complete |
| FS-03 | Phase 1 | Complete |
| FS-04 | Phase 1 | Complete |
| FS-05 | Phase 1 | Complete |
| FS-06 | Phase 1 | Complete |
| FS-07 | Phase 1 | Complete |
| OSA-01 | Phase 2 | Complete |
| OSA-02 | Phase 2 | Complete |
| OSA-03 | Phase 2 | Complete |
| OSA-04 | Phase 2 | Complete |
| OSA-05 | Phase 2 | Complete |
| BKT-01 | Phase 2 | Complete |
| BKT-02 | Phase 2 | Complete |
| BKT-03 | Phase 2 | Complete |
| BKT-04 | Phase 2 | Complete |
| BKT-05 | Phase 2 | Complete |
| BKT-06 | Phase 2 | Complete |
| OAK-01 | Phase 2 | Complete |
| OAK-02 | Phase 2 | Complete |
| OAK-03 | Phase 2 | Complete |
| OAK-04 | Phase 2 | Complete |
| OAK-05 | Phase 2 | Complete |
| NFP-01 | Phase 3 | Complete |
| NFP-02 | Phase 3 | Complete |
| NFP-03 | Phase 3 | Complete |
| NFP-04 | Phase 3 | Pending |
| NFP-05 | Phase 3 | Pending |
| NFR-01 | Phase 3 | Complete |
| NFR-02 | Phase 3 | Complete |
| NFR-03 | Phase 3 | Complete |
| NFR-04 | Phase 3 | Pending |
| SMP-01 | Phase 3 | Complete |
| SMP-02 | Phase 3 | Complete |
| SMP-03 | Phase 3 | Complete |
| SMP-04 | Phase 3 | Pending |
| SMP-05 | Phase 3 | Pending |
| SMR-01 | Phase 3 | Complete |
| SMR-02 | Phase 3 | Complete |
| SMR-03 | Phase 3 | Complete |
| SMR-04 | Phase 3 | Pending |
| SNP-01 | Phase 3 | Complete |
| SNP-02 | Phase 3 | Complete |
| SNP-03 | Phase 3 | Complete |
| SNP-04 | Phase 3 | Pending |
| SNP-05 | Phase 3 | Pending |
| SNR-01 | Phase 3 | Complete |
| SNR-02 | Phase 3 | Complete |
| SNR-03 | Phase 3 | Complete |
| SNR-04 | Phase 3 | Pending |
| OAP-01 | Phase 4 | Pending |
| OAP-02 | Phase 4 | Pending |
| OAP-03 | Phase 4 | Pending |
| OAP-04 | Phase 4 | Pending |
| OAP-05 | Phase 4 | Pending |
| OAR-01 | Phase 4 | Pending |
| OAR-02 | Phase 4 | Pending |
| OAR-03 | Phase 4 | Pending |
| OAR-04 | Phase 4 | Pending |
| NAP-01 | Phase 4 | Pending |
| NAP-02 | Phase 4 | Pending |
| NAP-03 | Phase 4 | Pending |
| NAP-04 | Phase 4 | Pending |
| NAP-05 | Phase 4 | Pending |
| NAR-01 | Phase 4 | Pending |
| NAR-02 | Phase 4 | Pending |
| NAR-03 | Phase 4 | Pending |
| NAR-04 | Phase 4 | Pending |
| QTP-01 | Phase 4 | Pending |
| QTP-02 | Phase 4 | Pending |
| QTP-03 | Phase 4 | Pending |
| QTP-04 | Phase 4 | Pending |
| QTP-05 | Phase 4 | Pending |
| QTR-01 | Phase 4 | Pending |
| QTR-02 | Phase 4 | Pending |
| QTR-03 | Phase 4 | Pending |
| QTR-04 | Phase 4 | Pending |
| ADM-01 | Phase 4 | Pending |
| ADM-02 | Phase 4 | Pending |
| ADM-03 | Phase 4 | Pending |
| ADM-04 | Phase 4 | Pending |
| ADM-05 | Phase 4 | Pending |
| QUA-01 | Phase 5 | Pending |
| QUA-02 | Phase 5 | Pending |
| QUA-03 | Phase 5 | Pending |
| QUA-04 | Phase 5 | Pending |
| QUA-05 | Phase 5 | Pending |
| QUA-06 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 95 total (recount from requirement IDs — REQUIREMENTS.md header had 76 which was incorrect)
- Mapped to phases: 95
- Unmapped: 0

---
*Requirements defined: 2026-03-26*
*Last updated: 2026-03-26 — traceability populated after roadmap creation*

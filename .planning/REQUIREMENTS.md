# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-30
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v2.1 Requirements

Requirements for Bucket Advanced Features. Adds missing bucket sub-resources and inline config attributes from the FlashBlade REST API v2.22.

### Bucket Inline Attributes

- [ ] **BKT-01**: Bucket resource supports eradication_config (eradication_delay, eradication_mode, manual_eradication) on create and update
- [ ] **BKT-02**: Bucket resource supports object_lock_config (freeze_locked_objects, default_retention, default_retention_mode, object_lock_enabled) on create and update
- [ ] **BKT-03**: Bucket resource supports public_access_config (block_new_public_policies, block_public_access) on update
- [ ] **BKT-04**: Bucket resource exposes public_status as computed read-only attribute

### Lifecycle Rules

- [ ] **LCR-01**: Operator can create a lifecycle rule on a bucket with prefix, version retention, and multipart upload cleanup via Terraform
- [ ] **LCR-02**: Operator can update lifecycle rule settings (enabled, retention periods, prefix) via Terraform apply
- [ ] **LCR-03**: Operator can delete a lifecycle rule via Terraform destroy
- [ ] **LCR-04**: Operator can import an existing lifecycle rule into Terraform state with no drift on subsequent plan
- [ ] **LCR-05**: Lifecycle rule data source reads existing rules by bucket name

### Bucket Access Policies

- [ ] **BAP-01**: Operator can create a bucket access policy with rules (actions, effect, principals, resources) via Terraform
- [ ] **BAP-02**: Operator can delete a bucket access policy via Terraform destroy
- [ ] **BAP-03**: Operator can create/delete individual bucket access policy rules independently
- [ ] **BAP-04**: Operator can import existing bucket access policies into Terraform state
- [ ] **BAP-05**: Bucket access policy data source reads existing policy by bucket name

### Bucket Audit Filters

- [ ] **BAF-01**: Operator can create a bucket audit filter with actions and S3 prefix filtering via Terraform
- [ ] **BAF-02**: Operator can update audit filter settings (actions, s3_prefixes) via Terraform apply
- [ ] **BAF-03**: Operator can delete a bucket audit filter via Terraform destroy
- [ ] **BAF-04**: Operator can import an existing audit filter into Terraform state

### QoS Policies

- [ ] **QOS-01**: Operator can create a QoS policy with max_total_bytes_per_sec and max_total_ops_per_sec via Terraform
- [ ] **QOS-02**: Operator can update QoS policy limits via Terraform apply
- [ ] **QOS-03**: Operator can delete a QoS policy via Terraform destroy
- [ ] **QOS-04**: Operator can assign a QoS policy to buckets and file systems as members
- [ ] **QOS-05**: Operator can import existing QoS policies into Terraform state
- [ ] **QOS-06**: QoS policy data source reads existing policy by name

### Testing & Documentation

- [ ] **TST-01**: Unit tests for all new resources and bucket attribute additions (Read + NotFound + Lifecycle)
- [ ] **TST-02**: Mock handlers for all new API endpoints
- [ ] **DOC-01**: Import documentation for all new importable resources
- [ ] **DOC-02**: Workflow example showing bucket with lifecycle rules, access policy, audit filter, and QoS

## Future Requirements

### v2.2+

- **FUT-01**: Audit object store policies resource (CRUD + member assignment)
- **FUT-02**: Audit file systems policies resource (CRUD + member assignment + rules)
- **FUT-03**: CORS policies for buckets (if API supports it in future versions)
- **FUT-04**: Array connection resource (create/delete — currently data source only)
- **FUT-05**: File system replica links
- **FUT-06**: Cascading replication

## Out of Scope

| Feature | Reason |
|---------|--------|
| Audit policies (file system + object store) | Complex (log targets, rules, members) — defer to v2.2 |
| Array connection resource | Data source sufficient for replication — defer to v2.2 |
| File system replica links | Different API pattern than bucket links — defer to v2.2 |
| CORS policies | Not visible in API v2.22 |
| Pulumi bridge | Provider structure compatible but separate effort |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| BKT-01 | — | Pending |
| BKT-02 | — | Pending |
| BKT-03 | — | Pending |
| BKT-04 | — | Pending |
| LCR-01 | — | Pending |
| LCR-02 | — | Pending |
| LCR-03 | — | Pending |
| LCR-04 | — | Pending |
| LCR-05 | — | Pending |
| BAP-01 | — | Pending |
| BAP-02 | — | Pending |
| BAP-03 | — | Pending |
| BAP-04 | — | Pending |
| BAP-05 | — | Pending |
| BAF-01 | — | Pending |
| BAF-02 | — | Pending |
| BAF-03 | — | Pending |
| BAF-04 | — | Pending |
| QOS-01 | — | Pending |
| QOS-02 | — | Pending |
| QOS-03 | — | Pending |
| QOS-04 | — | Pending |
| QOS-05 | — | Pending |
| QOS-06 | — | Pending |
| TST-01 | — | Pending |
| TST-02 | — | Pending |
| DOC-01 | — | Pending |
| DOC-02 | — | Pending |

**Coverage:**
- v2.1 requirements: 28 total
- Mapped to phases: 0
- Unmapped: 28 ⚠️

---
*Requirements defined: 2026-03-30*
*Last updated: 2026-03-30 after v2.1 requirement definition*

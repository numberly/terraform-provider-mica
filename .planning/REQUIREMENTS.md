# Requirements: Terraform Provider FlashBlade

**Defined:** 2026-03-31
**Core Value:** Operational teams can reliably create, update, delete, and reconcile drift on FlashBlade storage resources through Terraform with zero surprises

## v2.1.1 Requirements

Requirements for Network Interfaces (VIPs). Adds subnet, network interface (VIP), and LAG resources/data sources to enable operators to manage FlashBlade networking infrastructure as code.

### Subnet

- [x] **SUB-01**: Operator can create a subnet with name, prefix, gateway, mtu, vlan, and link_aggregation_group via Terraform
- [x] **SUB-02**: Operator can update subnet settings (gateway, prefix, mtu, vlan, link_aggregation_group) via Terraform apply
- [x] **SUB-03**: Operator can delete a subnet via Terraform destroy
- [x] **SUB-04**: Operator can read any existing subnet by name via data source
- [x] **SUB-05**: Operator can import an existing subnet into Terraform state with no drift on subsequent plan
- [x] **SUB-06**: Drift detection logs changes when subnet is modified outside Terraform

### Network Interface (VIP)

- [x] **NI-01**: Operator can create a network interface with name, address, subnet, type, services, and attached_servers via Terraform
- [x] **NI-02**: Operator can update network interface settings (address, services, attached_servers) via Terraform apply
- [x] **NI-03**: Operator can delete a network interface via Terraform destroy
- [x] **NI-04**: subnet and type are immutable after creation (RequiresReplace)
- [x] **NI-05**: services accepts a single value from: data, sts, egress-only, replication
- [x] **NI-06**: attached_servers is required for data/sts services and forbidden for egress-only/replication services
- [x] **NI-07**: Operator can read an existing network interface by name via data source
- [x] **NI-08**: Operator can import an existing network interface into Terraform state with no drift on subsequent plan
- [x] **NI-09**: Drift detection logs changes when network interface is modified outside Terraform
- [x] **NI-10**: All read-only fields exposed as computed (enabled, gateway, mtu, netmask, vlan, realms)

### LAG (Link Aggregation Group)

- [x] **LAG-01**: Operator can read an existing LAG by name via data source (name, status, ports, port_speed, lag_speed, mac_address)

### Documentation & Workflow

- [ ] **DOC-01**: Import documentation (import.sh) exists for all new importable resources with correct syntax
- [ ] **DOC-02**: Workflow example in examples/networking/ demonstrates full LAG → subnet → VIP → server stack
- [ ] **DOC-03**: tfplugindocs generates documentation for all new resources and data sources without errors
- [ ] **DOC-04**: README coverage table includes networking resources category with correct counts

### Server Enrichment

- [ ] **SRV-01**: Server resource and data source expose associated VIPs as computed network_interfaces list
- [ ] **SRV-02**: Server schema version bumped from 0 to 1 with StateUpgrader migration

## v2.1 Requirements (completed)

Requirements for Bucket Advanced Features. Adds missing bucket sub-resources and inline config attributes from the FlashBlade REST API v2.22.

### Bucket Inline Attributes

- [x] **BKT-01**: Bucket resource supports eradication_config (eradication_delay, eradication_mode, manual_eradication) on create and update
- [x] **BKT-02**: Bucket resource supports object_lock_config (freeze_locked_objects, default_retention, default_retention_mode, object_lock_enabled) on create and update
- [x] **BKT-03**: Bucket resource supports public_access_config (block_new_public_policies, block_public_access) on update
- [x] **BKT-04**: Bucket resource exposes public_status as computed read-only attribute

### Lifecycle Rules

- [x] **LCR-01**: Operator can create a lifecycle rule on a bucket with prefix, version retention, and multipart upload cleanup via Terraform
- [x] **LCR-02**: Operator can update lifecycle rule settings (enabled, retention periods, prefix) via Terraform apply
- [x] **LCR-03**: Operator can delete a lifecycle rule via Terraform destroy
- [x] **LCR-04**: Operator can import an existing lifecycle rule into Terraform state with no drift on subsequent plan
- [x] **LCR-05**: Lifecycle rule data source reads existing rules by bucket name

### Bucket Access Policies

- [x] **BAP-01**: Operator can create a bucket access policy with rules (actions, effect, principals, resources) via Terraform
- [x] **BAP-02**: Operator can delete a bucket access policy via Terraform destroy
- [x] **BAP-03**: Operator can create/delete individual bucket access policy rules independently
- [x] **BAP-04**: Operator can import existing bucket access policies into Terraform state

### Bucket Audit Filters

- [x] **AUD-01**: Operator can create a bucket audit filter with actions and S3 prefix filtering via Terraform
- [x] **AUD-02**: Operator can update audit filter settings via Terraform apply
- [x] **AUD-03**: Operator can delete a bucket audit filter via Terraform destroy
- [x] **AUD-04**: Operator can import an existing bucket audit filter into Terraform state

### QoS Policies

- [x] **QOS-01**: Operator can create a QoS policy with bandwidth and IOPS limits via Terraform
- [x] **QOS-02**: Operator can update QoS policy settings via Terraform apply
- [x] **QOS-03**: Operator can delete a QoS policy via Terraform destroy
- [x] **QOS-04**: Operator can assign buckets and file systems as QoS policy members
- [x] **QOS-05**: Operator can import existing QoS policies and members into Terraform state

## Out of Scope

| Feature | Reason |
|---------|--------|
| Realms | Not relevant for current usage |
| Network interface connectors | Physical infrastructure, not managed via Terraform |
| TLS policies on network interfaces | Defer to future milestone |
| Network interface ping/trace | Diagnostic tools, not resource management |
| Pulumi bridge | Deferred, provider structure compatible |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SUB-01 | Phase 28 | Complete |
| SUB-02 | Phase 28 | Complete |
| SUB-03 | Phase 28 | Complete |
| SUB-04 | Phase 28 | Complete |
| SUB-05 | Phase 28 | Complete |
| SUB-06 | Phase 28 | Complete |
| NI-01 | Phase 29 | Complete |
| NI-02 | Phase 29 | Complete |
| NI-03 | Phase 29 | Complete |
| NI-04 | Phase 29 | Complete |
| NI-05 | Phase 29 | Complete |
| NI-06 | Phase 29 | Complete |
| NI-07 | Phase 29 | Complete |
| NI-08 | Phase 29 | Complete |
| NI-09 | Phase 29 | Complete |
| NI-10 | Phase 29 | Complete |
| LAG-01 | Phase 28 | Complete |
| SRV-01 | Phase 30 | Pending |
| SRV-02 | Phase 30 | Pending |
| DOC-01 | Phase 31 | Pending |
| DOC-02 | Phase 31 | Pending |
| DOC-03 | Phase 31 | Pending |
| DOC-04 | Phase 31 | Pending |

**Coverage:**
- v2.1.1 requirements: 23 total
- Mapped to phases: 23
- Unmapped: 0

---
*Requirements defined: 2026-03-31*
*Last updated: 2026-03-31 after milestone v2.1.1 definition*

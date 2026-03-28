---
phase: quick
plan: 1
type: execute
wave: 1
depends_on: []
files_modified:
  - examples/workflows/object-store-setup/main.tf
  - examples/workflows/nfs-file-share/main.tf
  - examples/workflows/multi-protocol-file-system/main.tf
  - examples/workflows/array-admin-baseline/main.tf
  - examples/workflows/secured-s3-bucket/main.tf
  - README.md
autonomous: true
requirements: ["QUICK-01"]

must_haves:
  truths:
    - "Ops engineer can copy-paste any workflow .tf file and adapt it for their environment"
    - "Each workflow shows resource composition with correct cross-resource references"
    - "Comments explain WHY each attribute is configured, not just WHAT it does"
    - "README links to the workflows directory for discoverability"
  artifacts:
    - path: "examples/workflows/object-store-setup/main.tf"
      provides: "Account -> bucket -> access key full S3 workflow"
      contains: "flashblade_object_store_account"
    - path: "examples/workflows/nfs-file-share/main.tf"
      provides: "File system + NFS export policy + rules"
      contains: "flashblade_nfs_export_policy_rule"
    - path: "examples/workflows/multi-protocol-file-system/main.tf"
      provides: "NFS + SMB on same FS with both policies"
      contains: "multi_protocol"
    - path: "examples/workflows/array-admin-baseline/main.tf"
      provides: "DNS + NTP + SMTP day-1 setup"
      contains: "flashblade_array_dns"
    - path: "examples/workflows/secured-s3-bucket/main.tf"
      provides: "Bucket + network access + OAP policy stack"
      contains: "flashblade_object_store_access_policy_rule"
  key_links:
    - from: "examples/workflows/object-store-setup/main.tf"
      to: "bucket references account"
      via: "account attribute referencing account resource name"
      pattern: "flashblade_object_store_account\\..+\\.name"
    - from: "examples/workflows/nfs-file-share/main.tf"
      to: "file system references NFS policy"
      via: "nfs_export_policy attribute"
      pattern: "nfs_export_policy.*=.*flashblade_nfs_export_policy"
---

<objective>
Create 5 production workflow examples showing how FlashBlade Terraform resources compose together in real ops team scenarios.

Purpose: Existing per-resource examples show isolated usage. Ops engineers need complete, copy-pasteable workflows showing how resources wire together for common tasks (S3 setup, NFS shares, multi-protocol, day-1 admin, secured buckets).

Output: 5 self-contained .tf files in examples/workflows/ plus README update for discoverability.
</objective>

<execution_context>
@/home/gule/.claude/get-shit-done/workflows/execute-plan.md
@/home/gule/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@examples/resources/flashblade_object_store_account/resource.tf
@examples/resources/flashblade_bucket/resource.tf
@examples/resources/flashblade_object_store_access_key/resource.tf
@examples/resources/flashblade_file_system/resource.tf
@examples/resources/flashblade_nfs_export_policy/resource.tf
@examples/resources/flashblade_nfs_export_policy_rule/resource.tf
@examples/resources/flashblade_smb_share_policy/resource.tf
@examples/resources/flashblade_smb_share_policy_rule/resource.tf
@examples/resources/flashblade_array_dns/resource.tf
@examples/resources/flashblade_array_ntp/resource.tf
@examples/resources/flashblade_array_smtp/resource.tf
@examples/resources/flashblade_network_access_policy/resource.tf
@examples/resources/flashblade_network_access_policy_rule/resource.tf
@examples/resources/flashblade_object_store_access_policy/resource.tf
@examples/resources/flashblade_object_store_access_policy_rule/resource.tf
@docs/resources/file_system.md
@README.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create all 5 workflow example files</name>
  <files>
    examples/workflows/object-store-setup/main.tf
    examples/workflows/nfs-file-share/main.tf
    examples/workflows/multi-protocol-file-system/main.tf
    examples/workflows/array-admin-baseline/main.tf
    examples/workflows/secured-s3-bucket/main.tf
  </files>
  <action>
Create the `examples/workflows/` directory structure and write 5 complete .tf files. Each file must:
- Start with a header comment block explaining the workflow scenario and what it provisions
- Include the provider block with variable-driven config (endpoint + api_token from vars)
- Use `variable` blocks for all environment-specific values (endpoint, token, CIDR ranges, email addresses, domain names)
- Use Terraform references between resources (not hardcoded names) to show proper composition
- Include inline comments explaining WHY each attribute value is set (ops context, not schema docs)

**Workflow 1: Object Store Setup** (`object-store-setup/main.tf`)
Complete S3-compatible storage workflow: account -> bucket (with versioning + 100 GiB quota, hard limit) -> access key -> outputs for key_id and secret.
- Account: `var.account_name`, quota_limit 1 TiB, hard_limit_enabled false (soft warn, not block)
- Bucket: references `flashblade_object_store_account.this.name`, versioning "enabled" (compliance/audit trail), quota_limit 100 GiB hard limit, destroy_eradicate_on_delete false (protect production data)
- Access key: references account name, enabled true
- Outputs: access_key_id (plain), secret_access_key (sensitive)

**Workflow 2: NFS File Share** (`nfs-file-share/main.tf`)
Team shared storage: file system + NFS export policy + 2 rules (app servers rw, backup servers ro).
- File system: 50 GiB provisioned, NFS enabled with v4_1_enabled, nfs_export_policy referencing the policy resource name, default_quotas with user_quota 5 GiB
- NFS export policy: enabled
- Rule 1: app servers subnet (`var.app_subnet`, default "10.10.0.0/16"), permission "rw", access "root-squash", security ["sys"] -- comment: root-squash prevents app containers running as root from having root on NFS
- Rule 2: backup subnet (`var.backup_subnet`, default "10.20.0.0/16"), permission "ro", access "root-squash", security ["sys"] -- comment: read-only for backup agents, they pull snapshots

**Workflow 3: Multi-Protocol File System** (`multi-protocol-file-system/main.tf`)
Windows + Linux access on same FS: file system with both NFS and SMB enabled, separate policies for each.
- File system: 100 GiB, NFS enabled (v3 + v4.1), SMB enabled (access_based_enumeration_enabled true, smb_encryption_enabled true), multi_protocol block (access_control_style "nfs", safeguard_acls true), nfs_export_policy and smb_share_policy referencing respective policy resources
- NFS export policy + rule: Linux subnet, rw, no-root-squash (trusted admin hosts), security ["sys", "krb5"]
- SMB share policy + rule: principal "Domain Users", read "allow", change "allow", full_control "deny" -- comment: change=allow lets users create/modify files, full_control=deny prevents ACL/ownership changes

**Workflow 4: Array Admin Baseline** (`array-admin-baseline/main.tf`)
Day-1 array setup: DNS + NTP + SMTP configuration.
- DNS: domain `var.domain` (default "corp.example.com"), nameservers from `var.dns_servers` (default ["10.0.0.53", "10.0.1.53"]) -- comment: internal DNS for forward+reverse resolution of array hostname
- NTP: ntp_servers from `var.ntp_servers` (default ["0.pool.ntp.org", "1.pool.ntp.org", "2.pool.ntp.org"]) -- comment: minimum 2 servers for redundancy, 3 preferred for quorum
- SMTP: relay_host `var.smtp_relay` (default "smtp.corp.example.com"), sender_domain `var.domain`, encryption_mode "tls" -- comment: TLS mandatory for PCI/SOC2 compliance
  - alert_watchers: ops-team email at warning level (day-to-day capacity/performance alerts), oncall email at error level (pages for hardware failures, space critical)

**Workflow 5: Secured S3 Bucket** (`secured-s3-bucket/main.tf`)
Bucket with full policy stack: account + bucket + network access policy + NAP rule + object store access policy + OAP rule.
- Account + Bucket: similar to workflow 1 but focused on security, no access key (keys managed separately)
- Network access policy: name "default" (singleton), enabled true -- comment: singleton on FlashBlade, Terraform adopts it
- NAP rule: client `var.allowed_cidr` (default "10.0.0.0/8"), effect "allow", interfaces ["s3"] -- comment: restrict S3 protocol to internal network only
- Object store access policy: name "app-readonly", description "Read-only S3 access for application tier"
- OAP rule: name "allow-bucket-read", effect "allow", actions ["s3:GetObject", "s3:ListBucket", "s3:GetBucketLocation"], resources referencing the bucket ARN pattern -- comment: least-privilege, no write/delete actions

All numeric byte values must use inline comments showing human-readable size (e.g., `# 50 GiB`).
  </action>
  <verify>
    <automated>find examples/workflows -name "main.tf" -type f | wc -l | grep -q "5" && echo "PASS: 5 workflow files created" || echo "FAIL"</automated>
  </verify>
  <done>5 workflow .tf files exist, each self-contained with provider block, variables, resources with cross-references, and ops-context comments</done>
</task>

<task type="auto">
  <name>Task 2: Add workflows section to README and validate HCL</name>
  <files>README.md</files>
  <action>
1. Edit README.md to add a "Workflow Examples" section after the "Data Sources" table and before "Development". Content:

```
## Workflow Examples

Production-ready configurations showing how resources compose together:

| Workflow | Description | Resources Used |
|----------|-------------|----------------|
| [Object Store Setup](examples/workflows/object-store-setup/) | S3-compatible storage: account, bucket, access key | account, bucket, access_key |
| [NFS File Share](examples/workflows/nfs-file-share/) | Team shared storage with export policy | file_system, nfs_export_policy, nfs_export_policy_rule |
| [Multi-Protocol File System](examples/workflows/multi-protocol-file-system/) | Windows + Linux access on same FS | file_system, nfs_export_policy, smb_share_policy |
| [Array Admin Baseline](examples/workflows/array-admin-baseline/) | Day-1 DNS, NTP, SMTP configuration | array_dns, array_ntp, array_smtp |
| [Secured S3 Bucket](examples/workflows/secured-s3-bucket/) | Bucket with network + access policies | bucket, network_access_policy, object_store_access_policy |
```

2. Run `terraform fmt -check -recursive examples/workflows/` to validate HCL formatting. Fix any formatting issues with `terraform fmt -recursive examples/workflows/`.

3. Run `terraform validate` in each workflow directory (init with `-backend=false` first) to catch syntax errors. If terraform binary is not available, at minimum verify HCL is parseable with `terraform fmt`.
  </action>
  <verify>
    <automated>grep -q "Workflow Examples" README.md && terraform fmt -check -recursive examples/workflows/ && echo "PASS" || echo "FAIL"</automated>
  </verify>
  <done>README contains workflow examples section with links, all .tf files pass terraform fmt validation</done>
</task>

</tasks>

<verification>
- All 5 workflow files exist in examples/workflows/{name}/main.tf
- Each file contains provider block, variable blocks, resource blocks with cross-references
- Each file has inline comments explaining WHY (not just WHAT)
- terraform fmt passes on all files
- README links to all 5 workflows
</verification>

<success_criteria>
- 5 complete, self-contained .tf workflow files in examples/workflows/
- Every resource reference uses Terraform expressions (no hardcoded names between resources)
- Comments provide ops context (security rationale, sizing reasoning, compliance notes)
- README updated with discoverable links to all workflows
- All HCL passes terraform fmt validation
</success_criteria>

<output>
After completion, create `.planning/quick/1-add-production-workflow-examples-to-docu/1-SUMMARY.md`
</output>

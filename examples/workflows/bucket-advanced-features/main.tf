# =============================================================================
# Workflow: Bucket Advanced Features
# =============================================================================
# Provisions a bucket with the complete v2.1 advanced features stack:
#   - Object store account     (namespace for the bucket)
#   - Bucket                   (versioning, eradication protection, object lock, quota)
#   - Lifecycle rule           (automated version cleanup + incomplete upload expiry)
#   - Bucket access policy     (IAM-style policy shell on the bucket)
#   - Access policy rule       (least-privilege read-only rule)
#   - Bucket audit filter      (S3 operation audit logging)
#   - QoS policy               (bandwidth + IOPS ceiling)
#   - QoS policy member        (assigns the QoS policy to the bucket)
#
# Typical use case: production data bucket where compliance requires audit
# logging, retention guarantees (object lock), automated lifecycle cleanup,
# access control, and performance isolation via QoS.
#
# Access keys are provisioned separately so this workflow focuses on the
# storage infrastructure and policy stack.
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 2.0"
    }
  }
}

provider "flashblade" {
  endpoint = var.flashblade_endpoint

  auth = {
    api_token = var.flashblade_api_token
  }
}

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------

variable "flashblade_endpoint" {
  type        = string
  description = "FlashBlade management endpoint URL."
}

variable "flashblade_api_token" {
  type        = string
  sensitive   = true
  description = "FlashBlade API token."
}

variable "account_name" {
  type        = string
  description = "Object store account name."
}

variable "bucket_name" {
  type        = string
  description = "S3 bucket name."
}

# ---------------------------------------------------------------------------
# Object store account
# ---------------------------------------------------------------------------

resource "flashblade_object_store_account" "this" {
  name = var.account_name

  # 1 TiB ceiling — production bucket with compliance data needs bounded capacity
  # to prevent runaway ingestion from consuming shared array space.
  quota_limit        = "1099511627776" # 1 TiB
  hard_limit_enabled = true
}

# ---------------------------------------------------------------------------
# Bucket — versioned with eradication protection and object lock
# ---------------------------------------------------------------------------

resource "flashblade_bucket" "this" {
  name    = var.bucket_name
  account = flashblade_object_store_account.this.name

  # Versioning retains all object versions — required for object lock
  # and protects against accidental overwrites.
  versioning = "enabled"

  # Eradication protection prevents permanent data loss from operator error.
  # 24-hour delay gives time to recover from accidental terraform destroy.
  eradication_config = {
    eradication_delay      = 86400000 # 24 hours in milliseconds
    eradication_mode       = "retention-based"
    manual_eradication     = "disabled"
  }

  # Object lock with 90-day default retention — compliance requirement for
  # audit data that must not be modified or deleted within the retention window.
  object_lock_config = {
    enabled                    = true
    default_retention_mode     = "compliance"
    default_retention_days     = 90
    freeze_locked_objects      = false
  }

  # 500 GiB quota with hard limit — prevents uncontrolled growth.
  quota_limit        = "536870912000" # 500 GiB
  hard_limit_enabled = true

  # Soft-delete only — eradication protection handles permanent deletion timing.
  destroy_eradicate_on_delete = false
}

# ---------------------------------------------------------------------------
# Lifecycle rule — automated version cleanup
# ---------------------------------------------------------------------------

resource "flashblade_lifecycle_rule" "cleanup" {
  bucket_name = flashblade_bucket.this.name
  enabled     = true

  # Apply to all objects in the bucket
  prefix = "/"

  # Delete previous versions after 30 days — keeps storage costs bounded
  # while still allowing recovery from recent accidental overwrites.
  keep_previous_version_for = 30

  # Abort incomplete multipart uploads after 7 days — prevents orphaned
  # upload parts from accumulating when clients crash mid-upload.
  abort_incomplete_multipart_uploads_after = 7
}

# ---------------------------------------------------------------------------
# Bucket access policy — IAM-style access control
# ---------------------------------------------------------------------------

resource "flashblade_bucket_access_policy" "this" {
  # Creates the policy shell on the bucket. Rules are managed separately
  # so teams can add/remove permissions without touching the policy itself.
  bucket_name = flashblade_bucket.this.name
}

# ---------------------------------------------------------------------------
# Access policy rule — least-privilege read-only
# ---------------------------------------------------------------------------

resource "flashblade_bucket_access_policy_rule" "read_only" {
  bucket_name = flashblade_bucket.this.name
  name        = "allow-read"

  # Read-only: GetObject for downloads, ListBucket for object discovery.
  # Write and delete operations require a separate policy rule.
  actions = ["s3:GetObject", "s3:ListBucket"]
  effect  = "allow"

  # Scope to this specific bucket and all objects within it.
  resources = [
    "arn:aws:s3:::${flashblade_bucket.this.name}",
    "arn:aws:s3:::${flashblade_bucket.this.name}/*",
  ]

  depends_on = [flashblade_bucket_access_policy.this]
}

# ---------------------------------------------------------------------------
# Audit filter — S3 operation logging for compliance
# ---------------------------------------------------------------------------

resource "flashblade_bucket_audit_filter" "this" {
  bucket_name = flashblade_bucket.this.name

  # Log read, write, and delete operations — the three actions compliance
  # teams need for access auditing and data lineage tracking.
  actions = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
}

# ---------------------------------------------------------------------------
# QoS policy — performance isolation
# ---------------------------------------------------------------------------

resource "flashblade_qos_policy" "standard" {
  name    = "${var.bucket_name}-qos"
  enabled = true

  # 1 GiB/s bandwidth ceiling — prevents a single bucket from saturating
  # the array's network capacity during large bulk transfers.
  max_total_bytes_per_sec = 1073741824

  # 10,000 IOPS ceiling — limits metadata-heavy workloads (listing, HEAD
  # requests) from impacting other tenants on shared arrays.
  max_total_ops_per_sec = 10000
}

# ---------------------------------------------------------------------------
# QoS policy member — assign QoS policy to the bucket
# ---------------------------------------------------------------------------

resource "flashblade_qos_policy_member" "bucket" {
  policy_name = flashblade_qos_policy.standard.name
  member_name = flashblade_bucket.this.name
  member_type = "buckets"
}

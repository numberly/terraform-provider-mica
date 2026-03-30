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
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 2.1"
    }
  }
}

provider "flashblade" {
  endpoint = var.flashblade_endpoint
  auth     = { api_token = var.flashblade_api_token }
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
  name               = var.account_name
  quota_limit        = 1099511627776 # 1 TiB
  hard_limit_enabled = true
}

# ---------------------------------------------------------------------------
# Bucket — versioned with eradication protection and object lock
# ---------------------------------------------------------------------------

resource "flashblade_bucket" "this" {
  name    = var.bucket_name
  account = flashblade_object_store_account.this.name

  versioning         = "enabled"
  quota_limit        = 536870912000 # 500 GiB
  hard_limit_enabled = true

  eradication_config = {
    eradication_delay  = 86400000        # 24 hours in ms
    eradication_mode   = "retention-based"
    manual_eradication = "disabled"
  }

  object_lock_config = {
    object_lock_enabled    = true
    default_retention_mode = "compliance"
    default_retention      = 7776000000  # 90 days in ms
    freeze_locked_objects  = false
  }

  destroy_eradicate_on_delete = false
}

# ---------------------------------------------------------------------------
# Lifecycle rule — automated version cleanup
# ---------------------------------------------------------------------------

resource "flashblade_lifecycle_rule" "cleanup" {
  bucket_name = flashblade_bucket.this.name
  rule_id     = "cleanup-old-versions"
  enabled     = true
  prefix      = ""

  # Delete previous versions after 30 days (in ms)
  keep_previous_version_for = 2592000000

  # Abort incomplete multipart uploads after 7 days (in ms)
  abort_incomplete_multipart_uploads_after = 604800000
}

# ---------------------------------------------------------------------------
# Bucket access policy + rule
# ---------------------------------------------------------------------------

resource "flashblade_bucket_access_policy" "this" {
  bucket_name = flashblade_bucket.this.name
}

# Note: effect is read-only (always "allow", set by the API).
# The principals format depends on your FlashBlade firmware version.
resource "flashblade_bucket_access_policy_rule" "read_only" {
  bucket_name = flashblade_bucket.this.name
  name        = "allowread"
  actions     = ["s3:GetObject", "s3:ListBucket"]
  principals  = ["${var.account_name}/admin"]
  resources   = ["*"]

  depends_on = [flashblade_bucket_access_policy.this]
}

# ---------------------------------------------------------------------------
# Audit filter — S3 operation logging for compliance
# ---------------------------------------------------------------------------

resource "flashblade_bucket_audit_filter" "this" {
  name        = "auditops"
  bucket_name = flashblade_bucket.this.name
  actions     = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
}

# ---------------------------------------------------------------------------
# QoS policy — performance isolation
# ---------------------------------------------------------------------------

resource "flashblade_qos_policy" "standard" {
  name                    = "${var.bucket_name}-qos"
  enabled                 = true
  max_total_bytes_per_sec = 1073741824  # 1 GiB/s
  max_total_ops_per_sec   = 10000       # 10k IOPS
}

# Note: QoS member_type only supports "file-systems" and "realms" on API v2.22.
# Bucket QoS assignment is managed via the bucket's qos_policy attribute instead.
# resource "flashblade_qos_policy_member" "bucket" {
#   policy_name = flashblade_qos_policy.standard.name
#   member_name = "my-filesystem"
#   member_type = "file-systems"
# }

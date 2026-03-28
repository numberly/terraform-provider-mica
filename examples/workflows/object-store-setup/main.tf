# =============================================================================
# Workflow: Object Store Setup
# =============================================================================
# Provisions a complete S3-compatible storage stack:
#   - Object store account  (namespace, quota ceiling)
#   - Bucket               (versioning + hard quota for compliance)
#   - Access key pair      (credentials for app or service account)
#
# Typical use case: onboarding a new application team onto FlashBlade S3.
# The account acts as a billing/quota boundary; the bucket is the S3 endpoint.
# Access keys are output as sensitive values and should be stored in a vault.
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "soulkyu/flashblade"
      version = "~> 1.0"
    }
  }
}

provider "flashblade" {
  endpoint = var.flashblade_endpoint

  auth {
    api_token = var.flashblade_api_token
  }
}

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------

variable "flashblade_endpoint" {
  type        = string
  description = "FlashBlade management endpoint URL (e.g. https://flashblade.corp.example.com)."
}

variable "flashblade_api_token" {
  type        = string
  sensitive   = true
  description = "FlashBlade API token. Use FLASHBLADE_API_TOKEN env var to avoid storing in state."
}

variable "account_name" {
  type        = string
  description = "Object store account name. Acts as the S3 namespace and quota boundary for a team or application."
}

variable "bucket_name" {
  type        = string
  description = "S3 bucket name. Must be globally unique within the account."
}

# ---------------------------------------------------------------------------
# Object store account
# ---------------------------------------------------------------------------

resource "flashblade_object_store_account" "this" {
  name = var.account_name

  # 1 TiB soft ceiling across all buckets in this account.
  # Soft limit = alert+warn, not block — gives time to react before apps hit a hard wall.
  quota_limit        = "1099511627776" # 1 TiB
  hard_limit_enabled = false
}

# ---------------------------------------------------------------------------
# Bucket
# ---------------------------------------------------------------------------

resource "flashblade_bucket" "this" {
  name    = var.bucket_name
  account = flashblade_object_store_account.this.name

  # Versioning keeps all object revisions — required for audit trails and
  # allows recovery from accidental deletes without a separate backup.
  versioning = "enabled"

  # 100 GiB hard limit on this specific bucket.
  # Hard limit prevents a runaway upload from consuming the full account quota.
  quota_limit        = "107374182400" # 100 GiB
  hard_limit_enabled = true

  # Production safety: soft-delete only on terraform destroy.
  # Eradication requires an explicit opt-in so accidental `terraform destroy`
  # does not permanently delete data. Eradicate manually from the array UI.
  destroy_eradicate_on_delete = false
}

# ---------------------------------------------------------------------------
# Access key
# ---------------------------------------------------------------------------

resource "flashblade_object_store_access_key" "this" {
  account = flashblade_object_store_account.this.name
  enabled = true

  # Access keys have no ImportState — the secret is only available at creation.
  # Store the outputs in Vault/AWS Secrets Manager immediately after apply.
}

# ---------------------------------------------------------------------------
# Outputs
# ---------------------------------------------------------------------------

output "access_key_id" {
  description = "S3 access key ID. Configure in your application's AWS SDK or S3 client."
  value       = flashblade_object_store_access_key.this.access_key_id
}

output "secret_access_key" {
  description = "S3 secret access key. Store in a secrets manager — this value is not recoverable after apply."
  value       = flashblade_object_store_access_key.this.secret_access_key
  sensitive   = true
}

output "bucket_endpoint" {
  description = "Logical S3 endpoint path for this bucket."
  value       = "${var.flashblade_endpoint}/${flashblade_bucket.this.name}"
}

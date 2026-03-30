# =============================================================================
# Workflow: Secured S3 Bucket
# =============================================================================
# Provisions a bucket with a full security policy stack:
#   - Object store account  (namespace + quota)
#   - Bucket                (versioning, no access key in this workflow)
#   - Network access policy (singleton, restricts S3 to internal network)
#   - NAP rule              (allow S3 from internal CIDR only)
#   - Object store access policy (IAM-style, least-privilege read-only)
#   - OAP rule              (GetObject + ListBucket for application tier)
#
# Typical use case: S3 bucket for an application that reads assets/artifacts.
# Access keys are provisioned separately (e.g. via a secrets rotation workflow)
# so this workflow focuses purely on the network and IAM policy stack.
#
# NOTE: flashblade_network_access_policy is a singleton — it already exists on
# the array. Terraform adopts it via GET+PATCH rather than creating a new one.
# Running destroy will set enabled=false on the singleton (not delete it).
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

variable "allowed_cidr" {
  type        = string
  default     = "10.0.0.0/8"
  description = "CIDR block allowed to reach the S3 protocol. Restricts S3 access to the internal network — external IPs cannot connect even with valid credentials."
}

# ---------------------------------------------------------------------------
# Object store account
# ---------------------------------------------------------------------------

resource "flashblade_object_store_account" "this" {
  name = var.account_name

  # 500 GiB ceiling — tighter than the default for a security-sensitive bucket.
  # Prevents a misconfigured upload loop from consuming unbounded capacity.
  quota_limit        = "536870912000" # 500 GiB
  hard_limit_enabled = false
}

# ---------------------------------------------------------------------------
# Bucket
# ---------------------------------------------------------------------------

resource "flashblade_bucket" "this" {
  name    = var.bucket_name
  account = flashblade_object_store_account.this.name

  # Versioning retains all object versions — protects against accidental overwrites
  # and ransomware (cannot overwrite without creating a versioned entry).
  versioning = "enabled"

  # Hard limit blocks writes once quota is reached.
  # Preferred for security buckets to prevent exfiltration via bucket-fill attack.
  quota_limit        = "107374182400" # 100 GiB
  hard_limit_enabled = true

  # Soft-delete only — production data requires explicit eradication opt-in.
  destroy_eradicate_on_delete = false
}

# ---------------------------------------------------------------------------
# Network access policy (singleton — Terraform adopts existing policy)
# ---------------------------------------------------------------------------

resource "flashblade_network_access_policy" "default" {
  # "default" is the singleton NAP name on FlashBlade.
  # This resource performs GET+PATCH to adopt it — no POST is issued.
  name    = "default"
  enabled = true
}

# ---------------------------------------------------------------------------
# NAP rule — restrict S3 to internal network only
# ---------------------------------------------------------------------------

resource "flashblade_network_access_policy_rule" "internal_only" {
  policy_name = flashblade_network_access_policy.default.name
  name        = "internal-s3-only"
  client      = var.allowed_cidr
  effect      = "allow"

  # interfaces = ["s3"] restricts this rule to S3 traffic only.
  # NFS/SMB remain governed by separate rules on the same policy.
  # Without this restriction, external IPs could attempt S3 connections
  # and rely solely on credential-based auth — defense in depth requires
  # network-layer controls as well.
  interfaces = ["s3"]
}

# ---------------------------------------------------------------------------
# Object store access policy (IAM-style)
# ---------------------------------------------------------------------------

resource "flashblade_object_store_access_policy" "app_readonly" {
  # FlashBlade requires the account-scoped format: account-name/policy-name
  name        = "${flashblade_object_store_account.this.name}/readonly"
  description = "Read-only S3 access for application tier"

  # description is a POST-only field — changing it requires destroy+recreate.
  # Keep it stable; rename the policy if the purpose changes.
}

# ---------------------------------------------------------------------------
# OAP rule — least-privilege read actions
# ---------------------------------------------------------------------------

resource "flashblade_object_store_access_policy_rule" "allow_bucket_read" {
  policy_name = flashblade_object_store_access_policy.app_readonly.name
  name        = "allow-bucket-read"

  # effect RequiresReplace — set to "allow" once and do not change.
  effect = "allow"

  # Least-privilege: only the three actions an app needs to read objects.
  # s3:GetObject         — download an object by key
  # s3:ListBucket        — list objects (required for pagination/discovery)
  # s3:GetBucketLocation — needed by some SDK region resolution logic
  #
  # Notably absent: s3:PutObject, s3:DeleteObject — application tier cannot
  # write or delete. Use a separate policy for upload workflows.
  actions = [
    "s3:GetObject",
    "s3:ListBucket",
    "s3:GetBucketLocation",
  ]

  # Scope the rule to this specific bucket and all objects within it.
  # The ARN pattern matches FlashBlade's IAM-compatible resource naming.
  # Using a reference ensures the rule follows the bucket name if it changes.
  resources = [
    "arn:aws:s3:::${flashblade_bucket.this.name}",
    "arn:aws:s3:::${flashblade_bucket.this.name}/*",
  ]
}

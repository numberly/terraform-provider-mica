# =============================================================================
# S3 Tenant Full-Stack Onboarding
# =============================================================================
#
# Complete workflow for onboarding a new S3 tenant on FlashBlade.
# This is the canonical example — every layer from server to bucket.
#
# Architecture:
#
#   [Server] (data source — pre-provisioned)
#       |
#   [Object Store Account] (S3 namespace per tenant)
#       |
#       +-- [Account Export] --> links account to server
#       |       |
#       |   [S3 Export Policy + Rule] --> transport-level access control
#       |
#       +-- [Access Policy + Rule] --> IAM-style S3 operation control
#       |
#       +-- [Access Key] --> credentials (stored in output or Vault)
#       |
#       +-- [Bucket] --> actual storage with versioning + quota
#
# Usage:
#   1. Copy terraform.tfvars.example to terraform.tfvars
#   2. Fill in your FlashBlade endpoint and server name
#   3. terraform init && terraform apply
#
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 1.0"
    }
  }
}

# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------

variable "flashblade_endpoint" {
  type        = string
  description = "FlashBlade management endpoint (e.g. https://flashblade.example.com)."
}

variable "flashblade_api_token" {
  type        = string
  sensitive   = true
  description = "FlashBlade API token for authentication."
}

variable "tenant_name" {
  type        = string
  description = "Name of the tenant being onboarded. Used for account, bucket, and policy naming."
}

variable "server_name" {
  type        = string
  description = "Name of the existing FlashBlade server to export the account to."
}

variable "bucket_quota_bytes" {
  type        = number
  description = "Bucket quota in bytes. Default: 1 TiB."
  default     = 1099511627776 # 1 TiB
}

# -----------------------------------------------------------------------------
# Provider
# -----------------------------------------------------------------------------

provider "flashblade" {
  endpoint             = var.flashblade_endpoint
  auth                 = { api_token = var.flashblade_api_token }
  insecure_skip_verify = true
}

# -----------------------------------------------------------------------------
# Step 1: Reference the existing server
# -----------------------------------------------------------------------------
# The server is pre-provisioned infrastructure. We reference it to export
# the account and make S3 reachable through this server's network path.

data "flashblade_server" "this" {
  name = var.server_name
}

# -----------------------------------------------------------------------------
# Step 2: Create object store account
# -----------------------------------------------------------------------------
# Each tenant gets its own account — isolation boundary for buckets and keys.
# skip_default_export prevents the auto-created _array_server export since
# we manage exports explicitly below.

resource "flashblade_object_store_account" "tenant" {
  name                = var.tenant_name
  skip_default_export = true
}

# -----------------------------------------------------------------------------
# Step 3: S3 export policy + account export
# -----------------------------------------------------------------------------
# Two layers here:
# - The S3 export policy controls WHICH S3 operations are allowed at the
#   transport level (think: firewall for S3 traffic)
# - The account export links the account to the server through this policy

resource "flashblade_s3_export_policy" "tenant" {
  name    = "${var.tenant_name}-s3-export"
  enabled = true
}

resource "flashblade_s3_export_policy_rule" "allow_s3" {
  policy_name = flashblade_s3_export_policy.tenant.name
  name        = "allows3"
  actions     = ["pure:S3Access"]
  effect      = "allow"
  resources   = ["*"]
}

resource "flashblade_object_store_account_export" "tenant" {
  account_name = flashblade_object_store_account.tenant.name
  server_name  = data.flashblade_server.this.name
  policy_name  = flashblade_s3_export_policy.tenant.name
  enabled      = true
}

# -----------------------------------------------------------------------------
# Step 4: IAM-style access policy
# -----------------------------------------------------------------------------
# Controls what S3 operations the tenant's user can perform. This is the
# fine-grained authorization layer (separate from the export policy which
# controls transport-level access).

resource "flashblade_object_store_access_policy" "tenant_rw" {
  name = "${flashblade_object_store_account.tenant.name}/rw"
}

resource "flashblade_object_store_access_policy_rule" "allow_all" {
  policy_name = flashblade_object_store_access_policy.tenant_rw.name
  name        = "allowrw"
  effect      = "allow"
  actions     = ["s3:*"]
  resources   = ["*"]
}

# -----------------------------------------------------------------------------
# Step 5: Generate access key
# -----------------------------------------------------------------------------
# The access key is immutable and the secret is only available at creation.
# In production, pipe this to Vault (see vault-s3-onboarding workflow).

resource "flashblade_object_store_access_key" "tenant" {
  object_store_account = flashblade_object_store_account.tenant.name
}

# -----------------------------------------------------------------------------
# Step 6: Create bucket
# -----------------------------------------------------------------------------
# The bucket belongs to the tenant's account. Versioning enabled by default
# for data protection. Quota enforces storage limits.

resource "flashblade_bucket" "tenant" {
  name               = "${var.tenant_name}-data"
  account            = flashblade_object_store_account.tenant.name
  versioning         = "enabled"
  quota_limit        = var.bucket_quota_bytes
  hard_limit_enabled = true

  # Keep recoverable on destroy. Set to true only when you're certain
  # the data can be permanently deleted.
  destroy_eradicate_on_delete = false
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------

output "account_name" {
  description = "The object store account name."
  value       = flashblade_object_store_account.tenant.name
}

output "bucket_name" {
  description = "The bucket name."
  value       = flashblade_bucket.tenant.name
}

output "access_key_id" {
  description = "The S3 access key ID. Use with secret_access_key to authenticate."
  value       = flashblade_object_store_access_key.tenant.access_key_id
}

output "secret_access_key" {
  description = "The S3 secret access key. Only available at creation time."
  value       = flashblade_object_store_access_key.tenant.secret_access_key
  sensitive   = true
}

output "server_name" {
  description = "The FlashBlade server the account is exported to."
  value       = data.flashblade_server.this.name
}

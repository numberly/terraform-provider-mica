# =============================================================================
# S3 Bucket Replication — Bidirectional Cross-Array
# =============================================================================
#
# Complete workflow for setting up bidirectional S3 bucket replication between
# two FlashBlade arrays. Both sides get mirrored infrastructure so that writes
# on either array replicate to the other.
#
# Architecture:
#
#   [FlashBlade Primary]                 [FlashBlade Secondary]
#       |                                     |
#   [Object Store Account]               [Object Store Account]
#       |                                     |
#       +-- [Bucket (versioning)]         +-- [Bucket (versioning)]
#       |                                     |
#       +-- [S3 Export Policy + Rule]     +-- [S3 Export Policy + Rule]
#       |                                     |
#       +-- [Account Export]              +-- [Account Export]
#       |                                     |
#       +-- [Access Key] ----secret----> [Remote Credentials]
#       |                                     |
#       +-- [Remote Credentials] <--secret-- [Access Key]
#       |                                     |
#       +-- [Bucket Replica Link] -------> (remote bucket)
#       |                                     |
#       (remote bucket) <------- [Bucket Replica Link]
#
# Prerequisites:
#   - Two FlashBlade arrays with an array connection already established
#   - REST API v2.22+ on both arrays
#   - Network connectivity between arrays on replication ports
#
# Usage:
#   1. Copy terraform.tfvars.example to terraform.tfvars
#   2. Fill in both endpoints, server names, and API tokens
#   3. terraform init && terraform apply
#
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 2.0"
    }
  }
}

# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------

variable "primary_endpoint" {
  type        = string
  description = "Primary FlashBlade management endpoint (e.g. https://fb-primary.example.com)."
}

variable "secondary_endpoint" {
  type        = string
  description = "Secondary FlashBlade management endpoint (e.g. https://fb-secondary.example.com)."
}

variable "primary_api_token" {
  type        = string
  sensitive   = true
  description = "API token for the primary FlashBlade."
}

variable "secondary_api_token" {
  type        = string
  sensitive   = true
  description = "API token for the secondary FlashBlade."
}

variable "server_name_primary" {
  type        = string
  description = "Name of the existing server on the primary FlashBlade."
}

variable "server_name_secondary" {
  type        = string
  description = "Name of the existing server on the secondary FlashBlade."
}

variable "tenant_name" {
  type        = string
  description = "Tenant name used for accounts, buckets, and policies on both arrays."
  default     = "replication-tenant"
}

# -----------------------------------------------------------------------------
# Providers — one per FlashBlade array
# -----------------------------------------------------------------------------

provider "flashblade" {
  alias                = "primary"
  endpoint             = var.primary_endpoint
  auth                 = { api_token = var.primary_api_token }
  insecure_skip_verify = true
}

provider "flashblade" {
  alias                = "secondary"
  endpoint             = var.secondary_endpoint
  auth                 = { api_token = var.secondary_api_token }
  insecure_skip_verify = true
}

# =============================================================================
# Step 1: Verify array connections (read-only)
# =============================================================================
# Confirm the two arrays can see each other before creating any resources.

data "flashblade_array_connection" "primary_to_secondary" {
  provider    = flashblade.primary
  remote_name = "secondary"
}

data "flashblade_array_connection" "secondary_to_primary" {
  provider    = flashblade.secondary
  remote_name = "primary"
}

# =============================================================================
# Step 2: Object store accounts on both arrays
# =============================================================================
# Each side gets the same tenant account for symmetry.

resource "flashblade_object_store_account" "primary" {
  provider            = flashblade.primary
  name                = var.tenant_name
  skip_default_export = true
}

resource "flashblade_object_store_account" "secondary" {
  provider            = flashblade.secondary
  name                = var.tenant_name
  skip_default_export = true
}

# =============================================================================
# Step 3: Buckets with versioning enabled (required for replication)
# =============================================================================

resource "flashblade_bucket" "primary" {
  provider   = flashblade.primary
  name       = "${var.tenant_name}-data"
  account    = flashblade_object_store_account.primary.name
  versioning = "enabled"

  destroy_eradicate_on_delete = false
}

resource "flashblade_bucket" "secondary" {
  provider   = flashblade.secondary
  name       = "${var.tenant_name}-data"
  account    = flashblade_object_store_account.secondary.name
  versioning = "enabled"

  destroy_eradicate_on_delete = false
}

# =============================================================================
# Step 4: S3 export policies + rules + account exports on both arrays
# =============================================================================
# Each array needs an export policy allowing S3 access, then the account is
# exported to the server through that policy.

resource "flashblade_s3_export_policy" "primary" {
  provider = flashblade.primary
  name     = "${var.tenant_name}-s3-export"
  enabled  = true
}

resource "flashblade_s3_export_policy_rule" "primary" {
  provider    = flashblade.primary
  policy_name = flashblade_s3_export_policy.primary.name
  name        = "allows3"
  actions     = ["pure:S3Access"]
  effect      = "allow"
  resources   = ["*"]
}

resource "flashblade_object_store_account_export" "primary" {
  provider     = flashblade.primary
  account_name = flashblade_object_store_account.primary.name
  server_name  = var.server_name_primary
  policy_name  = flashblade_s3_export_policy.primary.name
  enabled      = true
}

resource "flashblade_s3_export_policy" "secondary" {
  provider = flashblade.secondary
  name     = "${var.tenant_name}-s3-export"
  enabled  = true
}

resource "flashblade_s3_export_policy_rule" "secondary" {
  provider    = flashblade.secondary
  policy_name = flashblade_s3_export_policy.secondary.name
  name        = "allows3"
  actions     = ["pure:S3Access"]
  effect      = "allow"
  resources   = ["*"]
}

resource "flashblade_object_store_account_export" "secondary" {
  provider     = flashblade.secondary
  account_name = flashblade_object_store_account.secondary.name
  server_name  = var.server_name_secondary
  policy_name  = flashblade_s3_export_policy.secondary.name
  enabled      = true
}

# =============================================================================
# Step 5: Access keys on both arrays
# =============================================================================
# Each array generates an access key. The secret from one side is used as
# remote credentials on the other side.

resource "flashblade_object_store_access_key" "primary" {
  provider             = flashblade.primary
  object_store_account = flashblade_object_store_account.primary.name
}

resource "flashblade_object_store_access_key" "secondary" {
  provider             = flashblade.secondary
  object_store_account = flashblade_object_store_account.secondary.name

  # Cross-array replication: share the same name and secret as the primary key.
  # The API requires both name and secret_access_key when providing an explicit secret.
  name              = flashblade_object_store_access_key.primary.name
  secret_access_key = flashblade_object_store_access_key.primary.secret_access_key
}

# =============================================================================
# Step 6: Remote credentials — each side stores the other's key
# =============================================================================
# Primary stores secondary's access key to authenticate replication writes.
# Secondary stores primary's access key for the reverse direction.

# Name must be formatted as <remote-name>/<credentials-name>
resource "flashblade_object_store_remote_credentials" "primary_to_secondary" {
  provider          = flashblade.primary
  name              = "secondary/secondary-creds"
  access_key_id     = flashblade_object_store_access_key.secondary.access_key_id
  secret_access_key = flashblade_object_store_access_key.secondary.secret_access_key
  remote_name       = "secondary"
}

resource "flashblade_object_store_remote_credentials" "secondary_to_primary" {
  provider          = flashblade.secondary
  name              = "primary/primary-creds"
  access_key_id     = flashblade_object_store_access_key.primary.access_key_id
  secret_access_key = flashblade_object_store_access_key.primary.secret_access_key
  remote_name       = "primary"
}

# =============================================================================
# Step 7: Bucket replica links — bidirectional replication
# =============================================================================
# Each array replicates its local bucket to the remote bucket on the other
# array. This creates a fully bidirectional replication topology.

resource "flashblade_bucket_replica_link" "primary_to_secondary" {
  provider                = flashblade.primary
  local_bucket_name       = flashblade_bucket.primary.name
  remote_bucket_name      = flashblade_bucket.secondary.name
  remote_credentials_name = flashblade_object_store_remote_credentials.primary_to_secondary.name
}

resource "flashblade_bucket_replica_link" "secondary_to_primary" {
  provider                = flashblade.secondary
  local_bucket_name       = flashblade_bucket.secondary.name
  remote_bucket_name      = flashblade_bucket.primary.name
  remote_credentials_name = flashblade_object_store_remote_credentials.secondary_to_primary.name
}

# =============================================================================
# Outputs
# =============================================================================

output "primary_access_key_id" {
  description = "S3 access key ID on the primary array."
  value       = flashblade_object_store_access_key.primary.access_key_id
}

output "secondary_access_key_id" {
  description = "S3 access key ID on the secondary array."
  value       = flashblade_object_store_access_key.secondary.access_key_id
}

output "primary_bucket_name" {
  description = "Bucket name on the primary array."
  value       = flashblade_bucket.primary.name
}

output "secondary_bucket_name" {
  description = "Bucket name on the secondary array."
  value       = flashblade_bucket.secondary.name
}

output "primary_replication_status" {
  description = "Replication status from primary to secondary."
  value       = flashblade_bucket_replica_link.primary_to_secondary.status
}

output "secondary_replication_status" {
  description = "Replication status from secondary to primary."
  value       = flashblade_bucket_replica_link.secondary_to_primary.status
}

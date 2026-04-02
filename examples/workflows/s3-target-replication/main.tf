# =============================================================================
# S3 Target Replication — FlashBlade to External S3 Endpoint
# =============================================================================
#
# Complete workflow for setting up S3 bucket replication from a FlashBlade
# to an external S3 endpoint (e.g. AWS S3, MinIO, or any S3-compatible store).
#
# Architecture:
#
#   [FlashBlade]
#       |
#   [Object Store Account]
#       |
#       +-- [Bucket (versioning)]
#       |
#       +-- [Target (external S3 address)]
#       |
#       +-- [Remote Credentials (target_name → access key + secret)]
#       |
#       +-- [Bucket Replica Link] ------> [External S3 Bucket]
#
# Prerequisites:
#   - FlashBlade with REST API v2.22+ (Purity//FB 4.6.7+)
#   - Network connectivity from FlashBlade to the external S3 endpoint
#   - Valid S3 credentials (access key ID + secret access key) for the remote
#
# Usage:
#   1. Copy terraform.tfvars.example to terraform.tfvars
#   2. Fill in endpoint, api_token, s3_target_address, and S3 credentials
#   3. terraform init && terraform apply
#
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 2.2"
    }
  }
}

# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------

variable "endpoint" {
  type        = string
  description = "FlashBlade management endpoint (e.g. https://fb.example.com)."
}

variable "api_token" {
  type        = string
  sensitive   = true
  description = "API token for the FlashBlade."
}

variable "tenant_name" {
  type        = string
  description = "Tenant name used for the object store account and bucket."
  default     = "s3-target-tenant"
}

variable "s3_target_address" {
  type        = string
  description = "External S3 endpoint address, e.g. s3.us-east-1.amazonaws.com."
}

variable "s3_access_key_id" {
  type        = string
  sensitive   = true
  description = "Access key ID for the external S3 endpoint."
}

variable "s3_secret_access_key" {
  type        = string
  sensitive   = true
  description = "Secret access key for the external S3 endpoint."
}

# -----------------------------------------------------------------------------
# Provider
# -----------------------------------------------------------------------------

provider "flashblade" {
  endpoint             = var.endpoint
  auth                 = { api_token = var.api_token }
  insecure_skip_verify = true
}

# =============================================================================
# Step 1: Object store account
# =============================================================================
# Create the tenant account that will own the replicated bucket.

resource "flashblade_object_store_account" "tenant" {
  name                = var.tenant_name
  skip_default_export = true
}

# =============================================================================
# Step 2: Bucket with versioning
# =============================================================================
# Versioning must be enabled on the local bucket for replication to work.

resource "flashblade_bucket" "replica" {
  name       = "${var.tenant_name}-data"
  account    = flashblade_object_store_account.tenant.name
  versioning = "enabled"

  destroy_eradicate_on_delete = false
}

# =============================================================================
# Step 3: S3 replication target
# =============================================================================
# A target represents an external S3 endpoint. The name is used as a namespace
# prefix for remote credentials associated with this target.
#
# Set ca_certificate_group to reference a certificate group already present on
# the array when the S3-compatible endpoint uses a custom or self-signed TLS
# certificate.

resource "flashblade_target" "s3_endpoint" {
  name    = "external-s3"
  address = var.s3_target_address
}

# =============================================================================
# Step 4: Remote credentials referencing the target
# =============================================================================
# Name must be formatted as <target-name>/<credentials-name>.
# Use target_name (not remote_name) when the remote is an external S3 endpoint
# rather than another FlashBlade array. target_name and remote_name are
# mutually exclusive.

resource "flashblade_object_store_remote_credentials" "s3_creds" {
  name              = "external-s3/s3-replication-creds"
  access_key_id     = var.s3_access_key_id
  secret_access_key = var.s3_secret_access_key
  target_name       = flashblade_target.s3_endpoint.name
}

# =============================================================================
# Step 5: Bucket replica link to external S3
# =============================================================================
# remote_bucket_name is the bucket name on the external S3 endpoint. It must
# already exist or be created out-of-band on the remote side before replication
# can succeed.

resource "flashblade_bucket_replica_link" "to_s3" {
  local_bucket_name       = flashblade_bucket.replica.name
  remote_bucket_name      = "${var.tenant_name}-data"
  remote_credentials_name = flashblade_object_store_remote_credentials.s3_creds.name
  paused                  = false
}

# =============================================================================
# Outputs
# =============================================================================

output "target_name" {
  description = "Name of the replication target."
  value       = flashblade_target.s3_endpoint.name
}

output "target_address" {
  description = "Address of the external S3 endpoint."
  value       = flashblade_target.s3_endpoint.address
}

output "target_status" {
  description = "Connectivity status of the replication target."
  value       = flashblade_target.s3_endpoint.status
}

output "replication_status" {
  description = "Replication status of the bucket replica link."
  value       = flashblade_bucket_replica_link.to_s3.status
}

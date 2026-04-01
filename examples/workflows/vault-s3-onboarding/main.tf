# =============================================================================
# Vault-Integrated S3 Onboarding Workflow
# =============================================================================
#
# Production workflow for onboarding a new S3 tenant on FlashBlade.
# Zero secrets in code — Vault owns credentials end-to-end.
#
# Architecture:
#
#   [Vault] --> FlashBlade API token
#       |
#   [Server] (data source — pre-provisioned)
#       |
#   [Object Store Account]
#       +-- [Account Export] + [S3 Export Policy]
#       +-- [Access Policy + Rule]
#       +-- [Named User] + [User Policy] --> per-tenant user with policy
#       +-- [Access Key] --> bound to named user, written back to Vault
#       +-- [Bucket]
#       |
#   [Vault] <-- S3 credentials stored for apps
#
# Security model:
#   - FlashBlade API token: read from Vault at apply time
#   - S3 access key: written to Vault immediately after creation
#   - Terraform outputs: endpoint only, zero secrets
#   - Applications: read S3 credentials from Vault, never from Terraform
#
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 2.1"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "~> 4.0"
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

variable "vault_secret_path" {
  type        = string
  description = "Vault KV path where the FlashBlade API token is stored."
  default     = "secret/infrastructure/flashblade"
}

variable "vault_flashblade_key" {
  type        = string
  description = "Key within the Vault secret that holds the FlashBlade API token."
  default     = "api_token"
}

variable "tenant_name" {
  type        = string
  description = "Name of the tenant being onboarded."
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

variable "s3_data_vip" {
  type        = string
  description = "FlashBlade S3 data VIP or FQDN for the bucket endpoint output."
}

# -----------------------------------------------------------------------------
# Step 1: Retrieve FlashBlade API key from Vault
# -----------------------------------------------------------------------------

data "vault_kv_secret_v2" "flashblade" {
  mount = split("/", var.vault_secret_path)[0]
  name  = join("/", slice(split("/", var.vault_secret_path), 1, length(split("/", var.vault_secret_path))))
}

provider "flashblade" {
  endpoint             = var.flashblade_endpoint
  auth                 = { api_token = data.vault_kv_secret_v2.flashblade.data[var.vault_flashblade_key] }
  insecure_skip_verify = true
}

# -----------------------------------------------------------------------------
# Step 2: Reference the existing server
# -----------------------------------------------------------------------------

data "flashblade_server" "this" {
  name = var.server_name
}

# -----------------------------------------------------------------------------
# Step 3: Create object store account
# -----------------------------------------------------------------------------

resource "flashblade_object_store_account" "tenant" {
  name                = var.tenant_name
  skip_default_export = true
}

# -----------------------------------------------------------------------------
# Step 4: S3 export policy + account export
# -----------------------------------------------------------------------------

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
# Step 5: IAM-style access policy
# -----------------------------------------------------------------------------

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
# Step 6: Named S3 user + policy association
# -----------------------------------------------------------------------------

resource "flashblade_object_store_user" "tenant" {
  name = "${flashblade_object_store_account.tenant.name}/${var.tenant_name}-user"
}

resource "flashblade_object_store_user_policy" "tenant_rw" {
  user_name   = flashblade_object_store_user.tenant.name
  policy_name = flashblade_object_store_access_policy.tenant_rw.name
}

# -----------------------------------------------------------------------------
# Step 7: Generate access key (bound to named user) + store in Vault
# -----------------------------------------------------------------------------

resource "flashblade_object_store_access_key" "tenant" {
  object_store_account = flashblade_object_store_account.tenant.name
  user                 = flashblade_object_store_user.tenant.name
}

resource "vault_kv_secret_v2" "tenant_s3_credentials" {
  mount = split("/", var.vault_secret_path)[0]
  name  = "tenants/${var.tenant_name}/s3"

  data_json = jsonencode({
    access_key_id     = flashblade_object_store_access_key.tenant.access_key_id
    secret_access_key = flashblade_object_store_access_key.tenant.secret_access_key
    endpoint          = "https://${var.s3_data_vip}"
    bucket            = flashblade_bucket.tenant.name
  })

  lifecycle {
    replace_triggered_by = [flashblade_object_store_access_key.tenant]
  }
}

# -----------------------------------------------------------------------------
# Step 8: Create bucket
# -----------------------------------------------------------------------------

resource "flashblade_bucket" "tenant" {
  name               = "${var.tenant_name}-data"
  account            = flashblade_object_store_account.tenant.name
  versioning         = "enabled"
  quota_limit        = var.bucket_quota_bytes
  hard_limit_enabled = true

  destroy_eradicate_on_delete = false
}

# -----------------------------------------------------------------------------
# Outputs: endpoint only (no secrets)
# -----------------------------------------------------------------------------

output "bucket_endpoint" {
  description = "S3 endpoint for the tenant's bucket. Credentials are in Vault."
  value       = "https://${var.s3_data_vip}/${flashblade_bucket.tenant.name}"
}

output "vault_credentials_path" {
  description = "Vault path where S3 credentials are stored."
  value       = "tenants/${var.tenant_name}/s3"
}

output "server_name" {
  description = "The FlashBlade server the account is exported to."
  value       = data.flashblade_server.this.name
}

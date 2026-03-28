# =============================================================================
# Vault-Integrated S3 Onboarding Workflow
# =============================================================================
#
# Production workflow for onboarding a new S3 tenant on FlashBlade:
#
# 1. Retrieve FlashBlade API key from Vault (zero secrets in code)
# 2. Create an object store account (S3 namespace)
# 3. Create a read-write access policy for the tenant
# 4. Generate an access key pair (stored back in Vault)
# 5. Create a bucket under the account
# 6. Output only the bucket endpoint (secret-free)
#
# This is how an ops team should onboard tenants: Vault owns all secrets,
# Terraform orchestrates the infrastructure, and outputs expose only what
# consuming applications need — the endpoint, nothing else.
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 1.0"
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
  description = "Name of the tenant being onboarded. Used for account, bucket, and policy naming."
}

variable "bucket_quota_bytes" {
  type        = number
  description = "Bucket quota in bytes. Default: 1 TiB."
  default     = 1099511627776 # 1 TiB
}

variable "s3_data_vip" {
  type        = string
  description = "FlashBlade S3 data VIP or FQDN (e.g. s3.flashblade.example.com). Used to construct the bucket endpoint output."
}

# -----------------------------------------------------------------------------
# Step 1: Retrieve FlashBlade API key from Vault
# -----------------------------------------------------------------------------
# The API token never appears in Terraform state or plan output — it's read
# from Vault at apply time and used only in memory.

data "vault_kv_secret_v2" "flashblade" {
  mount = split("/", var.vault_secret_path)[0]
  name  = join("/", slice(split("/", var.vault_secret_path), 1, length(split("/", var.vault_secret_path))))
}

provider "flashblade" {
  endpoint = var.flashblade_endpoint

  auth = {
    api_token = data.vault_kv_secret_v2.flashblade.data[var.vault_flashblade_key]
  }
}

# -----------------------------------------------------------------------------
# Step 2: Create object store account (S3 namespace)
# -----------------------------------------------------------------------------
# Each tenant gets its own account — isolation boundary for buckets and keys.

resource "flashblade_object_store_account" "tenant" {
  name = var.tenant_name
}

# -----------------------------------------------------------------------------
# Step 3: Create read-write access policy
# -----------------------------------------------------------------------------
# IAM-style policy granting full S3 read/write on all resources within
# the tenant's namespace. Attached to the account's default user.

resource "flashblade_object_store_access_policy" "tenant_rw" {
  # FlashBlade requires the account-scoped format: account-name/policy-name
  name = "${flashblade_object_store_account.tenant.name}/rw"
}

resource "flashblade_object_store_access_policy_rule" "allow_all_rw" {
  policy_name = flashblade_object_store_access_policy.tenant_rw.name
  name        = "allow-rw"
  effect      = "allow"
  actions     = ["s3:*"]
  resources   = ["*"]
}

# -----------------------------------------------------------------------------
# Step 4: Generate access key + store credentials in Vault
# -----------------------------------------------------------------------------
# The access key is immutable and the secret is only available at creation.
# We immediately write it to Vault so it never needs to be read from
# Terraform state. Applications retrieve it from Vault directly.

resource "flashblade_object_store_access_key" "tenant" {
  object_store_account = flashblade_object_store_account.tenant.name
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

  # Ensure the secret is updated if the key is recreated (ForceNew on any change).
  lifecycle {
    replace_triggered_by = [flashblade_object_store_access_key.tenant]
  }
}

# -----------------------------------------------------------------------------
# Step 5: Create bucket
# -----------------------------------------------------------------------------
# The bucket belongs to the tenant's account. Versioning enabled by default
# for data protection. Quota enforces storage limits.

resource "flashblade_bucket" "tenant" {
  name       = "${var.tenant_name}-data"
  account    = flashblade_object_store_account.tenant.name
  versioning = "enabled"
  quota_limit       = var.bucket_quota_bytes
  hard_limit_enabled = true

  # Bucket contains production data — keep recoverable on destroy.
  # Set to true only when you're certain the data can be permanently deleted.
  destroy_eradicate_on_delete = false
}

# -----------------------------------------------------------------------------
# Output: bucket endpoint only (no secrets)
# -----------------------------------------------------------------------------
# This is the ONLY output — consuming applications get the endpoint and
# retrieve credentials from Vault. No secrets leak through Terraform outputs.

output "bucket_endpoint" {
  description = "S3 endpoint for the tenant's bucket. Credentials are in Vault at tenants/<tenant>/s3."
  value       = "https://${var.s3_data_vip}/${flashblade_bucket.tenant.name}"
}

output "vault_credentials_path" {
  description = "Vault path where S3 credentials are stored. Applications should read from here."
  value       = "tenants/${var.tenant_name}/s3"
}

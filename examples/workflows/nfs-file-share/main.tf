# =============================================================================
# Workflow: NFS File Share
# =============================================================================
# Provisions a team shared storage volume with export policy:
#   - File system  (50 GiB provisioned, NFSv4.1, per-user quota)
#   - NFS export policy
#   - Rule 1: app servers — read-write with root-squash
#   - Rule 2: backup agents — read-only with root-squash
#
# Typical use case: shared dataset for a batch processing team where multiple
# application pods and a backup agent need differentiated NFS access.
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/flashblade"
      version = "~> 1.0"
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

variable "filesystem_name" {
  type        = string
  description = "Name of the file system to create."
}

variable "app_subnet" {
  type        = string
  default     = "10.10.0.0/16"
  description = "CIDR range for application servers that need read-write NFS access."
}

variable "backup_subnet" {
  type        = string
  default     = "10.20.0.0/16"
  description = "CIDR range for backup agents that need read-only NFS access."
}

# ---------------------------------------------------------------------------
# NFS export policy
# ---------------------------------------------------------------------------

resource "flashblade_nfs_export_policy" "this" {
  name    = "${var.filesystem_name}-nfs"
  enabled = true
}

# ---------------------------------------------------------------------------
# Rule 1: Application servers — read-write
# ---------------------------------------------------------------------------

resource "flashblade_nfs_export_policy_rule" "app_rw" {
  policy_name = flashblade_nfs_export_policy.this.name
  client      = var.app_subnet
  permission  = "rw"

  # root-squash maps UID 0 on the client to the anonymous UID on the array.
  # App containers often run as root; squashing prevents root on NFS from
  # bypassing POSIX permissions set by other users on shared files.
  access = "root-squash"

  # sys = standard UNIX AUTH_SYS security. Adequate when network access is
  # already restricted to a trusted subnet (e.g. pod network).
  security = ["sys"]
}

# ---------------------------------------------------------------------------
# Rule 2: Backup agents — read-only
# ---------------------------------------------------------------------------

resource "flashblade_nfs_export_policy_rule" "backup_ro" {
  policy_name = flashblade_nfs_export_policy.this.name
  client      = var.backup_subnet
  permission  = "ro"

  # Backup agents pull snapshots — read-only ensures they cannot modify live data
  # even if the backup host is compromised.
  access   = "root-squash"
  security = ["sys"]
}

# ---------------------------------------------------------------------------
# File system
# ---------------------------------------------------------------------------

resource "flashblade_file_system" "this" {
  name        = var.filesystem_name
  provisioned = 53687091200 # 50 GiB

  nfs {
    enabled = true

    # v4.1 required for per-client stateful mounts, locks, and delegations.
    # Most modern Linux kernels (3.1+) support it. Disable v3 to avoid
    # insecure stateless mounts from legacy clients sneaking in.
    v4_1_enabled = true
    v3_enabled   = false
  }

  # Attach the export policy — file system will immediately enforce the rules above.
  nfs_export_policy = flashblade_nfs_export_policy.this.name

  default_quotas {
    # 5 GiB soft quota per user prevents a single user from filling shared space.
    # Quota is advisory (warn, not block) by default — configurable per use case.
    user_quota = 5368709120 # 5 GiB
  }
}

# =============================================================================
# Workflow: Multi-Protocol File System
# =============================================================================
# Provisions a single file system accessible by both Linux (NFS) and Windows (SMB):
#   - File system        (NFS + SMB dual-protocol, 100 GiB)
#   - NFS export policy  (Linux admin hosts, rw, no-root-squash, Kerberos)
#   - SMB share policy   (Windows domain users, change=allow, full_control=deny)
#
# Typical use case: shared home directories or project data accessible from
# both Linux compute nodes and Windows workstations on the same dataset.
#
# Key design decision: access_control_style = "nfs" means POSIX UIDs/GIDs
# are authoritative. Windows ACLs are translated by the array. If your org
# is Windows-primary, flip to "smb" and review safeguard_acls semantics.
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
  description = "Name of the dual-protocol file system."
}

variable "linux_admin_subnet" {
  type        = string
  default     = "10.30.0.0/24"
  description = "CIDR of trusted Linux admin hosts that need no-root-squash access (e.g. NFS provisioners, sysadmins)."
}

# ---------------------------------------------------------------------------
# NFS export policy — Linux access
# ---------------------------------------------------------------------------

resource "flashblade_nfs_export_policy" "linux" {
  name    = "${var.filesystem_name}-nfs"
  enabled = true
}

resource "flashblade_nfs_export_policy_rule" "linux_admin" {
  policy_name = flashblade_nfs_export_policy.linux.name
  client      = var.linux_admin_subnet
  permission  = "rw"

  # no-root-squash: trusted admin hosts retain root privilege on NFS.
  # Only use this for known, controlled subnets — never for broad app networks.
  access = "no-root-squash"

  # sys + krb5 allows clients to use either AUTH_SYS or Kerberos 5.
  # krb5 provides mutual authentication — ensures the client is who it claims to be,
  # which matters when the same dataset is exposed to AD-joined workstations.
  security = ["sys", "krb5"]
}

# ---------------------------------------------------------------------------
# SMB share policy — Windows access
# ---------------------------------------------------------------------------

resource "flashblade_smb_share_policy" "windows" {
  name    = "${var.filesystem_name}-smb"
  enabled = true
}

resource "flashblade_smb_share_policy_rule" "domain_users" {
  policy_name = flashblade_smb_share_policy.windows.name
  name        = "domain-users-rw"

  # "Domain Users" is the default group for all AD domain members.
  principal = "Domain Users"

  # read=allow: users can browse and read files.
  read = "allow"

  # change=allow: users can create, modify, and delete their own files.
  # This is the standard "user can edit" permission — equivalent to write in POSIX.
  change = "allow"

  # full_control=deny: prevents users from changing ACLs, taking ownership,
  # or modifying permissions on other users' files. Keeps the security model intact
  # even if a user account is compromised.
  full_control = "deny"
}

# ---------------------------------------------------------------------------
# File system — dual-protocol
# ---------------------------------------------------------------------------

resource "flashblade_file_system" "this" {
  name        = var.filesystem_name
  provisioned = 107374182400 # 100 GiB

  nfs {
    enabled = true

    # v3 enabled for legacy Linux clients (RHEL 6, older NAS tools).
    # v4.1 for modern clients — prefer v4.1 when all clients support it.
    v3_enabled   = true
    v4_1_enabled = true
  }

  smb {
    enabled = true

    # access_based_enumeration: Windows clients only see directories they have
    # permission to enter. Reduces confusion and information leakage in shared namespaces.
    access_based_enumeration_enabled = true

    # smb_encryption_enabled: forces SMB3 encryption for this share.
    # Required for SOC2/HIPAA compliance when traffic crosses untrusted switches.
    smb_encryption_enabled = true
  }

  multi_protocol {
    # nfs = POSIX UIDs/GIDs are authoritative; Windows SIDs are mapped.
    # Use this when Linux provisioned the data (e.g. Kubernetes PVCs) and
    # Windows needs secondary read access.
    access_control_style = "nfs"

    # safeguard_acls = true prevents the array from silently discarding ACLs
    # when a client writes with a protocol that doesn't support them.
    # Protects against Windows ACLs being lost by an NFS rewrite.
    safeguard_acls = true
  }

  nfs_export_policy = flashblade_nfs_export_policy.linux.name
  smb_share_policy  = flashblade_smb_share_policy.windows.name
}

# =============================================================================
# Workflow: Array Admin Baseline
# =============================================================================
# Day-1 array configuration for a newly racked FlashBlade:
#   - DNS   (internal domain + nameservers for hostname resolution)
#   - NTP   (time synchronisation — critical for replication and Kerberos)
#   - SMTP  (alert routing — ops team daily alerts, on-call pager for critical)
#
# Run this before deploying any storage resources. DNS and NTP failures
# cascade into authentication errors (Kerberos is time-sensitive) and
# broken replication schedules.
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
  description = "FlashBlade management endpoint URL."
}

variable "flashblade_api_token" {
  type        = string
  sensitive   = true
  description = "FlashBlade API token."
}

variable "domain" {
  type        = string
  default     = "corp.example.com"
  description = "Internal DNS domain. Appended to unqualified hostnames by the array (e.g. resolves 'backup-server' -> 'backup-server.corp.example.com')."
}

variable "dns_servers" {
  type        = list(string)
  default     = ["10.0.0.53", "10.0.1.53"]
  description = "Internal DNS resolver IPs. Two servers for redundancy — the array tries them in order. Use internal resolvers so the array hostname resolves correctly for NFS/SMB clients."
}

variable "ntp_servers" {
  type        = list(string)
  default     = ["0.pool.ntp.org", "1.pool.ntp.org", "2.pool.ntp.org"]
  description = "NTP server hostnames or IPs. Minimum 2 for redundancy; 3 preferred so the array can detect and ignore a drifting server via majority vote (RFC 5905 clock selection)."
}

variable "smtp_relay" {
  type        = string
  default     = "smtp.corp.example.com"
  description = "Internal SMTP relay hostname. The array sends alerts through this relay — it should be reachable from the management network."
}

variable "ops_team_email" {
  type        = string
  default     = "ops-team@corp.example.com"
  description = "Ops team distribution list. Receives warning-level alerts for day-to-day monitoring (capacity thresholds, drive health, replication lag)."
}

variable "oncall_email" {
  type        = string
  default     = "oncall@corp.example.com"
  description = "On-call pager address. Receives error-level alerts that require immediate response (hardware failure, space critical, array unreachable)."
}

# ---------------------------------------------------------------------------
# DNS configuration (singleton)
# ---------------------------------------------------------------------------

resource "flashblade_array_dns" "this" {
  domain      = var.domain
  nameservers = var.dns_servers

  # DNS Create uses GET-first, then PATCH — safe to apply on both new and
  # existing arrays without clobbering the existing config on first run.
}

# ---------------------------------------------------------------------------
# NTP configuration (singleton)
# ---------------------------------------------------------------------------

resource "flashblade_array_ntp" "this" {
  # PATCH sends only ntp_servers — other array settings are not touched.
  # Order matters: the array uses the first reachable server as primary.
  ntp_servers = var.ntp_servers
}

# ---------------------------------------------------------------------------
# SMTP + alert watchers (singleton)
# ---------------------------------------------------------------------------

resource "flashblade_array_smtp" "this" {
  relay_host    = var.smtp_relay
  sender_domain = var.domain

  # TLS is mandatory for PCI-DSS, SOC2, and HIPAA compliance.
  # If your relay doesn't support TLS, use "starttls" as a fallback,
  # but never "none" in production — alert emails may contain array hostnames
  # and capacity data that should not traverse the network in plaintext.
  encryption_mode = "tls"

  alert_watchers = [
    {
      email   = var.ops_team_email
      enabled = true

      # warning: day-to-day operational alerts (capacity approaching threshold,
      # single drive failure, replication behind schedule).
      # The ops team reviews these during business hours.
      minimum_notification_severity = "warning"
    },
    {
      email   = var.oncall_email
      enabled = true

      # error: requires immediate human response — hardware failure, space critical
      # (less than 5% free), array controller failover, data unavailability.
      # This address should route to a paging system (PagerDuty, OpsGenie, etc.).
      minimum_notification_severity = "error"
    },
  ]
}

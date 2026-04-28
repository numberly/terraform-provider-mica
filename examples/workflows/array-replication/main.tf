# =============================================================================
# Array Replication — Bidirectional FlashBlade-to-FlashBlade Connection
# =============================================================================
#
# Complete workflow for establishing encrypted replication between two
# FlashBlade arrays, including certificate trust, key exchange, and
# replication address configuration on both sides.
#
# Architecture:
#
#   [FlashBlade A (initiator)]              [FlashBlade B (passive)]
#       |                                        |
#   [Certificate Group]                     [Certificate Group]
#       |-- [Self-signed cert member]           |-- [Self-signed cert member]
#       |-- [Custom cert member]                |-- [Custom cert member]
#       |                                        |
#       |   connection_key ←────── [Connection Key generated on B]
#       |                                        |
#   [Array Connection] ──── encrypted ────> [Array Connection (auto-created)]
#       replication_addresses: [A-data-vip]      replication_addresses: [B-data-vip]
#
# Flow:
#   1. Create certificate groups on both arrays (trust bundles)
#   2. Add self-signed + custom certificates to each group
#   3. Generate a connection key on array B
#   4. Array A initiates the connection using B's key + management address
#   5. Array B adopts the auto-created passive connection and sets replication addresses
#
# Prerequisites:
#   - Two FlashBlade arrays with REST API v2.22+ (Purity//FB 4.6.7+)
#   - Network connectivity between management and data VIPs
#   - Self-signed certificates already present on each array
#
# Usage:
#   1. Copy terraform.tfvars.example to terraform.tfvars
#   2. Fill in endpoints, API tokens, and network addresses
#   3. terraform init && terraform apply
#
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source = "numberly/mica"
    }
  }
}

# --- Variables ---------------------------------------------------------------

variable "array_a_endpoint" {
  description = "Management endpoint of FlashBlade A (initiator)."
  type        = string
}

variable "array_a_api_token" {
  description = "API token for FlashBlade A."
  type        = string
  sensitive   = true
}

variable "array_b_endpoint" {
  description = "Management endpoint of FlashBlade B (passive)."
  type        = string
}

variable "array_b_api_token" {
  description = "API token for FlashBlade B."
  type        = string
  sensitive   = true
}

variable "array_a_name" {
  description = "Name of FlashBlade A as it appears in the other array's connection list."
  type        = string
}

variable "array_b_name" {
  description = "Name of FlashBlade B as it appears in the other array's connection list."
  type        = string
}

variable "array_a_management_address" {
  description = "Management IP or FQDN of FlashBlade A (used by B to connect)."
  type        = string
}

variable "array_b_management_address" {
  description = "Management IP or FQDN of FlashBlade B (used by A to connect)."
  type        = string
}

variable "array_a_replication_address" {
  description = "Data VIP of FlashBlade A for replication traffic."
  type        = string
}

variable "array_b_replication_address" {
  description = "Data VIP of FlashBlade B for replication traffic."
  type        = string
}

variable "array_a_cert_name" {
  description = "Name of the self-signed certificate on FlashBlade A."
  type        = string
}

variable "array_b_cert_name" {
  description = "Name of the self-signed certificate on FlashBlade B."
  type        = string
}

variable "custom_cert_name" {
  description = "Name of a custom certificate present on both arrays (e.g. imported via flashblade_certificate)."
  type        = string
  default     = ""
}

variable "encrypted" {
  description = "Whether replication traffic is encrypted."
  type        = bool
  default     = true
}

# --- Providers ---------------------------------------------------------------

provider "flashblade" {
  alias     = "array_a"
  endpoint  = var.array_a_endpoint
  api_token = var.array_a_api_token
}

provider "flashblade" {
  alias     = "array_b"
  endpoint  = var.array_b_endpoint
  api_token = var.array_b_api_token
}

# --- Step 1: Certificate Groups (trust bundles) ------------------------------

resource "flashblade_certificate_group" "array_a" {
  provider = flashblade.array_a
  name     = "_default_replication_certs"
}

resource "flashblade_certificate_group" "array_b" {
  provider = flashblade.array_b
  name     = "_default_replication_certs"
}

# --- Step 2: Add certificates to groups -------------------------------------

# Each array's self-signed certificate
resource "flashblade_certificate_group_member" "array_a_self" {
  provider         = flashblade.array_a
  group_name       = flashblade_certificate_group.array_a.name
  certificate_name = var.array_a_cert_name
}

resource "flashblade_certificate_group_member" "array_b_self" {
  provider         = flashblade.array_b
  group_name       = flashblade_certificate_group.array_b.name
  certificate_name = var.array_b_cert_name
}

# Custom certificate (optional — e.g. imported via flashblade_certificate)
resource "flashblade_certificate_group_member" "array_a_custom" {
  count            = var.custom_cert_name != "" ? 1 : 0
  provider         = flashblade.array_a
  group_name       = flashblade_certificate_group.array_a.name
  certificate_name = var.custom_cert_name
}

resource "flashblade_certificate_group_member" "array_b_custom" {
  count            = var.custom_cert_name != "" ? 1 : 0
  provider         = flashblade.array_b
  group_name       = flashblade_certificate_group.array_b.name
  certificate_name = var.custom_cert_name
}

# --- Step 3: Generate connection key on array B ------------------------------

resource "flashblade_array_connection_key" "array_b" {
  provider = flashblade.array_b
}

# --- Step 4: Array A initiates connection to B (active side) -----------------

resource "flashblade_array_connection" "a_to_b" {
  provider           = flashblade.array_a
  remote_name        = var.array_b_name
  management_address = var.array_b_management_address
  connection_key     = flashblade_array_connection_key.array_b.connection_key
  encrypted          = var.encrypted

  replication_addresses = [var.array_a_replication_address]
}

# --- Step 5: Array B adopts passive connection and sets replication address --

resource "flashblade_array_connection" "b_to_a" {
  provider              = flashblade.array_b
  remote_name           = var.array_a_name
  encrypted             = var.encrypted
  replication_addresses = [var.array_b_replication_address]

  # Wait for A to connect first — B's passive connection is auto-created by FlashBlade.
  depends_on = [flashblade_array_connection.a_to_b]
}

# --- Outputs -----------------------------------------------------------------

output "connection_a_to_b_status" {
  description = "Status of the connection from array A to array B."
  value       = flashblade_array_connection.a_to_b.status
}

output "connection_b_to_a_status" {
  description = "Status of the connection from array B to array A."
  value       = flashblade_array_connection.b_to_a.status
}

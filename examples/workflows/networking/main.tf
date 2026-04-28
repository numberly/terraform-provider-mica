# =============================================================================
# Networking Stack
# =============================================================================
#
# Complete workflow for configuring FlashBlade networking: LAG, subnet, and
# virtual IPs (VIPs) with server attachment.
#
# Architecture:
#
#   [LAG] (data source — hardware-managed)
#       |
#   [Subnet] (layer-3 network segment on the LAG)
#       |
#       +-- [VIP: data] --> attached to server (client data traffic)
#       +-- [VIP: sts] --> attached to server (S3 STS traffic)
#       +-- [VIP: egress-only] --> no server (outbound replication)
#       |
#   [Server] (data source — reads back attached VIPs)
#
# Usage:
#   1. Copy terraform.tfvars.example to terraform.tfvars
#   2. Fill in your FlashBlade endpoint, LAG name, and server name
#   3. terraform init && terraform apply
#
# =============================================================================

terraform {
  required_providers {
    flashblade = {
      source  = "numberly/mica"
      version = "~> 2.1"
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

variable "lag_name" {
  type        = string
  description = "Name of the existing link aggregation group to attach the subnet to."
  default     = "lag1"
}

variable "subnet_name" {
  type        = string
  description = "Name for the new subnet."
}

variable "subnet_prefix" {
  type        = string
  description = "CIDR prefix for the subnet (e.g. 10.0.0.0/24)."
}

variable "subnet_gateway" {
  type        = string
  description = "Gateway IP address for the subnet."
}

variable "server_name" {
  type        = string
  description = "Name of the existing FlashBlade server to attach VIPs to."
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
# Step 1: Read the LAG
# -----------------------------------------------------------------------------
# The link aggregation group is hardware-managed and cannot be created or
# deleted via Terraform. We reference it to attach the subnet to the correct
# physical bonded interface.

data "flashblade_link_aggregation_group" "this" {
  name = var.lag_name
}

# -----------------------------------------------------------------------------
# Step 2: Create subnet
# -----------------------------------------------------------------------------
# Layer-3 network segment on the LAG. All VIPs in this subnet share the same
# physical interface bond, VLAN tag, and MTU.

resource "flashblade_subnet" "this" {
  name     = var.subnet_name
  prefix   = var.subnet_prefix
  gateway  = var.subnet_gateway
  mtu      = 9000
  vlan     = 100
  lag_name = data.flashblade_link_aggregation_group.this.name
}

# -----------------------------------------------------------------------------
# Step 3: Create VIPs
# -----------------------------------------------------------------------------

# Client data traffic — must be attached to a server
resource "flashblade_network_interface" "data_vip" {
  address          = "10.0.0.10"
  subnet_name      = flashblade_subnet.this.name
  type             = "vip"
  services         = "data"
  attached_servers = [var.server_name]
}

# S3 STS traffic — must be attached to a server
resource "flashblade_network_interface" "sts_vip" {
  address          = "10.0.0.11"
  subnet_name      = flashblade_subnet.this.name
  type             = "vip"
  services         = "sts"
  attached_servers = [var.server_name]
}

# Outbound replication — no server attachment required
resource "flashblade_network_interface" "egress_vip" {
  address     = "10.0.0.12"
  subnet_name = flashblade_subnet.this.name
  type        = "vip"
  services    = "egress-only"
}

# -----------------------------------------------------------------------------
# Step 4: Read server with VIPs
# -----------------------------------------------------------------------------
# Reads back the server after VIPs are attached to expose the full network
# interface list as an output.

data "flashblade_server" "this" {
  name = var.server_name

  depends_on = [
    flashblade_network_interface.data_vip,
    flashblade_network_interface.sts_vip,
  ]
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------

output "lag_name" {
  description = "The link aggregation group name."
  value       = data.flashblade_link_aggregation_group.this.name
}

output "subnet_name" {
  description = "The subnet name."
  value       = flashblade_subnet.this.name
}

output "data_vip_address" {
  description = "Address of the data VIP."
  value       = flashblade_network_interface.data_vip.address
}

output "sts_vip_address" {
  description = "Address of the STS VIP."
  value       = flashblade_network_interface.sts_vip.address
}

output "egress_vip_address" {
  description = "Address of the egress-only VIP."
  value       = flashblade_network_interface.egress_vip.address
}

output "server_network_interfaces" {
  description = "Network interfaces attached to the server."
  value       = data.flashblade_server.this.network_interfaces
}

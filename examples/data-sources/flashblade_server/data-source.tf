data "flashblade_server" "example" {
  name = "existing-server"
}

# DNS configurations assigned to this server
output "server_dns" {
  value = data.flashblade_server.example.dns
}

# Directory services associated with this server (read-only)
output "server_directory_services" {
  value = data.flashblade_server.example.directory_services
}

# VIPs attached to this server (discovered automatically)
output "server_vips" {
  value = data.flashblade_server.example.network_interfaces
}

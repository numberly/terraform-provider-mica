data "flashblade_network_interface" "example" {
  name = "existing-vip"
}

output "network_interface_address" {
  value = data.flashblade_network_interface.example.address
}

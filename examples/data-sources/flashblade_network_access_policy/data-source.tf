data "flashblade_network_access_policy" "example" {
  name = "default"
}

output "nap_enabled" {
  value = data.flashblade_network_access_policy.example.enabled
}

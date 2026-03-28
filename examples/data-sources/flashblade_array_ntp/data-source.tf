data "flashblade_array_ntp" "example" {}

output "ntp_servers" {
  value = data.flashblade_array_ntp.example.ntp_servers
}

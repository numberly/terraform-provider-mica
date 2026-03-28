data "flashblade_array_dns" "example" {}

output "dns_nameservers" {
  value = data.flashblade_array_dns.example.nameservers
}

data "flashblade_array_dns" "example" {
  name = "dns-config"
}

output "dns_nameservers" {
  value = data.flashblade_array_dns.example.nameservers
}

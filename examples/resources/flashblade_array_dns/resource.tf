resource "flashblade_array_dns" "example" {
  name        = "dns-config"
  domain      = "example.com"
  nameservers = ["8.8.8.8", "8.8.4.4"]
}

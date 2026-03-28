# Singleton resource — manages the DNS configuration of the FlashBlade array.
resource "flashblade_array_dns" "example" {
  domain      = "example.com"
  nameservers = ["8.8.8.8", "8.8.4.4"]
}

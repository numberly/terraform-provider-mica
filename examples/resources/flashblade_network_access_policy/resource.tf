# Network access policies are system singletons on FlashBlade.
# This resource adopts an existing policy by name and manages its enabled state.
resource "flashblade_network_access_policy" "example" {
  name    = "default"
  enabled = true
}

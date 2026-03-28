resource "flashblade_network_access_policy_rule" "example" {
  policy_name = "default"
  client      = "10.0.0.0/8"
  effect      = "allow"

  # Optional: restrict to specific protocol interfaces
  interfaces = ["nfs", "smb"]
}

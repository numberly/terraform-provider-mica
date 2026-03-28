data "flashblade_snapshot_policy" "example" {
  name = "existing-snapshot-policy"
}

output "snapshot_policy_enabled" {
  value = data.flashblade_snapshot_policy.example.enabled
}

data "flashblade_nfs_export_policy" "example" {
  name = "existing-nfs-policy"
}

output "nfs_policy_enabled" {
  value = data.flashblade_nfs_export_policy.example.enabled
}

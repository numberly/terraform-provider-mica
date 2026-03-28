data "flashblade_smb_share_policy" "example" {
  name = "existing-smb-policy"
}

output "smb_policy_enabled" {
  value = data.flashblade_smb_share_policy.example.enabled
}

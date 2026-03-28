data "flashblade_smb_client_policy" "example" {
  name = "existing-smb-client-policy"
}

output "smb_client_policy_enabled" {
  value = data.flashblade_smb_client_policy.example.enabled
}

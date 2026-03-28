data "flashblade_s3_export_policy" "example" {
  name = "existing-s3-policy"
}

output "s3_export_policy_enabled" {
  value = data.flashblade_s3_export_policy.example.enabled
}

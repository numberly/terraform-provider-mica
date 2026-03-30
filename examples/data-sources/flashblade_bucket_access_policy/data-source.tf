data "flashblade_bucket_access_policy" "example" {
  bucket_name = "my-bucket"
}

output "bucket_access_policy_rule_count" {
  value = data.flashblade_bucket_access_policy.example.rule_count
}

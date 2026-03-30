data "flashblade_lifecycle_rule" "example" {
  bucket_name = "my-bucket"
  rule_id     = "rule-001"
}

output "lifecycle_rule_enabled" {
  value = data.flashblade_lifecycle_rule.example.enabled
}

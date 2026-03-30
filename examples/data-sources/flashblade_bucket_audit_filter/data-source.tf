data "flashblade_bucket_audit_filter" "example" {
  bucket_name = "my-bucket"
}

output "audit_filter_actions" {
  value = data.flashblade_bucket_audit_filter.example.actions
}

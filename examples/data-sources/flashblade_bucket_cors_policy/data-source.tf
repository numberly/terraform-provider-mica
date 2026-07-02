data "flashblade_bucket_cors_policy" "lookup" {
  bucket_name = "my-bucket"
}

output "cors_policy_id" {
  value = data.flashblade_bucket_cors_policy.lookup.id
}

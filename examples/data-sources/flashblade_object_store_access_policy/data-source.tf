data "flashblade_object_store_access_policy" "example" {
  name = "existing-s3-policy"
}

output "oap_arn" {
  value = data.flashblade_object_store_access_policy.example.arn
}

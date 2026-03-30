resource "flashblade_bucket_access_policy" "example" {
  # Creates the bucket access policy shell for the specified bucket.
  # Rules are managed separately via flashblade_bucket_access_policy_rule.
  bucket_name = "my-bucket"
}

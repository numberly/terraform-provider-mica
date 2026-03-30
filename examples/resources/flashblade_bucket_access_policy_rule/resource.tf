resource "flashblade_bucket_access_policy_rule" "example" {
  bucket_name = "my-bucket"
  name        = "allow-read"

  # Grant read-only access to the bucket and its objects
  actions = ["s3:GetObject", "s3:ListBucket"]
  effect  = "allow"

  resources = [
    "arn:aws:s3:::my-bucket",
    "arn:aws:s3:::my-bucket/*",
  ]
}

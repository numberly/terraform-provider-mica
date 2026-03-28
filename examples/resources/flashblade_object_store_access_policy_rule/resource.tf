resource "flashblade_object_store_access_policy_rule" "example" {
  policy_name = "terraform-s3-policy"
  name        = "allow-read"
  effect      = "allow"
  actions     = ["s3:GetObject", "s3:ListBucket"]
  resources   = ["arn:aws:s3:::my-bucket", "arn:aws:s3:::my-bucket/*"]
}

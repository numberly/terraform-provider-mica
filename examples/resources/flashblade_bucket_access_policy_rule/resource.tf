# Note: effect is read-only (always "allow", set by the API).
# The principals format depends on your FlashBlade firmware version.
resource "flashblade_bucket_access_policy_rule" "example" {
  bucket_name = "my-bucket"
  name        = "allow-read"

  actions    = ["s3:GetObject", "s3:ListBucket"]
  principals = ["my-account/admin"]
  resources  = ["*"]

  depends_on = [flashblade_bucket_access_policy.example]
}

# FlashBlade only supports a fully permissive (wildcard) CORS policy today, so this
# resource is a per-bucket toggle: its presence applies a wildcard rule (origins "*",
# headers "*", all HTTP methods) so browsers can use presigned URLs cross-origin.
# Destroying it removes the policy.
resource "flashblade_bucket_cors_policy" "example" {
  bucket_name = "my-bucket"
}

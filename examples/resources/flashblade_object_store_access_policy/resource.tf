# Object store access policies require the account-scoped name format:
# account-name/policy-name
resource "flashblade_object_store_access_policy" "example" {
  name        = "myaccount/s3-readonly"
  description = "Managed by Terraform — read-only S3 access"
}

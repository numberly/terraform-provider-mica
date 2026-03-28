resource "flashblade_object_store_access_policy" "example" {
  name        = "terraform-s3-policy"
  description = "Managed by Terraform — read-only S3 access"
}

resource "flashblade_bucket" "example" {
  name    = "terraform-example-bucket"
  account = "my-account"

  # Optional: set a quota limit in bytes (e.g. 100 GiB)
  quota_limit        = "107374182400"
  hard_limit_enabled = true
  versioning         = "enabled"

  # By default, destroy only soft-deletes the bucket.
  # Set to true to permanently eradicate on terraform destroy.
  destroy_eradicate_on_delete = false
}

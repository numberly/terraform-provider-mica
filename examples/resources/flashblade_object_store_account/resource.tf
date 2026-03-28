resource "flashblade_object_store_account" "example" {
  name = "my-account"

  # Optional: set a quota limit in bytes (e.g. 1 TiB)
  quota_limit        = "1099511627776"
  hard_limit_enabled = false
}

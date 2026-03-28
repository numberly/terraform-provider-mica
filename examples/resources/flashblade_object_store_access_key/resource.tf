resource "flashblade_object_store_access_key" "example" {
  object_store_account = "my-account"

  # Optional: keys are enabled by default
  enabled = true
}

output "access_key_id" {
  value = flashblade_object_store_access_key.example.access_key_id
}

output "secret_access_key" {
  value     = flashblade_object_store_access_key.example.secret_access_key
  sensitive = true
}

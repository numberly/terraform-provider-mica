data "flashblade_object_store_access_key" "example" {
  name = "my-account/admin/AKID1234567890"
}

output "access_key_enabled" {
  value = data.flashblade_object_store_access_key.example.enabled
}

# Auto-generated key (API generates name and secret)
resource "flashblade_object_store_access_key" "example" {
  object_store_account = "my-account"
}

# Cross-array replication: create a key with the same name and secret as the source
resource "flashblade_object_store_access_key" "replica" {
  object_store_account = "my-account"
  name                 = flashblade_object_store_access_key.example.name
  secret_access_key    = flashblade_object_store_access_key.example.secret_access_key
}

output "access_key_id" {
  value = flashblade_object_store_access_key.example.access_key_id
}

output "secret_access_key" {
  value     = flashblade_object_store_access_key.example.secret_access_key
  sensitive = true
}

data "flashblade_object_store_account" "example" {
  name = "existing-account"
}

output "object_count" {
  value = data.flashblade_object_store_account.example.object_count
}

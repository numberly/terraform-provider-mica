data "flashblade_object_store_account_export" "example" {
  name = "existing-account/export"
}

output "account_export_enabled" {
  value = data.flashblade_object_store_account_export.example.enabled
}

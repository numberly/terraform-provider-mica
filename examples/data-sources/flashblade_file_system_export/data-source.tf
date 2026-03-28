data "flashblade_file_system_export" "example" {
  name = "existing-fs/export"
}

output "file_system_export_enabled" {
  value = data.flashblade_file_system_export.example.enabled
}

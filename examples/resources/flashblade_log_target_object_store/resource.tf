resource "flashblade_log_target_object_store" "example" {
  name                = "my-log-target"
  bucket_name         = "audit-logs"
  log_name_prefix     = "audit"
  log_rotate_duration = 86400000
}

resource "flashblade_audit_object_store_policy" "example" {
  name    = "my-audit-policy"
  enabled = true

  log_targets = ["my-log-target"]
}

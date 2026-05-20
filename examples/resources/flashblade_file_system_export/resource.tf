resource "flashblade_file_system_export" "example" {
  file_system_name = "terraform-fs"
  server_name      = "terraform-server"
  policy_name      = "terraform-nfs-policy"

  # workload is auto-populated by the API when this export is associated with a workload.
  # It is read-only and cannot be set by the user.
}

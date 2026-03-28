resource "flashblade_file_system_export" "example" {
  file_system_name = "terraform-fs"
  server_name      = "terraform-server"
  policy_name      = "terraform-nfs-policy"
}

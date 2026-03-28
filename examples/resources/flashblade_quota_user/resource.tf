resource "flashblade_quota_user" "example" {
  file_system_name = "terraform-example"
  uid              = "1001"
  quota            = 5368709120 # 5 GiB in bytes
}

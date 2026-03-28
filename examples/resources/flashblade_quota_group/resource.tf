resource "flashblade_quota_group" "example" {
  file_system_name = "terraform-example"
  gid              = "1001"
  quota            = 10737418240 # 10 GiB in bytes
}

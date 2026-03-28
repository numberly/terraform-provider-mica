data "flashblade_quota_group" "example" {
  file_system_name = "my-filesystem"
  gid              = "1001"
}

output "group_quota_usage" {
  value = data.flashblade_quota_group.example.usage
}

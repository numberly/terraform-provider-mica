data "flashblade_quota_user" "example" {
  file_system_name = "my-filesystem"
  uid              = "1001"
}

output "user_quota_usage" {
  value = data.flashblade_quota_user.example.usage
}

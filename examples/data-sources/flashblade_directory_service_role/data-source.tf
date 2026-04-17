# Look up an existing directory service role by name.
data "flashblade_directory_service_role" "admins" {
  name = "array_admin"
}

output "admins_group" {
  value = data.flashblade_directory_service_role.admins.group
}

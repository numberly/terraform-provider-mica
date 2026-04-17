# Attach an additional management access policy to an existing directory service role.
# Use this resource to add policies without replacing the underlying role.
resource "flashblade_management_access_policy_directory_service_role_membership" "admins_readonly" {
  policy = "pure:policy/readonly"
  role   = "array_admin"
}

# Map an LDAP group to the FlashBlade array_admin management access policy.
resource "flashblade_directory_service_role" "admins" {
  group                      = "cn=fb-admins,ou=groups,dc=corp,dc=example,dc=com"
  group_base                 = "ou=groups,dc=corp,dc=example,dc=com"
  management_access_policies = ["pure:policy/array_admin"]
}

# Attach an additional management access policy to the same role. Each additional
# policy gets its own membership resource -- do NOT try to add them to the role's
# management_access_policies list (that attribute is immutable post-creation).
resource "flashblade_management_access_policy_directory_service_role_membership" "admins_storage" {
  policy = "pure:policy/storage_admin"
  role   = flashblade_directory_service_role.admins.name
}

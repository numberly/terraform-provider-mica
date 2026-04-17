# Import an existing directory service role by its server-generated name.
# The role name is derived from the first associated management access policy
# (for example: pure:policy/array_admin -> role name "array_admin").
terraform import flashblade_directory_service_role.admins array_admin

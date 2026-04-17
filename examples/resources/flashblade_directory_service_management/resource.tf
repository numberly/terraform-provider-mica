# LDAPS management directory service — authenticates FlashBlade admin logins
# against a corporate LDAP directory. ca_certificate_group references a
# certificate group already present on the array (see flashblade_certificate_group).
resource "flashblade_directory_service_management" "example" {
  enabled   = true
  uris      = ["ldaps://ldap.example.com:636"]
  base_dn   = "dc=example,dc=com"
  bind_user = "cn=flashblade,ou=service-accounts,dc=example,dc=com"

  # Sensitive, write-only. Not returned by the API; not surfaced in plan diffs.
  bind_password = var.ldap_bind_password

  ca_certificate_group = "corp-ca-group"

  # Management-specific LDAP attributes. Leave unset to let the array use its defaults
  # (sAMAccountName/User for Active Directory, uid/posixAccount|shadowAccount for OpenLDAP).
  user_login_attribute     = "sAMAccountName"
  user_object_class        = "User"
  ssh_public_key_attribute = "sshPublicKey"
}

variable "ldap_bind_password" {
  type      = string
  sensitive = true
}

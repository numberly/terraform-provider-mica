# Read the current management directory service configuration.
# Singleton — no arguments.
data "flashblade_directory_service_management" "current" {}

output "ldap_enabled" {
  value = data.flashblade_directory_service_management.current.enabled
}

output "ldap_uris" {
  value = data.flashblade_directory_service_management.current.uris
}

output "ca_cert_group_name" {
  value = data.flashblade_directory_service_management.current.ca_certificate_group.name
}

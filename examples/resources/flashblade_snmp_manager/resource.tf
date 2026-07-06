# Primary example: SNMPv3 trap destination.
resource "flashblade_snmp_manager" "prod_traps" {
  name         = "prod-snmp"
  host         = "snmp.example.com"
  notification = "trap"
  version      = "v3"

  v3 = {
    user               = "purity_user"
    auth_protocol      = "SHA"
    auth_passphrase    = "auth-secret-32max"
    privacy_protocol   = "AES"
    privacy_passphrase = "priv-secret-min8-max63"
  }
}

# Alternative: SNMPv2c (commented).
# resource "flashblade_snmp_manager" "v2c_example" {
#   name         = "legacy-snmp"
#   host         = "snmp-old.example.com"
#   notification = "inform"
#   version      = "v2c"
#
#   v2c = {
#     community = "public"
#   }
# }

# NOTE: switching `version` in place is permitted (no RequiresReplace). If you
# observe drift on the unused block after a switch, remove it via
# `terraform state rm` or taint+apply.

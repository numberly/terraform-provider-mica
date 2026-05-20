# Import by SNMP manager name. After import, sensitive fields
# (community, auth_passphrase, privacy_passphrase) are null in state.
# Set them in your HCL and `terraform apply` to materialise them.
terraform import flashblade_snmp_manager.prod_traps prod-snmp

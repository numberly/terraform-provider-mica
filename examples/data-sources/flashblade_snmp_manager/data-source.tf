data "flashblade_snmp_manager" "prod_traps" {
  name = "prod-snmp"
}

output "snmp_host" {
  value = data.flashblade_snmp_manager.prod_traps.host
}

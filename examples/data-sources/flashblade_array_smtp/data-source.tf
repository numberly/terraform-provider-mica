data "flashblade_array_smtp" "example" {}

output "smtp_relay_host" {
  value = data.flashblade_array_smtp.example.relay_host
}

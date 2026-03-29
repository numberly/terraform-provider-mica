data "flashblade_array_connection" "example" {
  remote_name = "remote-flashblade"
}

output "management_address" {
  value = data.flashblade_array_connection.example.management_address
}

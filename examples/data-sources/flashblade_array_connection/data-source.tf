data "flashblade_array_connection" "example" {
  remote_name = "remote-fb"
}

output "connection_status" {
  value = data.flashblade_array_connection.example.status
}

data "flashblade_server" "example" {
  name = "existing-server"
}

output "server_created" {
  value = data.flashblade_server.example.created
}

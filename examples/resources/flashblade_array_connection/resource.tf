resource "flashblade_array_connection" "example" {
  remote_name        = "remote-fb"
  management_address = "10.10.0.1"
  connection_key     = var.array_connection_key
  encrypted          = true

  replication_addresses = ["10.10.1.1", "10.10.1.2"]
}

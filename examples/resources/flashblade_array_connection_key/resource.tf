# Generate a connection key for use by the remote array.
# The key is ephemeral — each apply regenerates it.
resource "flashblade_array_connection_key" "key" {}

output "connection_key" {
  value     = flashblade_array_connection_key.key.connection_key
  sensitive = true
}

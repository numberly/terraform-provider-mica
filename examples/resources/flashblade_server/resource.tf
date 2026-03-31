resource "flashblade_server" "example" {
  name = "terraform-server"

  # Reference existing DNS configurations by name
  dns = ["management"]
}

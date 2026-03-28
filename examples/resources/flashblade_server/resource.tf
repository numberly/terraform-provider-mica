resource "flashblade_server" "example" {
  name = "terraform-server"

  dns {
    nameservers = ["8.8.8.8", "8.8.4.4"]
  }
}

resource "flashblade_network_interface" "example" {
  address          = "10.0.0.10"
  subnet_name      = "my-subnet"
  type             = "vip"
  services         = "data"
  attached_servers = ["srv-my-server"]
}

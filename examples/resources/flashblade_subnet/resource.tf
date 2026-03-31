resource "flashblade_subnet" "example" {
  name     = "my-subnet"
  prefix   = "10.0.0.0/24"
  gateway  = "10.0.0.1"
  mtu      = 9000
  vlan     = 100
  lag_name = "lag1"
}

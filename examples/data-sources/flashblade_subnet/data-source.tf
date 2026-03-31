data "flashblade_subnet" "example" {
  name = "existing-subnet"
}

output "subnet_enabled" {
  value = data.flashblade_subnet.example.enabled
}

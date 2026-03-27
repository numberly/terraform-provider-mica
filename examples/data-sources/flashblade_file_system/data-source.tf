data "flashblade_file_system" "example" {
  name = "existing-filesystem"
}

output "provisioned_size" {
  value = data.flashblade_file_system.example.provisioned
}

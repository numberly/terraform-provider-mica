data "flashblade_object_store_virtual_host" "example" {
  name = "existing-vhost"
}

output "virtual_host_hostname" {
  value = data.flashblade_object_store_virtual_host.example.hostname
}

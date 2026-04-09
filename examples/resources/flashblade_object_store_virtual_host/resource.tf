resource "flashblade_object_store_virtual_host" "example" {
  name     = "s3-vhost"
  hostname = "s3.example.com"
}

resource "flashblade_certificate_group_member" "example" {
  group_name       = flashblade_certificate_group.example.name
  certificate_name = flashblade_certificate.example.name
}

data "flashblade_certificate" "example" {
  name = "my-tls-cert"
}

output "certificate_status" {
  value = data.flashblade_certificate.example.status
}

output "certificate_valid_to" {
  value = data.flashblade_certificate.example.valid_to
}

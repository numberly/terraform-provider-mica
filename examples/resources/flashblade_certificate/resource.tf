resource "flashblade_certificate" "example" {
  name        = "my-tls-cert"
  certificate = file("${path.module}/cert.pem")
  private_key = file("${path.module}/key.pem")

  # Optional
  # intermediate_certificate = file("${path.module}/chain.pem")
  # passphrase               = var.cert_passphrase
}

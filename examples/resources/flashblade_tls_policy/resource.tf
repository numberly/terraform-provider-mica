resource "flashblade_tls_policy" "example" {
  name                  = "strict-tls"
  min_tls_version       = "TLSv1.2"
  enabled               = true
  appliance_certificate = "my-cert"
}

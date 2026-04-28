# Minimal target pointing at a remote S3 endpoint.
resource "flashblade_target" "primary" {
  name    = "s3-replication-target"
  address = "s3.us-east-1.amazonaws.com"
}

# Target with a CA certificate group for custom TLS validation.
# ca_certificate_group references a certificate group already present on the array.
resource "flashblade_target" "with_ca" {
  name                 = "private-s3-target"
  address              = "s3.internal.example.com"
  ca_certificate_group = "internal-ca-group"
}

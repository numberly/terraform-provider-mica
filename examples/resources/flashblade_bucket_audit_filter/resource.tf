resource "flashblade_bucket_audit_filter" "example" {
  bucket_name = "my-bucket"

  # Log read and write operations for compliance auditing
  actions = ["s3:GetObject", "s3:PutObject"]

  # Scope audit logging to specific prefixes
  s3_prefixes = ["logs/", "data/"]
}

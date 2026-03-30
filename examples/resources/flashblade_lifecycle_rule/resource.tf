resource "flashblade_lifecycle_rule" "example" {
  bucket_name = "my-bucket"
  enabled     = true

  # Apply only to objects under the /logs prefix
  prefix = "/logs"

  # Delete previous versions after 30 days
  keep_previous_version_for = 30

  # Abort incomplete multipart uploads after 7 days
  abort_incomplete_multipart_uploads_after = 7
}

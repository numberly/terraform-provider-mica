resource "flashblade_bucket_replica_link" "example" {
  local_bucket_name       = "my-local-bucket"
  remote_bucket_name      = "my-remote-bucket"
  remote_credentials_name = "my-remote-creds"

  # Replication starts immediately by default (paused = false)
  paused = false
}

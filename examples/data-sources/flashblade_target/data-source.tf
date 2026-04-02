# Read an existing replication target by name.
data "flashblade_target" "existing" {
  name = "s3-replication-target"
}

output "target_address" {
  value = data.flashblade_target.existing.address
}

output "target_status" {
  value = data.flashblade_target.existing.status
}

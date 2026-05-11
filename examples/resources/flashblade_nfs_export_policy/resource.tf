resource "flashblade_nfs_export_policy" "example" {
  name    = "terraform-nfs-policy"
  enabled = true

  # workload is a read-only computed field populated by the API when this policy
  # is associated with a workload. It cannot be set by the user.
  # workload = { id = "...", name = "..." }
}

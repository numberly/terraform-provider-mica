resource "flashblade_smb_share_policy" "example" {
  name    = "terraform-smb-policy"
  enabled = true

  # workload is a read-only, computed field populated by the API when the policy
  # is associated with a workload. It cannot be set by the user.
  # workload = { id = "...", name = "..." }  # managed by API
}

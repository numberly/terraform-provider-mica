resource "flashblade_smb_client_policy" "example" {
  name    = "terraform-smb-client-policy"
  enabled = true

  # workload is read-only (computed) — populated by the API when the policy
  # is associated with a workload. No configuration needed.
}

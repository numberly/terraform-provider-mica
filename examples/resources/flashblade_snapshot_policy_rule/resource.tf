resource "flashblade_snapshot_policy_rule" "example" {
  policy_name = "terraform-snapshot-policy"

  # Take a snapshot every 24 hours (86400000 ms)
  every = 86400000

  # Keep snapshots for 7 days (604800000 ms)
  keep_for = 604800000

  # Optional: custom suffix for snapshot names
  suffix = "daily"
}

# Assign a file system to a QoS policy.
# Valid member_type values: "file-systems", "realms"
# Note: "buckets" is NOT supported on FlashBlade API v2.22.
resource "flashblade_qos_policy_member" "example" {
  policy_name = "my-qos-policy"
  member_name = "my-filesystem"
  member_type = "file-systems"
}

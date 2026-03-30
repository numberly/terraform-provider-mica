resource "flashblade_qos_policy_member" "example" {
  policy_name = "my-qos-policy"
  member_name = "my-bucket"
  member_type = "buckets"
}

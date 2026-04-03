resource "flashblade_tls_policy_member" "example" {
  policy_name = "strict-tls"
  member_name = "data-vip1"
}

resource "flashblade_audit_object_store_policy_member" "example" {
  policy_name = "my-audit-policy"
  member_name = "my-bucket"
}

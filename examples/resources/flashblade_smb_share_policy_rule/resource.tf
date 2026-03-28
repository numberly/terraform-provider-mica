resource "flashblade_smb_share_policy_rule" "example" {
  policy_name = "terraform-smb-policy"
  principal   = "Everyone"

  read         = "allow"
  change       = "allow"
  full_control = "deny"
}

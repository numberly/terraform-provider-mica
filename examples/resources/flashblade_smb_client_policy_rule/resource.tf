resource "flashblade_smb_client_policy_rule" "example" {
  policy_name = "terraform-smb-client-policy"
  client      = "10.0.0.0/8"
}

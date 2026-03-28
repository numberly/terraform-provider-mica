resource "flashblade_s3_export_policy_rule" "example" {
  policy_name = "terraform-s3-policy"
  name        = "allowall"
  effect      = "allow"
  actions     = ["pure:S3Access"]
  resources   = ["*"]
}

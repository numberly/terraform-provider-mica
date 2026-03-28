resource "flashblade_object_store_account_export" "example" {
  account_name = "terraform-account"
  server_name  = "terraform-server"
  policy_name  = "terraform-s3-policy"
}

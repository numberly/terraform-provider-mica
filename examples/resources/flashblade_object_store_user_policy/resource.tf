resource "flashblade_object_store_user_policy" "example" {
  user_name   = "myaccount/myuser"
  policy_name = "my-access-policy"
}

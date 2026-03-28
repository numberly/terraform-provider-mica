resource "flashblade_nfs_export_policy_rule" "example" {
  policy_name = "terraform-nfs-policy"

  client     = "10.0.0.0/8"
  permission = "rw"
  access     = "root-squash"
  security   = ["sys"]
}

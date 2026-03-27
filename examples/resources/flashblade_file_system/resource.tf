resource "flashblade_file_system" "example" {
  name        = "terraform-example"
  provisioned = 1073741824 # 1 GiB

  nfs {
    enabled      = true
    v4_1_enabled = true
  }

  timeouts {
    create = "30m"
    delete = "60m"
  }
}

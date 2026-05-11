resource "flashblade_file_system" "example" {
  name        = "terraform-example"
  provisioned = 1073741824 # 1 GiB

  nfs {
    enabled      = true
    v4_1_enabled = true
  }

  # Uncomment to attach this file system to an existing workload (API 2.23+).
  # workload {
  #   name = "my-workload"
  # }

  timeouts {
    create = "30m"
    delete = "60m"
  }
}

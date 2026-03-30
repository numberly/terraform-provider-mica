resource "flashblade_qos_policy" "example" {
  name    = "my-qos-policy"
  enabled = true

  # 1 GiB/s bandwidth ceiling
  max_total_bytes_per_sec = 1073741824

  # 10,000 IOPS ceiling
  max_total_ops_per_sec = 10000
}

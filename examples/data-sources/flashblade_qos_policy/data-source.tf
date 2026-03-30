data "flashblade_qos_policy" "example" {
  name = "my-qos-policy"
}

output "qos_max_bandwidth" {
  value = data.flashblade_qos_policy.example.max_total_bytes_per_sec
}

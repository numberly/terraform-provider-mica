data "flashblade_workload" "example" {
  name = "my-workload"
}

output "workload_id" {
  value = data.flashblade_workload.example.id
}

output "workload_status" {
  value = data.flashblade_workload.example.status
}

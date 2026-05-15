data "flashblade_resiliency_group" "example" {
  name = "rg0"
}

output "resiliency_group_status" {
  value = data.flashblade_resiliency_group.example.status
}

output "resiliency_group_status_details" {
  value = data.flashblade_resiliency_group.example.status_details
}

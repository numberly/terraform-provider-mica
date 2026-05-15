data "flashblade_resiliency_group_member" "example" {
  resiliency_group_name = "rg0"
  member_name           = "fs-alpha"
}

output "resiliency_group_member_resource_type" {
  value = data.flashblade_resiliency_group_member.example.member_resource_type
}

output "resiliency_group_member_id" {
  value = data.flashblade_resiliency_group_member.example.member_id
}

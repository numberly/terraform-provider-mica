data "flashblade_link_aggregation_group" "example" {
  name = "lag1"
}

output "lag_speed" {
  value = data.flashblade_link_aggregation_group.example.lag_speed
}

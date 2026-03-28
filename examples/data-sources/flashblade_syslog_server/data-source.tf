data "flashblade_syslog_server" "example" {
  name = "existing-syslog"
}

output "syslog_server_uri" {
  value = data.flashblade_syslog_server.example.uri
}

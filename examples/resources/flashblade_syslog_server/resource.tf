resource "flashblade_syslog_server" "example" {
  name = "terraform-syslog"
  uri  = "udp://syslog.example.com:514"
}

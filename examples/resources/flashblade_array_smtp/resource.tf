# Singleton resource — manages SMTP relay config and alert watchers.
resource "flashblade_array_smtp" "example" {
  relay_host      = "smtp.example.com"
  sender_domain   = "example.com"
  encryption_mode = "tls"

  alert_watchers = [
    {
      email                         = "ops-team@example.com"
      enabled                       = true
      minimum_notification_severity = "warning"
    },
    {
      email                         = "oncall@example.com"
      enabled                       = true
      minimum_notification_severity = "error"
    },
  ]
}

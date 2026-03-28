# Singleton resource — manages the NTP server list of the FlashBlade array.
resource "flashblade_array_ntp" "example" {
  ntp_servers = ["0.pool.ntp.org", "1.pool.ntp.org"]
}

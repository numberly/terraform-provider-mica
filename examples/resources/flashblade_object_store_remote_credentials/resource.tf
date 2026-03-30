# Name must be formatted as <remote-name>/<credentials-name>
resource "flashblade_object_store_remote_credentials" "example" {
  name              = "remote-flashblade/my-remote-creds"
  access_key_id     = "PSABSSZRHPMEDKHMAAJPJBMMCMOEHOPS"
  secret_access_key = var.remote_secret_access_key
  remote_name       = "remote-flashblade"
}

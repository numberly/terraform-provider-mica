resource "flashblade_object_store_remote_credentials" "example" {
  name              = "my-remote-creds"
  access_key_id     = "PSABSSZRHPMEDKHMAAJPJBMMCMOEHOPS"
  secret_access_key = var.remote_secret_access_key
  remote_name       = "remote-flashblade"
}

provider "flashblade" {
  endpoint = "https://flashblade.example.com"
  auth = {
    api_token = var.flashblade_api_token
  }
}

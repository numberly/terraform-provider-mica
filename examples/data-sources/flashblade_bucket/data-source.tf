data "flashblade_bucket" "example" {
  name = "existing-bucket"
}

output "bucket_versioning" {
  value = data.flashblade_bucket.example.versioning
}

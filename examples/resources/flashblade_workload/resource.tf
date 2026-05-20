# Minimal example — only required fields.
resource "flashblade_workload" "example" {
  name        = "my-workload"
  preset_name = "my-preset"
}

# Full example — preset with parameters and eradication on delete.
resource "flashblade_workload" "full" {
  name        = "full-workload"
  preset_name = "my-preset"

  # Parameters are values passed to the preset at creation time.
  # Changing any parameter forces a new resource (RequiresReplace).
  parameters = [
    {
      name         = "capacity_gb"
      value_integer = 1024
    },
    {
      name         = "environment"
      value_string = "production"
    },
  ]

  # destroy_eradicate_on_delete: when true, permanently eradicates the workload
  # on `terraform destroy` (two-phase: soft-delete then eradicate).
  # When false (default), leaves the workload in the destroyed queue.
  destroy_eradicate_on_delete = true
}

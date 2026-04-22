import pulumi
import pulumi_flashblade as flashblade

# Configure the FlashBlade provider.
provider = flashblade.Provider("flashblade",
    endpoint="https://flashblade.example.com",
    auth={"api_token": "t.abc123"},
)

# Create a FlashBlade bucket with soft-delete (two-phase destroy).
# The default delete timeout is 30 minutes to allow for eradication polling.
# Set customTimeouts if your array requires more time.
bucket = flashblade.Bucket("example",
    name="pulumi-example-bucket",
    account="my-account",
    quota_limit="107374182400",
    hard_limit_enabled=True,
    versioning="enabled",
    destroy_eradicate_on_delete=False,
    opts=pulumi.ResourceOptions(
        provider=provider,
        custom_timeouts=pulumi.CustomTimeouts(
            create="20m",
            update="20m",
            delete="30m",
        ),
    ),
)

pulumi.export("bucket_name", bucket.name)

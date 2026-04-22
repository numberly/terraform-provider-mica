import pulumi
import pulumi_flashblade as flashblade

# Configure the FlashBlade provider.
# The endpoint and API token can also be set via environment variables:
#   FLASHBLADE_ENDPOINT, FLASHBLADE_AUTH_API_TOKEN
provider = flashblade.Provider("flashblade",
    endpoint="https://flashblade.example.com",
    auth={"api_token": "t.abc123"},
)

# Create a FlashBlade target (S3 replication endpoint).
target = flashblade.Target("primary",
    name="s3-replication-target",
    address="s3.us-east-1.amazonaws.com",
    opts=pulumi.ResourceOptions(provider=provider),
)

pulumi.export("target_name", target.name)

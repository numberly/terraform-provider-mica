import pulumi
import pulumi_flashblade as flashblade

# Configure the FlashBlade provider.
provider = flashblade.Provider("flashblade",
    endpoint="https://flashblade.example.com",
    auth={"api_token": "t.abc123"},
)

# Create remote credentials for cross-array replication.
# Name must be formatted as <remote-name>/<credentials-name>.
# secret_access_key is sensitive and write-only.
creds = flashblade.ObjectStoreRemoteCredentials("example",
    name="remote-flashblade/my-remote-creds",
    access_key_id="PSABSSZRHPMEDKHMAAJPJBMMCMOEHOPS",
    secret_access_key="sensitive-value-set-via-config-secret",
    remote_name="remote-flashblade",
    opts=pulumi.ResourceOptions(provider=provider),
)

pulumi.export("credentials_name", creds.name)

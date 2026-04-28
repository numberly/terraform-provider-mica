package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/numberly/terraform-provider-mica/pulumi/sdk/go/flashblade"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		provider, err := flashblade.NewProvider(ctx, "flashblade", &flashblade.ProviderArgs{
			Endpoint: pulumi.String("https://flashblade.example.com"),
			Auth: &flashblade.ProviderAuthArgs{ApiToken: pulumi.String("t.abc123")},
		})
		if err != nil {
			return err
		}

		// Create remote credentials for cross-array replication.
		// Name must be formatted as <remote-name>/<credentials-name>.
		// secret_access_key is sensitive and write-only.
		creds, err := flashblade.NewObjectStoreRemoteCredentials(ctx, "example", &flashblade.ObjectStoreRemoteCredentialsArgs{
			Name:            pulumi.String("remote-flashblade/my-remote-creds"),
			AccessKeyId:     pulumi.String("PSABSSZRHPMEDKHMAAJPJBMMCMOEHOPS"),
			SecretAccessKey: pulumi.String("sensitive-value-set-via-config-secret"),
			RemoteName:      pulumi.String("remote-flashblade"),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		ctx.Export("credentials_name", creds.Name)
		return nil
	})
}

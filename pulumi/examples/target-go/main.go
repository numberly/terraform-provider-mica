package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/numberly/terraform-provider-mica/pulumi/sdk/go/flashblade"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Configure the FlashBlade provider.
		// The endpoint and API token can also be set via environment variables:
		//   FLASHBLADE_ENDPOINT, FLASHBLADE_AUTH_API_TOKEN
		provider, err := flashblade.NewProvider(ctx, "flashblade", &flashblade.ProviderArgs{
			Endpoint: pulumi.String("https://flashblade.example.com"),
			Auth: &flashblade.ProviderAuthArgs{ApiToken: pulumi.String("t.abc123")},
		})
		if err != nil {
			return err
		}

		// Create a FlashBlade target (S3 replication endpoint).
		target, err := flashblade.NewTarget(ctx, "primary", &flashblade.TargetArgs{
			Name:    pulumi.String("s3-replication-target"),
			Address: pulumi.String("s3.us-east-1.amazonaws.com"),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		ctx.Export("target_name", target.Name)
		return nil
	})
}

package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go/flashblade"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		provider, err := flashblade.NewProvider(ctx, "flashblade", &flashblade.ProviderArgs{
			Endpoint: pulumi.String("https://flashblade.example.com"),
			Auth:     pulumi.StringMap{"api_token": pulumi.String("t.abc123")},
		})
		if err != nil {
			return err
		}

		// Create a FlashBlade bucket with soft-delete (two-phase destroy).
		// The default delete timeout is 30 minutes to allow for eradication polling.
		// Set customTimeouts if your array requires more time.
		bucket, err := flashblade.NewBucket(ctx, "example", &flashblade.BucketArgs{
			Name:                     pulumi.String("pulumi-example-bucket"),
			Account:                  pulumi.String("my-account"),
			QuotaLimit:               pulumi.String("107374182400"),
			HardLimitEnabled:         pulumi.Bool(true),
			Versioning:               pulumi.String("enabled"),
			DestroyEradicateOnDelete: pulumi.Bool(false),
		}, pulumi.Provider(provider), pulumi.Timeouts(&pulumi.CustomTimeouts{
			Create: "20m",
			Update: "20m",
			Delete: "30m",
		}))
		if err != nil {
			return err
		}

		ctx.Export("bucket_name", bucket.Name)
		return nil
	})
}

module github.com/numberly/opentofu-provider-flashblade/pulumi/provider

go 1.25

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-20260318212141-5525259d096b
	github.com/pulumi/pulumi-terraform-bridge/v3 v3.127.0
	github.com/pulumi/pulumi/pkg/v3 v3.231.0
	github.com/pulumi/pulumi/sdk/v3 v3.231.0
	github.com/numberly/opentofu-provider-flashblade v0.0.0
)

replace (
	github.com/hashicorp/terraform-plugin-sdk/v2 => github.com/pulumi/terraform-plugin-sdk/v2 v2.0.0-20260318212141-5525259d096b
	github.com/numberly/opentofu-provider-flashblade => ../../
)

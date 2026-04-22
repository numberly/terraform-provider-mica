module github.com/numberly/opentofu-provider-flashblade/pulumi/examples/remote-credentials-go

go 1.25

require (
	github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go v0.0.0
	github.com/pulumi/pulumi/sdk/v3 v3.231.0
)

replace github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go => ../../sdk/go

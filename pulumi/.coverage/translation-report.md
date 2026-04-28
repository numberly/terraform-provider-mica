# Pulumi HCL Translation Report

Generated: 2026-04-28T14:50:24Z

## Summary

This report captures the output of running tfgen with PULUMI_CONVERT=1.
The bridge attempts to convert Terraform HCL examples to Pulumi Python and Go.
Failures are non-blocking and indicate examples that need manual conversion.

## Conversion Output

```
warning: Unable to find the upstream provider's documentation:
The upstream repository is expected to be at "github.com/terraform-providers/terraform-provider-mica".

If the expected path is not correct, you should check the values of these fields (current values shown):
tfbridge.ProviderInfo{
	GitHubHost:              "github.com",
	GitHubOrg:               "terraform-providers",
	Name:                    "mica",
	TFProviderModuleVersion: "",
}
The original error is: error running 'go mod download -json' in "/home/gule/Workspace/team-infrastructure/terraform-provider-flashblade/pulumi/provider" dir for module: exit status 1

Output: {
	"Path": "github.com/terraform-providers/terraform-provider-mica",
	"Error": "module github.com/terraform-providers/terraform-provider-mica: not a known dependency"
}


Additional example conversion stats are available by setting COVERAGE_OUTPUT_DIR.

Provider:     mica
Success rate: NaN% (0/0)


General metrics:
	54 total resources containing 268 total inputs.
	41 total functions.

Argument metrics:
	0 argument descriptions were parsed from the upstream docs
	0 top-level input property descriptions came from an upstream attribute (as opposed to an argument). Nested arguments are not included in this count.
	1 of 268 resource inputs (0.37%) are missing descriptions in the schema

```

## Next Steps

- Review failures above and hand-write Pulumi examples in pulumi/examples/.
- Re-run 'make docs' after fixing HCL examples to verify conversion succeeds.

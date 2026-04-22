package provider

import (
	"context"
	"fmt"

	pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
	tfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"

	fb "github.com/numberly/opentofu-provider-flashblade/internal/provider"
	"github.com/numberly/opentofu-provider-flashblade/pulumi/provider/pkg/version"
)

// mainPkg is the package prefix used by tfgen for generated SDKs.
const mainPkg = "flashblade"

// Provider builds the full tfbridge.ProviderInfo with ShimProvider + all overrides.
// Consumed by both cmd/pulumi-tfgen-flashblade (build time) and
// cmd/pulumi-resource-flashblade (runtime).
func Provider() tfbridge.ProviderInfo {
	prov := tfbridge.ProviderInfo{
		Name:    "flashblade",
		Version: version.Version,
		// pftfbridge.ShimProvider wraps the terraform-plugin-framework provider so the bridge
		// can introspect its schema and route RPCs through it (BRIDGE-05, BRIDGE-01..03).
		P: pftfbridge.ShimProvider(fb.New(version.Version)()),

		DisplayName: "FlashBlade",
		Publisher:   "numberly",
		Description: "A Pulumi package for managing Pure Storage FlashBlade resources.",
		Keywords:    []string{"pulumi", "flashblade", "pure-storage", "category/infrastructure"},
		License:     "Apache-2.0",
		Homepage:    "https://github.com/numberly/opentofu-provider-flashblade",
		Repository:  "https://github.com/numberly/opentofu-provider-flashblade",
		// PluginDownloadURL — required for `pulumi plugin install --server
		// github://api.github.com/numberly ...` to resolve the runtime binary
		// from GitHub Releases (BRIDGE-05).
		PluginDownloadURL: "github://api.github.com/numberly",

		// Config — mirrors the TF provider schema 1:1 (D-01).
		// The TF provider uses nested blocks: auth.api_token, auth.oauth2.client_id, etc.
		// Fields already marked Sensitive in TF schema are auto-promoted to Pulumi Secrets
		// by the bridge. The entries below add belt-and-braces explicit Secret marks
		// on the top-level config attributes (SECRETS-01).
		Config: map[string]*tfbridge.SchemaInfo{
			"api_token": {
				Secret: tfbridge.True(),
			},
			"oauth2_client_secret": {
				Secret: tfbridge.True(),
			},
			"ca_certificate": {
				Secret: tfbridge.True(),
			},
		},

		MetadataInfo: tfbridge.NewProviderMetadata(nil),
	}

	// ---- Auto-tokenization across all 49 TF resources + data sources (D-02) ----
	// Module assignment driven by resource name prefix.
	prov.MustComputeTokens(tokens.KnownModules(
		"flashblade_", // TF resource prefix
		"index",       // default module
		[]string{
			"bucket",
			"filesystem",
			"policy",
			"objectstore",
			"array",
			"network",
		},
		tokens.MakeStandard(mainPkg),
	))

	// ---- Shared helper: omit the `timeouts` input block on every resource (MAPPING-02) ----
	// The `timeouts` block is a TF ergonomic; Pulumi expresses these via CustomTimeouts.
	//
	// ORDERING NOTE: `omitTimeoutsOnAll` runs AFTER `MustComputeTokens` populates
	// `prov.Resources`. Iterating an empty map before tokenization would be a no-op.
	// REQUIREMENTS.md MAPPING-02 was amended in commit ebc4eb8 to match this ordering,
	// and CONTEXT.md "Claude's Discretion" note about ordering is resolved by the same
	// rule: MustComputeTokens first, then omitTimeoutsOnAll, then per-resource overrides.
	// Do NOT reorder based on the original (now-amended) wording.
	omitTimeoutsOnAll(&prov)

	// ---- POC overrides (D-05) ----

	// flashblade_target — auto-tokenization baseline.
	// Note: CreateTimeout/UpdateTimeout/DeleteTimeout are not fields on ResourceInfo in
	// bridge v3.127.0. TF provider's timeouts block defaults (Create 20m, Update 20m,
	// Delete 20m) are inherited by the bridge from the shimmed TF schema. MAPPING-03
	// coverage is provided by the TF provider's existing timeouts block defaults.
	if _, ok := prov.Resources["flashblade_target"]; !ok {
		panic("flashblade_target resource not found after MustComputeTokens")
	}

	// flashblade_object_store_remote_credentials — write-once secret (SECRETS-02, PB3).
	// Belt-and-braces: mark secret_access_key as Secret in Fields so the bridge emits
	// it as secret in the Pulumi schema. AdditionalSecretOutputs is not available in
	// ResourceInfo in bridge v3.127.0; the TF Sensitive=true auto-promotion is the
	// runtime defense.
	if r, ok := prov.Resources["flashblade_object_store_remote_credentials"]; ok {
		if r.Fields == nil {
			r.Fields = map[string]*tfbridge.SchemaInfo{}
		}
		r.Fields["secret_access_key"] = &tfbridge.SchemaInfo{
			Secret: tfbridge.True(),
		}
	} else {
		panic("flashblade_object_store_remote_credentials resource not found after MustComputeTokens")
	}

	// flashblade_bucket — soft-delete + eradication polls (SOFTDELETE-01, PB1).
	// Note: DeleteTimeout field does not exist on ResourceInfo in bridge v3.127.0.
	// The TF provider's timeouts block default (Delete 30m) is inherited through the shim.
	// Pulumi users who need guaranteed extended timeouts should set customTimeouts on the
	// resource in their program.
	if _, ok := prov.Resources["flashblade_bucket"]; !ok {
		panic("flashblade_bucket resource not found after MustComputeTokens")
	}

	// flashblade_filesystem — same soft-delete pattern as bucket (SOFTDELETE-01).
	// Same note: DeleteTimeout not available; TF default 30m inherited via shim.
	if _, ok := prov.Resources["flashblade_filesystem"]; !ok {
		panic("flashblade_filesystem resource not found after MustComputeTokens")
	}

	// flashblade_object_store_access_policy_rule — composite ID with "/" separator,
	// string rule name (COMPOSITE-01, verified against
	// internal/provider/object_store_access_policy_rule_resource.go:361-387).
	if r, ok := prov.Resources["flashblade_object_store_access_policy_rule"]; ok {
		r.ComputeID = func(
			ctx context.Context,
			state resource.PropertyMap,
		) (resource.ID, error) {
			policyName, ok1 := state["policyName"]
			ruleName, ok2 := state["name"]
			if !ok1 || !ok2 {
				return "", fmt.Errorf(
					"object_store_access_policy_rule: missing policyName or name in state (got keys %v)",
					mapKeys(state),
				)
			}
			ps, psOk := policyName.V.(string)
			rs, rsOk := ruleName.V.(string)
			if !psOk || !rsOk {
				return "", fmt.Errorf(
					"object_store_access_policy_rule: policyName and name must be strings",
				)
			}
			return resource.ID(ps + "/" + rs), nil
		}
	} else {
		panic("flashblade_object_store_access_policy_rule resource not found after MustComputeTokens")
	}

	// ---- MAPPING-05: Autonaming deliberately omitted. FlashBlade names are operational. ----

	// Apply auto-aliases (D-02).
	prov.MustApplyAutoAliases()

	return prov
}

// omitTimeoutsOnAll hides the TF `timeouts` input block from the generated Pulumi schema
// across every resource (MAPPING-02 / PB7 / PB8). Must run AFTER MustComputeTokens (so
// prov.Resources is populated) and BEFORE per-resource overrides that touch Fields.
func omitTimeoutsOnAll(prov *tfbridge.ProviderInfo) {
	for _, r := range prov.Resources {
		if r == nil {
			continue
		}
		if r.Fields == nil {
			r.Fields = map[string]*tfbridge.SchemaInfo{}
		}
		r.Fields["timeouts"] = &tfbridge.SchemaInfo{Omit: true}
	}
}

func mapKeys(m resource.PropertyMap) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, string(k))
	}
	return out
}

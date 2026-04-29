package provider

import (
	"context"
	"fmt"

	pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
	tfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"

	fb "github.com/numberly/terraform-provider-mica/internal/provider"
	"github.com/numberly/terraform-provider-mica/pulumi/provider/pkg/version"
)

// mainPkg is the package prefix used by tfgen for generated SDKs.
const mainPkg = "mica"

// Provider builds the full tfbridge.ProviderInfo with ShimProvider + all overrides.
// Consumed by both cmd/pulumi-tfgen-mica (build time) and
// cmd/pulumi-resource-mica (runtime).
func Provider() tfbridge.ProviderInfo {
	prov := tfbridge.ProviderInfo{
		Name:    "mica",
		Version: version.Version,
		// pftfbridge.ShimProvider wraps the terraform-plugin-framework provider so the bridge
		// can introspect its schema and route RPCs through it (BRIDGE-05, BRIDGE-01..03).
		P: pftfbridge.ShimProvider(fb.New(version.Version)()),

		DisplayName: "Mica",
		Publisher:   "numberly",
		Description: "A Pulumi package for managing Pure Storage FlashBlade resources.",
		Keywords:    []string{"pulumi", "mica", "flashblade", "pure-storage", "category/infrastructure"},
		License:     "GPL-3.0-only",
		Homepage:    "https://github.com/numberly/terraform-provider-mica",
		Repository:  "https://github.com/numberly/terraform-provider-mica",
		// PluginDownloadURL — required for `pulumi plugin install --server
		// github://api.github.com/numberly ...` to resolve the runtime binary
		// from GitHub Releases (BRIDGE-05).
		PluginDownloadURL: "github://api.github.com/numberly",

		// Config — mirrors the TF provider schema 1:1 (D-01).
		// The TF provider uses a nested `auth` block containing api_token and oauth2 sub-block.
		// Top-level attributes: endpoint, ca_cert, ca_cert_file, insecure_skip_verify, max_retries.
		// Sensitive fields (auth.api_token, auth.oauth2.client_id, auth.oauth2.key_id) are
		// already marked Sensitive in the TF schema — the bridge auto-promotes them to Pulumi
		// Secrets. No explicit Config overrides are needed since all sensitive fields are
		// already handled by TF schema introspection (SECRETS-01).
		Config: map[string]*tfbridge.SchemaInfo{},

		MetadataInfo: tfbridge.NewProviderMetadata(nil),
	}

	// ---- Auto-tokenization across all TF resources + data sources (D-02) ----
	// SingleModule places every resource and data source in the `index` module.
	// This avoids the KnownModules limitation where resource names equal to a module
	// prefix (e.g. `flashblade_bucket` → name="" → error) and resources that don't
	// match any known prefix fall through to default.
	// Token form: mica:index:Bucket, mica:index:FileSystem, etc.
	prov.MustComputeTokens(tokens.SingleModule(
		"flashblade_", // TF resource prefix
		"index",       // single module
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

	// flashblade_array_connection_key — the `id` attribute is sensitive in the TF schema
	// (it holds the connection key value itself). The bridge requires explicit acknowledgment
	// that the ID will be exposed in Pulumi state (IDs cannot be encrypted in state).
	// Setting Secret: tfbridge.False() is the correct opt-in per bridge check_test.go
	// TestSensitiveIDWithOverride ("false" case passes, "true" case is a no-op that still fails).
	// NOTE: The error message in the bridge says "set Secret = tfbridge.True()" but that is
	// incorrect per the test suite — False() is required.
	if r, ok := prov.Resources["flashblade_array_connection_key"]; ok {
		if r.Fields == nil {
			r.Fields = map[string]*tfbridge.SchemaInfo{}
		}
		r.Fields["id"] = &tfbridge.SchemaInfo{
			Secret: tfbridge.False(),
		}
	} else {
		panic("flashblade_array_connection_key resource not found after MustComputeTokens")
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
		r.Fields["access_key_id"] = &tfbridge.SchemaInfo{
			Secret: tfbridge.True(),
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

	// flashblade_file_system — same soft-delete pattern as bucket (SOFTDELETE-01).
	// Same note: DeleteTimeout not available; TF default 30m inherited via shim.
	if _, ok := prov.Resources["flashblade_file_system"]; !ok {
		panic("flashblade_file_system resource not found after MustComputeTokens")
	}

	// ---- COMPOSITE-S3RULE: flashblade_s3_export_policy_rule ComputeID ----
	// The S3ExportPolicyRule API schema does NOT expose an `id` field (verified
	// against api_references/2.22.md and FlashBlade swagger). The TF "id" attribute
	// is Computed and assigned from rule.ID, but the API response leaves it empty,
	// so the Pulumi bridge fails Create with "empty resource.ID from create".
	// Composite ID matches the TF ImportState format: "policy_name/rule_index"
	// (verified against internal/provider/s3_export_policy_rule_resource.go ImportState).
	if r, ok := prov.Resources["flashblade_s3_export_policy_rule"]; ok {
		r.ComputeID = func(ctx context.Context, state resource.PropertyMap) (resource.ID, error) {
			policyName, ok1 := state["policyName"]
			index, ok2 := state["index"]
			if !ok1 || !ok2 {
				return "", fmt.Errorf(
					"s3_export_policy_rule: missing policyName or index in state (got keys %v)",
					mapKeys(state),
				)
			}
			ps, psOk := policyName.V.(string)
			if !psOk {
				return "", fmt.Errorf("s3_export_policy_rule: policyName must be a string")
			}
			idx, idxOk := index.V.(float64)
			if !idxOk {
				return "", fmt.Errorf("s3_export_policy_rule: index must be a number")
			}
			return resource.ID(fmt.Sprintf("%s/%d", ps, int64(idx))), nil
		}
	} else {
		panic("flashblade_s3_export_policy_rule resource not found after MustComputeTokens")
	}

	// flashblade_object_store_access_policy_rule — TF data.ID uses compositeID(policyName, ruleName)
	// = "policyName/ruleName". The TF "id" attribute is Computed and exposed in schema, so the
	// bridge picks it up via the shim without a ComputeID override. A redundant ComputeID would
	// create a second source of truth that could silently diverge if the TF helper changes (COMPOSITE-01).
	if _, ok := prov.Resources["flashblade_object_store_access_policy_rule"]; !ok {
		panic("flashblade_object_store_access_policy_rule resource not found after MustComputeTokens")
	}

	// ---- COMPOSITE-02: flashblade_bucket_access_policy_rule ComputeID ----
	// Composite ID: bucketName/ruleName (verified against
	// internal/provider/bucket_access_policy_rule_resource.go ImportState).
	if r, ok := prov.Resources["flashblade_bucket_access_policy_rule"]; ok {
		r.ComputeID = func(ctx context.Context, state resource.PropertyMap) (resource.ID, error) {
			bucketName, ok1 := state["bucketName"]
			ruleName, ok2 := state["name"]
			if !ok1 || !ok2 {
				return "", fmt.Errorf(
					"bucket_access_policy_rule: missing bucketName or name in state (got keys %v)",
					mapKeys(state),
				)
			}
			bs, bsOk := bucketName.V.(string)
			rs, rsOk := ruleName.V.(string)
			if !bsOk || !rsOk {
				return "", fmt.Errorf(
					"bucket_access_policy_rule: bucketName and name must be strings",
				)
			}
			return resource.ID(bs + "/" + rs), nil
		}
	} else {
		panic("flashblade_bucket_access_policy_rule resource not found after MustComputeTokens")
	}

	// flashblade_network_access_policy_rule — TF data.ID is rule.ID (UUID from the API).
	// No ComputeID override: the bridge uses the TF "id" attribute directly via the shim.
	// The prior ComputeID producing "policyName/ruleName" diverged from the UUID and would
	// have broken pulumi import / pulumi refresh round-trips (I1, COMPOSITE-03 corrected).
	if _, ok := prov.Resources["flashblade_network_access_policy_rule"]; !ok {
		panic("flashblade_network_access_policy_rule resource not found after MustComputeTokens")
	}

	// ---- COMPOSITE-04: flashblade_management_access_policy_directory_service_role_membership ComputeID ----
	// Composite ID: role/policy — role FIRST so SplitN("/", 2) correctly handles built-in
	// policy names containing slashes (e.g. "pure:policy/array_admin").
	// (verified against internal/provider/management_access_policy_directory_service_role_membership_resource.go:127)
	if r, ok := prov.Resources["flashblade_management_access_policy_directory_service_role_membership"]; ok {
		r.ComputeID = func(ctx context.Context, state resource.PropertyMap) (resource.ID, error) {
			roleName, ok1 := state["role"]
			policyName, ok2 := state["policy"]
			if !ok1 || !ok2 {
				return "", fmt.Errorf(
					"management_access_policy_dsr_membership: missing role or policy in state (got keys %v)",
					mapKeys(state),
				)
			}
			rs, rsOk := roleName.V.(string)
			ps, psOk := policyName.V.(string)
			if !rsOk || !psOk {
				return "", fmt.Errorf(
					"management_access_policy_dsr_membership: role and policy must be strings",
				)
			}
			return resource.ID(rs + "/" + ps), nil
		}
	} else {
		panic("flashblade_management_access_policy_directory_service_role_membership resource not found after MustComputeTokens")
	}

	// ---- SECRETS-02: Explicit Secret:tfbridge.True() for all remaining sensitive fields ----

	// flashblade_object_store_access_key — secret_access_key is write-only after creation (PB3).
	if r, ok := prov.Resources["flashblade_object_store_access_key"]; ok {
		if r.Fields == nil {
			r.Fields = map[string]*tfbridge.SchemaInfo{}
		}
		r.Fields["secret_access_key"] = &tfbridge.SchemaInfo{Secret: tfbridge.True()}
	} else {
		panic("flashblade_object_store_access_key resource not found after MustComputeTokens")
	}

	// flashblade_array_connection — connection_key is a sensitive credential (PB3).
	if r, ok := prov.Resources["flashblade_array_connection"]; ok {
		if r.Fields == nil {
			r.Fields = map[string]*tfbridge.SchemaInfo{}
		}
		r.Fields["connection_key"] = &tfbridge.SchemaInfo{Secret: tfbridge.True()}
	} else {
		panic("flashblade_array_connection resource not found after MustComputeTokens")
	}

	// flashblade_array_connection_key — connection_key is Sensitive:true in TF schema (PB3).
	// r.Fields already initialized above (id = False()). Add connection_key = True().
	if r, ok := prov.Resources["flashblade_array_connection_key"]; ok {
		r.Fields["connection_key"] = &tfbridge.SchemaInfo{Secret: tfbridge.True()}
	} else {
		panic("flashblade_array_connection_key resource not found after MustComputeTokens")
	}

	// flashblade_certificate — passphrase and private_key are sensitive credentials (PB3).
	if r, ok := prov.Resources["flashblade_certificate"]; ok {
		if r.Fields == nil {
			r.Fields = map[string]*tfbridge.SchemaInfo{}
		}
		r.Fields["passphrase"] = &tfbridge.SchemaInfo{Secret: tfbridge.True()}
		r.Fields["private_key"] = &tfbridge.SchemaInfo{Secret: tfbridge.True()}
	} else {
		panic("flashblade_certificate resource not found after MustComputeTokens")
	}

	// flashblade_directory_service_management — bind_password is a sensitive credential (PB3).
	if r, ok := prov.Resources["flashblade_directory_service_management"]; ok {
		if r.Fields == nil {
			r.Fields = map[string]*tfbridge.SchemaInfo{}
		}
		r.Fields["bind_password"] = &tfbridge.SchemaInfo{Secret: tfbridge.True()}
	} else {
		panic("flashblade_directory_service_management resource not found after MustComputeTokens")
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

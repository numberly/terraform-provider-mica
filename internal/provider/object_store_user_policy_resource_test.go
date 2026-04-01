package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// setupObjectStoreUserPolicyTest creates a mock server with object store user handlers
// (including policy sub-resource) and returns ProtoV6ProviderFactories.
func setupObjectStoreUserPolicyTest(t *testing.T) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	ms := testmock.NewMockServer()
	t.Cleanup(ms.Close)

	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterObjectStoreUserHandlers(ms.Mux, accountStore)

	t.Setenv("FLASHBLADE_HOST", ms.URL())
	t.Setenv("FLASHBLADE_API_TOKEN", "mock-test-token")

	return map[string]func() (tfprotov6.ProviderServer, error){
		"flashblade": providerserver.NewProtocol6WithError(New("test")()),
	}
}

// TestMocked_ObjectStoreUserPolicy_Lifecycle verifies Create, plan-convergence, and ImportState
// for the flashblade_object_store_user_policy member resource.
func TestMocked_ObjectStoreUserPolicy_Lifecycle(t *testing.T) {
	factories := setupObjectStoreUserPolicyTest(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Step 1: Create user + policy association
			{
				Config: `
					resource "flashblade_object_store_user" "test" {
						name = "testaccount/policyuser"
					}
					resource "flashblade_object_store_user_policy" "test" {
						user_name   = "testaccount/policyuser"
						policy_name = "test-access-policy"
						depends_on  = [flashblade_object_store_user.test]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("flashblade_object_store_user_policy.test", "user_name", "testaccount/policyuser"),
					resource.TestCheckResourceAttr("flashblade_object_store_user_policy.test", "policy_name", "test-access-policy"),
				),
			},
			// Step 2: Plan convergence — same config must produce zero diff.
			{
				Config: `
					resource "flashblade_object_store_user" "test" {
						name = "testaccount/policyuser"
					}
					resource "flashblade_object_store_user_policy" "test" {
						user_name   = "testaccount/policyuser"
						policy_name = "test-access-policy"
						depends_on  = [flashblade_object_store_user.test]
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: ImportState — import ID format: account/username/policyname.
			// The policy resource has no 'id' attribute — use user_name as the verify identifier.
			{
				ResourceName:                      "flashblade_object_store_user_policy.test",
				ImportState:                       true,
				ImportStateVerify:                 true,
				ImportStateVerifyIgnore:           []string{"timeouts"},
				ImportStateId:                     "testaccount/policyuser/test-access-policy",
				ImportStateVerifyIdentifierAttribute: "user_name",
			},
		},
	})
}

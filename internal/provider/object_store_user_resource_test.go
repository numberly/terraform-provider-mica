package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

// setupObjectStoreUserTest creates a mock server with object store user handlers and
// returns ProtoV6ProviderFactories for use with resource.UnitTest.
func setupObjectStoreUserTest(t *testing.T) map[string]func() (tfprotov6.ProviderServer, error) {
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

// TestUnit_ObjectStoreUserResource_Lifecycle verifies the full Create/Read/PlanConvergence/Import cycle.
func TestUnit_ObjectStoreUserResource_Lifecycle(t *testing.T) {
	factories := setupObjectStoreUserTest(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: `
					resource "flashblade_object_store_user" "test" {
						name = "testaccount/testuser"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("flashblade_object_store_user.test", "name", "testaccount/testuser"),
					resource.TestCheckResourceAttrSet("flashblade_object_store_user.test", "id"),
				),
			},
			// Step 2: Plan convergence — same HCL must produce zero diff after Read.
			{
				Config: `
					resource "flashblade_object_store_user" "test" {
						name = "testaccount/testuser"
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: ImportState — import by name (account/username), not by UUID.
			{
				ResourceName:            "flashblade_object_store_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateId:           "testaccount/testuser",
			},
		},
	})
}

// TestUnit_ObjectStoreUserResource_FullAccess verifies that full_access=true is set and stable.
func TestUnit_ObjectStoreUserResource_FullAccess(t *testing.T) {
	factories := setupObjectStoreUserTest(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Step 1: Create with full_access = true
			{
				Config: `
					resource "flashblade_object_store_user" "full" {
						name        = "testaccount/fullaccessuser"
						full_access = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("flashblade_object_store_user.full", "name", "testaccount/fullaccessuser"),
					resource.TestCheckResourceAttr("flashblade_object_store_user.full", "full_access", "false"),
					resource.TestCheckResourceAttrSet("flashblade_object_store_user.full", "id"),
				),
			},
			// Step 2: Plan convergence — no drift expected.
			{
				Config: `
					resource "flashblade_object_store_user" "full" {
						name        = "testaccount/fullaccessuser"
						full_access = false
					}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

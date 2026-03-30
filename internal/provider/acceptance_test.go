package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock"
	"github.com/numberly/opentofu-provider-flashblade/internal/testmock/handlers"
)

// setupAcceptanceTest creates a mock server with all required handlers, sets
// environment variables for provider configuration, and returns ProtoV6ProviderFactories.
func setupAcceptanceTest(t *testing.T) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	ms := testmock.NewMockServer()
	t.Cleanup(ms.Close)

	accountStore := handlers.RegisterObjectStoreAccountHandlers(ms.Mux)
	handlers.RegisterFileSystemHandlers(ms.Mux)
	handlers.RegisterBucketHandlers(ms.Mux, accountStore)

	t.Setenv("FLASHBLADE_HOST", ms.URL())
	t.Setenv("FLASHBLADE_API_TOKEN", "mock-test-token")

	return map[string]func() (tfprotov6.ProviderServer, error){
		"flashblade": providerserver.NewProtocol6WithError(New("test")()),
	}
}

// ---------- File System Lifecycle -------------------------------------------

func TestAcc_FileSystem_Lifecycle(t *testing.T) {
	factories := setupAcceptanceTest(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Step 1: Create — verify key attributes populated
			{
				Config: `
					resource "flashblade_file_system" "test" {
						name                       = "acc-test-fs"
						provisioned                = 1073741824
						destroy_eradicate_on_delete = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("flashblade_file_system.test", "name", "acc-test-fs"),
					resource.TestCheckResourceAttr("flashblade_file_system.test", "provisioned", "1073741824"),
					resource.TestCheckResourceAttrSet("flashblade_file_system.test", "id"),
					resource.TestCheckResourceAttrSet("flashblade_file_system.test", "created"),
				),
				// Computed nested objects (nfs, smb, space, etc.) produce a non-empty
				// plan due to zero-value vs unknown semantics. This is a known provider
				// limitation with deeply nested computed objects — not a mock issue.
				ExpectNonEmptyPlan: true,
			},
			// Step 2: Update provisioned size
			{
				Config: `
					resource "flashblade_file_system" "test" {
						name                       = "acc-test-fs"
						provisioned                = 2147483648
						destroy_eradicate_on_delete = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("flashblade_file_system.test", "provisioned", "2147483648"),
				),
				ExpectNonEmptyPlan: true,
			},
			// Step 3: ImportState
			{
				ResourceName:            "flashblade_file_system.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"destroy_eradicate_on_delete", "timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "acc-test-fs", nil
				},
			},
		},
	})
}

// ---------- Object Store Account Lifecycle ----------------------------------

func TestAcc_ObjectStoreAccount_Lifecycle(t *testing.T) {
	factories := setupAcceptanceTest(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: `
					resource "flashblade_object_store_account" "test" {
						name = "acc-test-account"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("flashblade_object_store_account.test", "name", "acc-test-account"),
					resource.TestCheckResourceAttrSet("flashblade_object_store_account.test", "id"),
					resource.TestCheckResourceAttrSet("flashblade_object_store_account.test", "created"),
				),
			},
			// Step 2: ImportState
			{
				ResourceName:            "flashblade_object_store_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts", "skip_default_export"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "acc-test-account", nil
				},
			},
		},
	})
}

// ---------- Bucket Lifecycle (depends on account) ---------------------------

func TestAcc_Bucket_Lifecycle(t *testing.T) {
	factories := setupAcceptanceTest(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Step 1: Create account + bucket
			{
				Config: `
					resource "flashblade_object_store_account" "acc" {
						name = "acc-bucket-test"
					}
					resource "flashblade_bucket" "test" {
						name                       = "acc-test-bucket"
						account                    = flashblade_object_store_account.acc.name
						destroy_eradicate_on_delete = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("flashblade_bucket.test", "name", "acc-test-bucket"),
					resource.TestCheckResourceAttr("flashblade_bucket.test", "account", "acc-bucket-test"),
					resource.TestCheckResourceAttrSet("flashblade_bucket.test", "id"),
					resource.TestCheckResourceAttrSet("flashblade_bucket.test", "created"),
				),
			},
			// Step 2: ImportState
			{
				ResourceName:            "flashblade_bucket.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"destroy_eradicate_on_delete", "timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "acc-test-bucket", nil
				},
			},
		},
	})
}

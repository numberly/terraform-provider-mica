package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestUnit_ObjectStoreUserDataSource_Basic verifies that the data source populates all
// attributes (name, id, full_access) after the resource is created.
func TestUnit_ObjectStoreUserDataSource_Basic(t *testing.T) {
	factories := setupObjectStoreUserTest(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "flashblade_object_store_user" "src" {
						name = "testaccount/dsuser"
					}

					data "flashblade_object_store_user" "test" {
						name       = flashblade_object_store_user.src.name
						depends_on = [flashblade_object_store_user.src]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.flashblade_object_store_user.test", "name", "testaccount/dsuser"),
					resource.TestCheckResourceAttrSet("data.flashblade_object_store_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.flashblade_object_store_user.test", "full_access"),
				),
			},
		},
	})
}

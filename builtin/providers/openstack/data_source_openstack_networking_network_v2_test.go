package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccOpenStackNetworkingNetworkV2DataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccOpenStackNetworkingNetworkV2DataSource_network,
			},
			resource.TestStep{
				Config: testAccOpenStackNetworkingNetworkV2DataSource_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingNetworkV2DataSourceID("data.openstack_networking_network_v2.net"),
					resource.TestCheckResourceAttr(
						"data.openstack_networking_network_v2.net", "name", "tf_test_network"),
					resource.TestCheckResourceAttr(
						"data.openstack_networking_network_v2.net", "admin_state_up", "true"),
				),
			},
		},
	})
}

func testAccCheckNetworkingNetworkV2DataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find image data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Image data source ID not set")
		}

		return nil
	}
}

// Standard CirrOS image
const testAccOpenStackNetworkingNetworkV2DataSource_network = `
resource "openstack_networking_network_v2" "net" {
        name = "tf_test_network"
        admin_state_up = "true"
}
`

var testAccOpenStackNetworkingNetworkV2DataSource_basic = fmt.Sprintf(`
%s

data "openstack_networking_network_v2" "net" {
	name = "${openstack_networking_network_v2.net.name}"
}
`, testAccOpenStackNetworkingNetworkV2DataSource_network)

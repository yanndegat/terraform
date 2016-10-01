package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"os"
)

func TestAccImageV2_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageV2Destroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccImageV2_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageV2Exists(t, "openstack_image_v2.foo"),
				),
			},
		},
	})
}

func testAccCheckImageV2Destroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	imageClient, err := config.imageV2Client(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("(testAccCheckImageV2Destroy) Error creating OpenStack Image: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_image_v2" {
			continue
		}

		_, err := images.Get(imageClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Image still exists")
		}
	}

	return nil
}

func testAccCheckImageV2Exists(t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		imageClient, err := config.imageV2Client(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("(testAccCheckImageV2Destroy) Error creating OpenStack Image: %s", err)
		}

		found, err := images.Get(imageClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Image not found")
		}

		return nil
	}
}

var testAccImageV2_basic = fmt.Sprintf(`
  resource "openstack_image_v2" "foo" {
      region = "%s"
      name   = "%s"
      image_source_url = "https://releases.rancher.com/os/latest/rancheros-openstack.img"
      container_format = "bare"
      disk_format = "qcow2"
  }`, os.Getenv("OS_REGION_NAME"), os.Getenv("OS_IMAGE_NAME"))

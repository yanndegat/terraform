package openstack

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
)

func dataSourceNetworkingNetworkV2() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetworkingNetworkV2Read,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OS_REGION_NAME", ""),
			},
			"admin_state_up": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"shared": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkingNetworkV2Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	networkingClient, err := config.networkingV2Client(GetRegion(d))
	name := d.Get("name").(string)
	netId, err := networks.IDFromName(networkingClient, name)
	if err != nil {
		return fmt.Errorf("Error finding network with name %#v: %s", name, err)
	}
	d.SetId(netId)

	//
	n, err := networks.Get(networkingClient, d.Id()).Extract()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Retrieved Network %s: %+v", d.Id(), n)

	d.Set("admin_state_up", strconv.FormatBool(n.AdminStateUp))
	d.Set("shared", strconv.FormatBool(n.Shared))
	d.Set("tenant_id", n.TenantID)
	d.Set("region", GetRegion(d))

	return nil
}

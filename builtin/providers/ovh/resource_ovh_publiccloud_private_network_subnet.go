package ovh

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/ovh/go-ovh/ovh"
)

func resourcePublicCloudPrivateNetworkSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourcePublicCloudPrivateNetworkSubnetCreate,
		Read:   resourcePublicCloudPrivateNetworkSubnetRead,
		Delete: resourcePublicCloudPrivateNetworkSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_PROJECT_ID", ""),
			},
			"network_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dhcp": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"start": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"end": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"no_gateway": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"gateway_ip": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"ip_pools": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"region": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"dhcp": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"end": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"start": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourcePublicCloudPrivateNetworkSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("project_id").(string)
	networkId := d.Get("network_id").(string)
	params := &PublicCloudPrivateNetworksCreateOpts{
		ProjectId: projectId,
		NetworkId: networkId,
		Dhcp:      d.Get("dhcp").(bool),
		NoGateway: d.Get("no_gateway").(bool),
		Start:     d.Get("start").(string),
		End:       d.Get("end").(string),
		Network:   d.Get("network").(string),
		Region:    d.Get("region").(string),
	}

	r := &PublicCloudPrivateNetworksResponse{}

	log.Printf("[DEBUG] Will create public cloud private network subnet: %s", params)

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet", projectId, networkId)

	err := config.OVHClient.Post(endpoint, params, r)
	if err != nil {
		return fmt.Errorf("calling %s with params %s:\n\t %q", endpoint, params, err)
	}

	log.Printf("[DEBUG] Created Private Network Subnet %s", r)

	//set id
	d.SetId(r.Id)

	return nil
}

func resourcePublicCloudPrivateNetworkSubnetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("project_id").(string)
	networkId := d.Get("network_id").(string)

	r := []*PublicCloudPrivateNetworksResponse{}

	log.Printf("[DEBUG] Will read public cloud private network subnet for project: %s, network: %s, id: %s", projectId, networkId, d.Id())

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet", projectId, networkId)

	err := config.OVHClient.Get(endpoint, &r)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %q", endpoint, err)
	}

	err = readPublicCloudPrivateNetworkSubnet(d, r)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Read Public Cloud Private Network %v", r)
	return nil
}

func resourcePublicCloudPrivateNetworkSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("project_id").(string)
	networkId := d.Get("network_id").(string)
	id := d.Id()

	log.Printf("[DEBUG] Will delete public cloud private network subnet for project: %s, network: %s, id: %s", projectId, networkId, id)

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet/%s", projectId, id, id)

	err := config.OVHClient.Delete(endpoint, nil)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %q", endpoint, err)
	}

	d.SetId("")

	log.Printf("[DEBUG] Deleted Public Cloud %s Private Network %s Subnet %s", projectId, networkId, id)
	return nil
}

func publicCloudPrivateNetworkSubnetExists(projectId, networkId, id string, c *ovh.Client) error {
	r := []*PublicCloudPrivateNetworksResponse{}

	log.Printf("[DEBUG] Will read public cloud private network subnet for project: %s, network: %s, id: %s", projectId, networkId, id)

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet", projectId, networkId)

	err := c.Get(endpoint, &r)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %q", endpoint, err)
	}

	s := findPublicCloudPrivateNetworkSubnet(r, id)
	if s == nil {
		return fmt.Errorf("Subnet %s doesn't exists for project %s and network %s", id, projectId, networkId)
	}

	return nil
}

func findPublicCloudPrivateNetworkSubnet(rs []*PublicCloudPrivateNetworksResponse, id string) *PublicCloudPrivateNetworksResponse {
	for i := range rs {
		if rs[i].Id == id {
			return rs[i]
		}
	}

	return nil
}

func readPublicCloudPrivateNetworkSubnet(d *schema.ResourceData, rs []*PublicCloudPrivateNetworksResponse) error {
	r := findPublicCloudPrivateNetworkSubnet(rs, d.Id())
	if r == nil {
		return fmt.Errorf("%s subnet not found", d.Id())
	}

	d.Set("gateway_ip", r.GatewayIp)
	d.Set("cidr", r.Cidr)

	ippools := make([]map[string]interface{}, 0)
	for i := range r.IPPools {
		ippool := make(map[string]interface{})
		ippool["network"] = r.IPPools[i].Network
		ippool["region"] = r.IPPools[i].Region
		ippool["dhcp"] = r.IPPools[i].Dhcp
		ippool["start"] = r.IPPools[i].Start
		ippool["end"] = r.IPPools[i].End
		ippools = append(ippools, ippool)
	}

	d.Set("network", ippools[0]["network"])
	d.Set("region", ippools[0]["region"])
	d.Set("dhcp", ippools[0]["dhcp"])
	d.Set("start", ippools[0]["start"])
	d.Set("end", ippools[0]["end"])
	d.Set("ip_pools", ippools)

	if r.GatewayIp == "" {
		d.Set("no_gateway", true)
	} else {
		d.Set("no_gateway", false)
	}

	d.SetId(r.Id)
	return nil
}

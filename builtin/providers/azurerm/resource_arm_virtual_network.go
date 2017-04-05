package azurerm

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceArmVirtualNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmVirtualNetworkCreate,
		Read:   resourceArmVirtualNetworkRead,
		Update: resourceArmVirtualNetworkCreate,
		Delete: resourceArmVirtualNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"address_space": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"dns_servers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"subnet": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"address_prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
						"security_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"route_table": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceAzureSubnetHash,
			},

			"location": locationSchema(),

			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmVirtualNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	vnetClient := client.vnetClient

	log.Printf("[INFO] preparing arguments for Azure ARM virtual network creation.")

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})

	vnet := network.VirtualNetwork{
		Name:                           &name,
		Location:                       &location,
		VirtualNetworkPropertiesFormat: getVirtualNetworkProperties(d, meta),
		Tags: expandTags(tags),
	}

	_, err := vnetClient.CreateOrUpdate(resGroup, name, vnet, make(chan struct{}))
	if err != nil {
		return err
	}

	read, err := vnetClient.Get(resGroup, name, "")
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Virtual Network %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmVirtualNetworkRead(d, meta)
}

func resourceArmVirtualNetworkRead(d *schema.ResourceData, meta interface{}) error {
	vnetClient := meta.(*ArmClient).vnetClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["virtualNetworks"]

	resp, err := vnetClient.Get(resGroup, name, "")
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure virtual network %s: %s", name, err)
	}

	vnet := *resp.VirtualNetworkPropertiesFormat

	// update appropriate values
	d.Set("resource_group_name", resGroup)
	d.Set("name", resp.Name)
	d.Set("location", resp.Location)
	d.Set("address_space", vnet.AddressSpace.AddressPrefixes)

	subnets := &schema.Set{
		F: resourceAzureSubnetHash,
	}

	for _, subnet := range *vnet.Subnets {
		s := map[string]interface{}{}

		s["name"] = *subnet.Name
		s["address_prefix"] = *subnet.SubnetPropertiesFormat.AddressPrefix
		if subnet.SubnetPropertiesFormat.NetworkSecurityGroup != nil {
			s["security_group"] = *subnet.SubnetPropertiesFormat.NetworkSecurityGroup.ID
		}

		subnets.Add(s)
	}
	d.Set("subnet", subnets)

	if vnet.DhcpOptions != nil && vnet.DhcpOptions.DNSServers != nil {
		dnses := []string{}
		for _, dns := range *vnet.DhcpOptions.DNSServers {
			dnses = append(dnses, dns)
		}
		d.Set("dns_servers", dnses)
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmVirtualNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vnetClient := meta.(*ArmClient).vnetClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["virtualNetworks"]

	_, err = vnetClient.Delete(resGroup, name, make(chan struct{}))

	return err
}

func getVirtualNetworkProperties(d *schema.ResourceData, meta interface{}) *network.VirtualNetworkPropertiesFormat {
	// first; get address space prefixes:
	prefixes := []string{}
	for _, prefix := range d.Get("address_space").([]interface{}) {
		prefixes = append(prefixes, prefix.(string))
	}

	// then; the dns servers:
	dnses := []string{}
	for _, dns := range d.Get("dns_servers").([]interface{}) {
		dnses = append(dnses, dns.(string))
	}

	// then; the subnets:
	subnets := []network.Subnet{}
	if subs := d.Get("subnet").(*schema.Set); subs.Len() > 0 {
		for _, subnet := range subs.List() {
			subnet := subnet.(map[string]interface{})

			name := subnet["name"].(string)
			log.Printf("[INFO] setting subnets inside vNet, processing %q", name)
			//since subnets can also be created outside of vNet definition (as root objects)
			// do a GET on subnet properties from the server before setting them
			resGroup := d.Get("resource_group_name").(string)
			vnetName := d.Get("name").(string)
			subnetObj := getExistingSubnet(resGroup, vnetName, name, meta)
			log.Printf("[INFO] Completer GET of Subnet props ")

			prefix := subnet["address_prefix"].(string)
			secGroup := subnet["security_group"].(string)
			routeTable := subnet["route_table"].(string)

			if name != "" {
				log.Printf("[INFO] setting SUBNETNAME TO %q", name)
				subnetObj.Name = &name
			}
			if subnetObj.SubnetPropertiesFormat == nil {
				subnetObj.SubnetPropertiesFormat = &network.SubnetPropertiesFormat{}
			}
			if subnetObj.SubnetPropertiesFormat.AddressPrefix == nil {
				subnetObj.SubnetPropertiesFormat.AddressPrefix = &prefix
			}

			if secGroup != "" {
				subnetObj.SubnetPropertiesFormat.NetworkSecurityGroup = &network.SecurityGroup{
					ID: &secGroup,
				}
			}

			if routeTable != "" {
				subnetObj.SubnetPropertiesFormat.RouteTable = &network.RouteTable{
					ID: &routeTable,
				}
			}

			subnets = append(subnets, subnetObj)
		}
	}

	// finally; return the struct:
	return &network.VirtualNetworkPropertiesFormat{
		AddressSpace: &network.AddressSpace{
			AddressPrefixes: &prefixes,
		},
		DhcpOptions: &network.DhcpOptions{
			DNSServers: &dnses,
		},
		Subnets: &subnets,
	}
}

func resourceAzureSubnetHash(v interface{}) int {
	m := v.(map[string]interface{})
	subnet := m["name"].(string) + m["address_prefix"].(string)
	if securityGroup, present := m["security_group"]; present {
		subnet = subnet + securityGroup.(string)
	}
	return hashcode.String(subnet)
}

func getExistingSubnet(resGroup string, vnetName string, subnetName string, meta interface{}) network.Subnet {
	//attempt to retrieve existing subnet from the server
	existingSubnet := network.Subnet{}
	subnetClient := meta.(*ArmClient).subnetClient
	resp, err := subnetClient.Get(resGroup, vnetName, subnetName, "")

	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return existingSubnet
		}
	}

	existingSubnet.SubnetPropertiesFormat = &network.SubnetPropertiesFormat{}
	existingSubnet.SubnetPropertiesFormat.AddressPrefix = resp.SubnetPropertiesFormat.AddressPrefix

	if resp.SubnetPropertiesFormat.NetworkSecurityGroup != nil {
		existingSubnet.SubnetPropertiesFormat.NetworkSecurityGroup = resp.SubnetPropertiesFormat.NetworkSecurityGroup
	}

	if resp.SubnetPropertiesFormat.RouteTable != nil {
		existingSubnet.SubnetPropertiesFormat.RouteTable = resp.SubnetPropertiesFormat.RouteTable
	}

	if resp.SubnetPropertiesFormat.IPConfigurations != nil {
		ips := make([]string, 0, len(*resp.SubnetPropertiesFormat.IPConfigurations))
		for _, ip := range *resp.SubnetPropertiesFormat.IPConfigurations {
			ips = append(ips, *ip.ID)
		}

		existingSubnet.SubnetPropertiesFormat.IPConfigurations = resp.SubnetPropertiesFormat.IPConfigurations
	}

	return existingSubnet
}

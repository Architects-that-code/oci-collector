package network

import (
	"context"
	"fmt"
	"sync"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type VcnCollector struct {
	Region          string          `json:"region"`
	CompartmentName string          `json:"compartmentname"`
	VCN             []core.Vcn      `json:"vcn"`
	Subnets         SubnetCollector `json:"subnets"`
}
type SubnetCollector struct {
	Subnet []core.Subnet
}

func GetAllVcn(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartments []identity.Compartment, networkFetch bool, networkCIDRFetch bool,
	networkInventoryFetch bool) {

	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	var allVCN []VcnCollector

	var wg sync.WaitGroup
	wg.Add(len(regions))
	var regionalSlices = make(chan []VcnCollector, len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			for _, compartment := range compartments {
				vcn := GetVCN(client, compartment, *region.RegionName)

				var subnets []core.Subnet
				for _, v := range vcn {
					subnets = append(subnets, GetSubnets(client, v, *region.RegionName)...)
				}

				if len(vcn) > 0 {
					var v = VcnCollector{
						Region:          *region.RegionName,
						CompartmentName: *compartment.Name,
						VCN:             vcn,
						Subnets:         SubnetCollector{Subnet: subnets},
					}
					allVCN = append(allVCN, v)
				}
				//fmt.Printf("comp %v", *compartment.Name)
			}
			regionalSlices <- allVCN
		}(region)
	}
	wg.Wait()

	fmt.Printf("\n\t Total vcn: %v\n", len(allVCN))

	//fmt.Printf("allVCN: %v\n", allVCN)
	for _, vc := range allVCN {

		fmt.Printf("DisplayName: %v CIDR: %v REGION: %v   COMP: %v \n", *vc.VCN[0].DisplayName, *vc.VCN[0].CidrBlock, vc.Region, vc.CompartmentName)
		for _, subnet := range vc.Subnets.Subnet {
			fmt.Printf("\tSubnet: %v\n", *subnet.DisplayName)
		}

		//fmt.Printf("VCN: %v\n", *vc.VCN[0].DisplayName)
	}
}

func getCompName(compartment []identity.Compartment, compartments []identity.Compartment) {
	fmt.Printf("comp: %v ", compartment)
}

func GetVCN(client core.VirtualNetworkClient, compartment identity.Compartment, region string) []core.Vcn {
	client.SetRegion(region)

	req := core.ListVcnsRequest{SortOrder: core.ListVcnsSortOrderAsc,
		CompartmentId:  compartment.Id,
		LifecycleState: core.VcnLifecycleStateAvailable}

	resp, err := client.ListVcns(context.Background(), req)

	helpers.FatalIfError(err)
	//fmt.Printf("GetVCNs: %v\n", resp.Items)

	return resp.Items
}

func GetSubnets(client core.VirtualNetworkClient, vcn core.Vcn, region string) []core.Subnet {
	client.SetRegion(region)
	req := core.ListSubnetsRequest{
		CompartmentId:  vcn.CompartmentId,
		LifecycleState: core.SubnetLifecycleStateAvailable,
		VcnId:          vcn.Id,
	}

	resp, err := client.ListSubnets(context.Background(), req)

	helpers.FatalIfError(err)

	//fmt.Printf("GetSubnets: %v\n", resp.Items)

	return resp.Items
}

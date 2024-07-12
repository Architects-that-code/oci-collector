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

func GetAllVcn(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartments []identity.Compartment, networkFetch bool, networkCIDRFetch bool,
	networkInventoryFetch bool) {

	client, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	var allVCN []core.Vcn

	var wg sync.WaitGroup
	wg.Add(len(regions))
	var regionalSlices = make(chan []core.Vcn, len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			for _, compartment := range compartments {
				vcn := GetVCN(client, compartment, *region.RegionName)
				allVCN = append(allVCN, vcn...)
			}
			regionalSlices <- allVCN
		}(region)
	}
	wg.Wait()

	fmt.Printf("Total vcn: %v\n", len(allVCN))

	//fmt.Printf("allVCN: %v\n", allVCN)
	for _, vcn := range allVCN {

		fmt.Printf("DisplayName %v CIDR %v \n", *vcn.DisplayName, *vcn.CidrBlock)

	}
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

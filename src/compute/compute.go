package compute

import (
	"context"
	"fmt"
	"sync"

	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func RunCompute(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartments []identity.Compartment) {
	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	fmt.Println("in RunCompute")

	//loop thru regionsconst

	//		in region loop thru compartments
	//TODO: ADD turbonium to region and compartments

	var allInstances []core.Instance
	var wg sync.WaitGroup
	wg.Add(len(regions))
	var regionalSlices = make(chan []core.Instance, len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			for _, compartment := range compartments {
				instances := GetInstances(client, compartment, *region.RegionName)
				allInstances = append(allInstances, instances...)
				//fmt.Printf("region: \t%v  \tcomp:%v: \t\t%v\n", *region.RegionName, *compartment.Name, len(instances))
			}
			regionalSlices <- allInstances
		}(region)
	}
	wg.Wait()

	fmt.Printf("Total instances: %v\n", len(allInstances))

	//fmt.Printf("allInstances: %v\n", allInstances)
	for _, instance := range allInstances {
		fmt.Printf("allInstances: Region: %v InstanceShape: %v Cpus %v Mem %v \n", *instance.Region, *instance.Shape, *instance.ShapeConfig.Ocpus, *instance.ShapeConfig.MemoryInGBs)
		fmt.Printf("tags: freeform: %v   defined: %v \n", instance.FreeformTags, instance.DefinedTags)
	}

	//fmt.Printf("all instannces %v\n", allInstances)
}

func GetInstances(client core.ComputeClient, compartment identity.Compartment, region string) []core.Instance {
	client.SetRegion(region)
	fmt.Printf("Checking: Region: %v\t Compartment: %v\n", region, *compartment.Name)
	req := core.ListInstancesRequest{
		CompartmentId:  compartment.Id,
		LifecycleState: core.InstanceLifecycleStateRunning,
	}

	// Send the request using the service client
	resp, err := client.ListInstances(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.

	return resp.Items
}

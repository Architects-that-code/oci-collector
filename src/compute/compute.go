package compute

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type InstanceGroups struct {
	Region      string
	Compartment string
	Instance    []core.Instance
}

func RunCompute(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartments []identity.Compartment) {
	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	fmt.Println("in RunCompute")

	//loop thru regionsconst

	//		in region loop thru compartments
	//TODO: ADD turbonium to region and compartments

	var allInstances []InstanceGroups
	var wg sync.WaitGroup
	wg.Add(len(regions))
	var regionalSlices = make(chan []InstanceGroups, len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			for _, compartment := range compartments {
				instances := GetInstances(client, compartment, *region.RegionName)
				allInstances = append(allInstances, InstanceGroups{Region: *region.RegionName, Compartment: *compartment.Name, Instance: instances})
				//fmt.Printf("region: \t%v  \tcomp:%v: \t\t%v\n", *region.RegionName, *compartment.Name, len(instances))
			}
			regionalSlices <- allInstances
		}(region)
	}
	wg.Wait()

	//fmt.Printf("Total instances: %v\n", len(allInstances))

	//fmt.Printf("allInstances: %v\n", allInstances)
	sort.Slice(allInstances, func(i, j int) bool {
		return len(allInstances[i].Instance) > len(allInstances[j].Instance)
	})
	fmt.Println("ONLY PRINTING for ACTIVE instances per region/compartment")
	for _, instanceGroup := range allInstances {
		//fmt.Printf("allInstances: Region: %v InstanceShape: %v Cpus %v Mem %v \n", &instanceGroup.Region, *&instanceGroup.Instance.Shape, *instance.ShapeConfig.Ocpus, *instance.ShapeConfig.MemoryInGBs)
		//fmt.Printf("tags: freeform: %v   defined: %v \n", instance.FreeformTags, instance.DefinedTags)

		if len(instanceGroup.Instance) > 0 {
			fmt.Printf("all instances: Region: %v Compartment: %v  NumInstance: %v \n", instanceGroup.Region, instanceGroup.Compartment, len(instanceGroup.Instance))
			for _, instance := range instanceGroup.Instance {
				//fmt.Printf("\tInstance: %v\tShape: %v\tCpus: %v\tMem: %v\tTags: %v\n", *instance.DisplayName, *instance.Shape, *instance.ShapeConfig.Ocpus, *instance.ShapeConfig.MemoryInGBs, *instance.DefinedTags)
				fmt.Printf("DisplayName: %v\t Shape: %v \t tags: freeform: %v\t   defined: %v \t\n", *instance.DisplayName, *instance.Shape, instance.FreeformTags, instance.DefinedTags)
			}
		}
		// can i sort by size

	}

	//fmt.Printf("all instannces %v\n", allInstances)
}

func GetInstances(client core.ComputeClient, compartment identity.Compartment, region string) []core.Instance {
	client.SetRegion(region)
	fmt.Printf("Checking: Region: %v\t Compartment: %v\t\n", region, *compartment.Name)
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

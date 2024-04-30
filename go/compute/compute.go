package compute

import (
	"context"
	"fmt"

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

	for _, compartment := range compartments {

		for _, region := range regions {

			GetInstances(client, compartment, *region.RegionName)

		}

	}
}

func GetInstances(client core.ComputeClient, compartment identity.Compartment, region string) []core.Instance {
	client.SetRegion(region)
	req := core.ListInstancesRequest{
		CompartmentId: compartment.Id,
	}

	// Send the request using the service client
	resp, err := client.ListInstances(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Printf("results in comp: \t%v  \treg:%v: \t%v\n", *compartment.Name, region, len(resp.Items))
	return resp.Items
}

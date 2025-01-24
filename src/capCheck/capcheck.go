package capacity

import (
	config "check-limits/config"
	"context"
	"fmt"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

// check for a shape (ocpu/memory) in a region - across all ad
// check for a shape (ocpu/memory) in a region - across all regions / ad
// provider, client, regions, tenancyID, compartments, *capacityFetch, *capacityShapeOCPUs, *capacityShapeMemory
func Check(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartment []identity.Compartment, capacityFetch bool, capacityShapeOCPUs int, capacityShapeMemory int, capacityShapeType string) {

	//for all regions loop thru region
	// for each region get ads
	// for each ad loop
	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	instanceType := makeInstanceShape(capacityShapeType)

	var adsAll []identity.AvailabilityDomain
	var wg sync.WaitGroup
	wg.Add(len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			//fmt.Printf("Region: %v\n", *region.RegionName)
			client.SetRegion(*region.RegionName)
			ads := config.GetADs(tenancyID, client)
			adsAll = append(adsAll, ads...)
			//fmt.Printf("ads: %v\n", ads)
			for _, ad := range ads {

				//fmt.Printf("ad: %v\n", *ad.Name)
				config := core.CapacityReportInstanceShapeConfig{
					Ocpus:       common.Float32(float32(capacityShapeOCPUs)),
					MemoryInGBs: common.Float32(float32(capacityShapeMemory)),
					//Nvmes:       new(int),
				}
				sadetails := make([]core.CreateCapacityReportShapeAvailabilityDetails, 1)

				sadetails[0] = core.CreateCapacityReportShapeAvailabilityDetails{
					InstanceShape: common.String(instanceType),
					//FaultDomain:         new(string),
					InstanceShapeConfig: &config,
				}

				ccrd := core.CreateComputeCapacityReportDetails{
					CompartmentId:       &tenancyID,
					AvailabilityDomain:  common.String(*ad.Name),
					ShapeAvailabilities: sadetails,
				}

				CreateComputeCapacityReport(context.Background(), provider, ccrd, region)

			}
		}(region)
	}
	wg.Wait()

}

func makeInstanceShape(capacityShapeType string) string {
	// if type is either E4, or E3 or E5 or A1 return "VM.Standard.{}.Flex"
	var shape string
	if capacityShapeType == "E4" || capacityShapeType == "E3" || capacityShapeType == "E5" || capacityShapeType == "A1" || capacityShapeType == "A2" {
		shape = "VM.Standard." + capacityShapeType + ".Flex"
	}
	if capacityShapeType == "X9" {
		shape = "VM.Standard3.Flex"
	}
	return shape
}

func CreateComputeCapacityReport(ctx context.Context, provider common.ConfigurationProvider, reportDetails core.CreateComputeCapacityReportDetails, region identity.RegionSubscription) (core.CreateComputeCapacityReportResponse, error) {
	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	client.SetRegion(*region.RegionName)
	helpers.FatalIfError(err)

	req := core.CreateComputeCapacityReportRequest{
		CreateComputeCapacityReportDetails: core.CreateComputeCapacityReportDetails{
			CompartmentId:       reportDetails.CompartmentId,
			AvailabilityDomain:  reportDetails.AvailabilityDomain,
			ShapeAvailabilities: reportDetails.ShapeAvailabilities,
			// Add other details from reportDetails here
		},
	}

	resp, err := client.CreateComputeCapacityReport(ctx, req)
	helpers.FatalIfError(err)
	fmt.Printf("\n")
	fmt.Printf(*region.RegionName)
	fmt.Printf("\tshape: %v\n", *resp.ShapeAvailabilities[0].InstanceShape)
	fmt.Printf("\t\tocpu: %v\n", *resp.ShapeAvailabilities[0].InstanceShapeConfig.Ocpus)
	fmt.Printf("\t\tmem:  %v\n", *resp.ShapeAvailabilities[0].InstanceShapeConfig.MemoryInGBs)

	fmt.Printf("\t\tavalabile? %v\n", resp.ShapeAvailabilities[0].AvailabilityStatus)
	if resp.ShapeAvailabilities[0].AvailableCount != nil {
		fmt.Printf("\t\t\tcapacity: %v\n", *resp.ShapeAvailabilities[0].AvailableCount)
	}

	return resp, nil
}

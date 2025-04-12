package capacity

import (
	config "check-limits/config"
	"context"
	"fmt"
	"strings"
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

func CheckFAMILY(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartment []identity.Compartment, capacityFetch bool, capacityShapeOCPUs int, capacityShapeMemory int, capacityShapeType string) {

	instanceType := makeInstanceFAMILYShapes(strings.ToUpper(capacityShapeType))

	//for all regions loop thru region
	// for each region get ads
	// for each ad loop thru the instance types to build up SET of types for
	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

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

				config := core.CapacityReportInstanceShapeConfig{
					Ocpus:       common.Float32(float32(capacityShapeOCPUs)),
					MemoryInGBs: common.Float32(float32(capacityShapeMemory)),
					//Nvmes:       new(int),
				}
				sadetails := make([]core.CreateCapacityReportShapeAvailabilityDetails, len(instanceType))
				for i := 0; i < len(instanceType); i++ {
					sadetails[i] = core.CreateCapacityReportShapeAvailabilityDetails{
						InstanceShape: common.String(instanceType[i]),
						//FaultDomain:         new(string),
						InstanceShapeConfig: &config}
				}
				//fmt.Printf("sadetails: %v\n", sadetails)
				/*
					sadetails[0] = core.CreateCapacityReportShapeAvailabilityDetails{
						InstanceShape: common.String("VM.Standard.E3.Flex"),
						//FaultDomain:         new(string),
						InstanceShapeConfig: &config,
					}
				*/
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
	if capacityShapeType == "E4" || capacityShapeType == "E3" || capacityShapeType == "E5" || capacityShapeType == "A1" || capacityShapeType == "A2" || capacityShapeType == "E6" {
		shape = "VM.Standard." + capacityShapeType + ".Flex"
	}
	if capacityShapeType == "X9" {
		shape = "VM.Standard3.Flex"
	}
	return shape
}

func makeInstanceFAMILYShapes(capacityShapeType string) []string {
	var result []string
	switch capacityShapeType {
	case "AMD":
		result = []string{"VM.Standard.E4.Flex", "VM.Standard.E3.Flex", "VM.Standard.E5.Flex", "VM.Standard.E6.Flex"}
	case "INTEL":
		result = []string{"VM.Standard3.Flex"}
	case "ARM":
		result = []string{"VM.Standard.A1.Flex", "VM.Standard.A2.Flex"}
	default:
		fmt.Println("Invalid capacity shape type")
	}
	return result
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
	for i, a := range resp.ComputeCapacityReport.ShapeAvailabilities {
		fmt.Printf("\n")
		fmt.Printf(*region.RegionName)
		fmt.Printf("\tshape: %v\n", *resp.ShapeAvailabilities[i].InstanceShape)
		fmt.Printf("\t\tocpu: %v\n", *resp.ShapeAvailabilities[i].InstanceShapeConfig.Ocpus)
		fmt.Printf("\t\tmem:  %v\n", *resp.ShapeAvailabilities[i].InstanceShapeConfig.MemoryInGBs)
		fmt.Printf("\t\tad:   %v\n", *reportDetails.AvailabilityDomain)

		fmt.Printf("\t\tavalabile? %v\n", a.AvailabilityStatus)

		if a.AvailableCount != nil {
			fmt.Printf("\t\t\tcapacity: %v\n", *resp.ShapeAvailabilities[i].AvailableCount)
		}

	}
	//fmt.Printf("resp: %v\n", resp)

	return resp, nil
}

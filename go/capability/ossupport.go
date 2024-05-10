package capability

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func OSSupport(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartment []identity.Compartment, capacityFetch bool, capacityShapeType string) {
	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	ctx := context.Background()
	/*
		listShapesRequest := core.ListShapesRequest{
			CompartmentId: &tenancyID,
		}
		listShapesResponse, err := client.ListShapes(ctx, listShapesRequest)
		helpers.FatalIfError(err)
	*/

	// List available images
	listImagesRequest := core.ListImagesRequest{
		CompartmentId: &tenancyID,
		Shape:         common.String(makeInstanceShape(capacityShapeType)),
	}
	listImagesResponse, err := client.ListImages(ctx, listImagesRequest)
	helpers.FatalIfError(err)
	/*
		for _, shape := range listShapesResponse.Items {
			fmt.Println("Shape:", *shape.Shape)
			for _, image := range listImagesResponse.Items {

					if *image.Shape == *shape.Shape {
						fmt.Println("  Supported OS:", *image.OperatingSystem)
					}

			}

		}
	*/

	for _, image := range listImagesResponse.Items {

		fmt.Printf("Image: Diplay  %v OS: %v  OSVersion: %v \n", *image.DisplayName, *image.OperatingSystem, *image.OperatingSystemVersion)

	}
	//fmt.Printf("Shapes: %v\n", len(listShapesResponse.Items))
	fmt.Printf("Images: %v\n", len(listImagesResponse.Items))
	//fmt.Printf("full response: %v\n", listImagesResponse)
}
func makeInstanceShape(capacityShapeType string) string {
	// if type is either E4, or E3 or E5 or A1 return "VM.Standard.{}.Flex"
	var shape string
	if capacityShapeType == "E4" || capacityShapeType == "E3" || capacityShapeType == "E5" || capacityShapeType == "A1" {
		shape = "VM.Standard." + capacityShapeType + ".Flex"
	}
	if capacityShapeType == "X9" {
		shape = "VM.Standard3.Flex"
	}
	return shape
}

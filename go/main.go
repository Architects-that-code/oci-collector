package main

import (
	"check-limits/configs"
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/limits"
	"gopkg.in/yaml.v2"
	"log/slog"
	"os"
)

// refactor to create CLI using OCI SDK for Go to interact with the OCI API. Use local auth configs but let user specify
// profile to use. The CLI should list the subscribed regions available to the specified profile and identify all the compartments and then loop thru each compartment in each region to query for
// the limits for each service. The CLI should output the limits to a file in the limits directory in the current working directory. The file should be named
func main() {
	err, config := getConfig()
	if err != nil {
		//fmt.Printf("%+v\n", err)
		slog.Info("%+v\n", err)
		os.Exit(1)
	}

	// Parse command line arguments
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: 'check-oci-limits  something' (if you do not pass in any extra arg it only prints your configs info")
		fmt.Println("Using profile:", config.ProfileName)
		fmt.Printf("Config: %v\n", config.ConfigPath)
		return
	}

	fmt.Println("Using profile:", config.ProfileName)
	fmt.Printf("Config: %v\n", config.ConfigPath)
	printSpace()
	provider := common.CustomProfileConfigProvider(config.ConfigPath, config.ProfileName)
	slog.Debug("provider: %v\n", provider)

	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	printSpace()
	helpers.FatalIfError(err)
	slog.Debug("client %v\n", client)

	tenancyID, err := provider.TenancyOCID()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	slog.Debug("TenancyOCID: %v\n", tenancyID)

	ads := getADs(tenancyID, err, client)
	printSpace()
	//fmt.Printf("ADs: %v\n", ads)
	slog.Debug("ads: ", ads)

	// getallregions
	// getall compartments

	compartments := getCompartments(err, client, tenancyID)
	for _, comp := range compartments {
		fmt.Printf("Compartment Name: %v CompartmentID: %v\n", *comp.Name, *comp.CompartmentId)
	}
	regions := getRegions(err, client, tenancyID)
	for _, region := range regions {
		fmt.Printf("Region: %v\n", *region.RegionName)
	}
	printSpace()
	slog.Debug("List of regions:", regions)

	limitsClient, err := limits.NewLimitsClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	printSpace()
	slog.Debug("limitsClient: %v\n", limitsClient)
	//for _, region := range []string{"us-ashburn-1"} {
	//	reg := region

	for _, region := range regions {
		reg := *region.RegionName

		services := getServices(limitsClient, err, tenancyID, reg)

		for _, s := range services.Items {
			svc := s.Name

			vals := getLimitsForService(err, limitsClient, tenancyID, *svc)
			for _, v := range vals.Items {
				limitName := v.Name
				ad := v.AvailabilityDomain
				var avails = limits.GetResourceAvailabilityResponse{}
				if ad == nil {
					avails = getLimitsAvailRegionScoped(err, limitsClient, tenancyID, *svc, *v.Name)
				} else {
					avails = getLimitAvailADScoped(err, limitsClient, tenancyID, *svc, *v.Name, *ad)
				}
				var avail *int64
				avail = avails.Available
				if avail == nil {
					avail = &[]int64{0}[0] //this is gross
				}
				var used *int64
				used = avails.Used
				if used == nil {
					used = &[]int64{0}[0]
				}

				fmt.Printf("region: %v service: %v valLimitName: %v avail: %v used: %v\n", reg, *svc, *limitName, *avail, *used)
			}

		}
	}

}
func getLimitsAvailRegionScoped(err error, limitsClient limits.LimitsClient, compartment string, svc string, limitName string) limits.GetResourceAvailabilityResponse {
	req := limits.GetResourceAvailabilityRequest{
		ServiceName:     &svc,
		LimitName:       &limitName,
		CompartmentId:   &compartment,
		RequestMetadata: common.RequestMetadata{},
	}
	avail, err := limitsClient.GetResourceAvailability(
		context.Background(), req)
	helpers.FatalIfError(err)
	return avail
}
func getLimitAvailADScoped(err error, limitsClient limits.LimitsClient, compartment string, svc string, limitName string, ad string) limits.GetResourceAvailabilityResponse {
	var req = limits.GetResourceAvailabilityRequest{}
	req = limits.GetResourceAvailabilityRequest{
		ServiceName:        &svc,
		LimitName:          &limitName,
		CompartmentId:      &compartment,
		AvailabilityDomain: &ad,
		RequestMetadata:    common.RequestMetadata{},
	}

	avail, err := limitsClient.GetResourceAvailability(
		context.Background(), req)

	helpers.FatalIfError(err)
	return avail
}

func getLimitDefs(err error, limitsClient limits.LimitsClient, tenancyID string, svc string) limits.ListLimitDefinitionsResponse {
	defs, err := limitsClient.ListLimitDefinitions(context.Background(), limits.ListLimitDefinitionsRequest{
		CompartmentId:   &tenancyID,
		ServiceName:     &svc,
		RequestMetadata: common.RequestMetadata{},
	})
	helpers.FatalIfError(err)
	return defs
}

func getLimitsForService(err error, limitsClient limits.LimitsClient, tenancyID string, svc string) limits.ListLimitValuesResponse {
	vals, err := limitsClient.ListLimitValues(context.Background(), limits.ListLimitValuesRequest{
		CompartmentId:   &tenancyID,
		ServiceName:     &svc,
		RequestMetadata: common.RequestMetadata{},
	})
	helpers.FatalIfError(err)
	return vals
}

func getServices(limitsClient limits.LimitsClient, err error, tenancyID string, region string) limits.ListServicesResponse {
	limitsClient.SetRegion(region)
	printSpace()
	slog.Debug("limitsClientUPDATED: \n", limitsClient.Endpoint())
	services, err := limitsClient.ListServices(context.Background(), limits.ListServicesRequest{
		CompartmentId:   &tenancyID,
		SortBy:          "",
		SortOrder:       "",
		Limit:           nil,
		Page:            nil,
		OpcRequestId:    nil,
		RequestMetadata: common.RequestMetadata{},
	})
	helpers.FatalIfError(err)
	return services
}

func printSpace() {
	fmt.Println("")
}

func getADs(tenancyID string, err error, client identity.IdentityClient) identity.ListAvailabilityDomainsResponse {
	request := identity.ListAvailabilityDomainsRequest{
		CompartmentId: &tenancyID,
	}
	r, err := client.ListAvailabilityDomains(context.Background(), request)
	helpers.FatalIfError(err)
	return r
}

func getCompartments(err error, client identity.IdentityClient, tenancyID string) []identity.Compartment {
	resComp, err := client.ListCompartments(context.Background(), identity.ListCompartmentsRequest{
		AccessLevel:            identity.ListCompartmentsAccessLevelAny,
		CompartmentId:          &tenancyID,
		CompartmentIdInSubtree: common.Bool(true),
		SortBy:                 identity.ListCompartmentsSortByName,
		SortOrder:              identity.ListCompartmentsSortOrderAsc,
		LifecycleState:         identity.CompartmentLifecycleStateActive,
		Limit:                  common.Int(208),
	})
	helpers.FatalIfError(err)
	//fmt.Printf("List of compartments: %v", resComp.Items)
	return resComp.Items
}

func getRegions(err error, client identity.IdentityClient, tenancyID string) []identity.RegionSubscription {
	reqReg, err := client.ListRegionSubscriptions(context.Background(), identity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyID,
	})
	helpers.FatalIfError(err)
	//fmt.Printf("List of regions: %v", reqReg.Items)
	return reqReg.Items
}

/*
func processLimits(c) {
	panic("unimplemented")
}

func getRegions(c) {
	panic("unimplemented")
}
*/

func getConfig() (error, configs.Config) {
	data, err := os.ReadFile("slurper.yaml")
	if err != nil {
		// handle error
	}

	var config configs.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		// handle error
	}
	return err, config
}

/*
func getRegions(profileName string, configs Config) ([]string, error) {
	// Create the identity client

	identityClient := common.CustomProfileConfigProvider(configs.ConfigPath, profileName)

	// Get the tenancy ID
	tenancyID, err := identityClient.GetTenancy(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	// Get the list of subscribed regions
	regionSubscriptions, err := identityClient.ListRegionSubscriptions(context.Background(), tenancyID)
	if err != nil {
		return nil, err
	}

	regionNames := make([]string, len(regionSubscriptions))

	for i, regionSubscription := range regionSubscriptions {
		regionNames[i] = regionSubscription.RegionName
	}

	// Return the list of RegionName values
	return regionNames, nil
}
*/
/*

func processLimits(profileName string, region string) {
	// Create the limits client
	limitsClient, err := limits.NewLimitsClientWithConfiguration(
		context.Background(),
		common.DefaultConfig(profileName),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the region
	limitsClient.BaseClient.SetRegion(region)

	// Get the list of services
	services, err := limitsClient.ListServices(context.Background(), tenancyID)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create the output directory
	outputDir := filepath.Join("limits", region)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println(err)
		return
	}

	// Process each service
	for _, service := range services {
		serviceName := service.Name
		fmt.Println("Processing service:", serviceName)

		// Get the list of limit definitions
		limitDefinitions, err := limitsClient.ListLimitDefinitions(context.Background(), tenancyID, serviceName)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Get the list of limit values
		limitValues, err := limitsClient.ListLimitValues(context.Background(), tenancyID, serviceName)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
*/

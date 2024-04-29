package setup

import (
	"check-limits/util"
	"context"
	"os"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"gopkg.in/yaml.v2"
)

func Getcompartments(err error, client identity.IdentityClient, tenancyID string) []identity.Compartment {
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

func GetADs(tenancyID string, err error, client identity.IdentityClient) []identity.AvailabilityDomain {
	adReq := identity.ListAvailabilityDomainsRequest{
		CompartmentId: &tenancyID,
	}
	adResp, err := client.ListAvailabilityDomains(context.Background(), adReq)
	helpers.FatalIfError(err)
	return adResp.Items
}
func GetALLADdata(err error, client identity.IdentityClient, tenancyID string, regions []identity.RegionSubscription) []identity.AvailabilityDomain {
	var ads []identity.AvailabilityDomain
	for _, region := range regions {
		//fmt.Printf("Region: %v\n", *region.RegionName)
		client.SetRegion(*region.RegionName)
		ads = GetADs(tenancyID, err, client)
		//fmt.Printf("ads: %v\n", ads)
	}
	return ads
}

func Getregions(err error, client identity.IdentityClient, tenancyID string) []identity.RegionSubscription {
	reqReg, err := client.ListRegionSubscriptions(context.Background(), identity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyID,
	})
	helpers.FatalIfError(err)
	//fmt.Printf("List of regions: %v", reqReg.Items)
	return reqReg.Items
}

func Getconfig() (error, Config) {
	data, err := os.ReadFile("slurper.yaml")
	helpers.FatalIfError(err)

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		// handle error
	}
	return err, config
}

type Config struct {
	ConfigPath  string `yaml:"configPath"`
	ProfileName string `yaml:"profileName"`
}

func GetProvider(config Config) common.ConfigurationProvider {
	provider := common.CustomProfileConfigProvider(config.ConfigPath, config.ProfileName)
	return provider
}

func GetIdentityClient(provider common.ConfigurationProvider) (identity.IdentityClient, error) {
	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	return client, err
}

func CommonSetup(err error, client identity.IdentityClient, tenancyID string, fetchADs bool) ([]identity.RegionSubscription, []identity.Compartment, []identity.AvailabilityDomain) {
	var wgDataPrep = sync.WaitGroup{}
	wgDataPrep.Add(2)
	var compartments []identity.Compartment
	go func() {
		defer wgDataPrep.Done()
		compartments = Getcompartments(err, client, tenancyID)
		/*
			for _, comp := range compartments {
				fmt.Printf("Compartment Name: %v CompartmentID: %v\n", *comp.Name, *comp.CompartmentId)
			}*/
	}()

	var regions []identity.RegionSubscription
	go func() {
		defer wgDataPrep.Done()
		regions = Getregions(err, client, tenancyID)
		/*
			for _, region := range regions {
				fmt.Printf("Region: %v\n", *region.RegionName)
			}*/
		util.PrintSpace()
		//slog.Debug("List of regions:", regions)
	}()
	wgDataPrep.Wait()
	var ads []identity.AvailabilityDomain
	if fetchADs {
		ads = GetALLADdata(err, client, tenancyID, regions)
	} else {
		ads = nil
	}

	return regions, compartments, ads
}

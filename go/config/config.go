package setup

import (
	"check-limits/util"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"gopkg.in/yaml.v2"
)

func Getcompartments(err error, client identity.IdentityClient, tenancyID string) []identity.Compartment {
	var allCompartments []identity.Compartment
	req := identity.ListCompartmentsRequest{
		AccessLevel:            identity.ListCompartmentsAccessLevelAny,
		CompartmentId:          &tenancyID,
		CompartmentIdInSubtree: common.Bool(true),
		SortBy:                 identity.ListCompartmentsSortByName,
		SortOrder:              identity.ListCompartmentsSortOrderAsc,
		LifecycleState:         identity.CompartmentLifecycleStateActive,
		Limit:                  common.Int(208),
	}
	for {
		respComp, err := client.ListCompartments(context.Background(), req)
		helpers.FatalIfError(err)
		allCompartments = append(allCompartments, respComp.Items...)
		if respComp.OpcNextPage != nil {
			req.Page = respComp.OpcNextPage
		} else {
			break
		}

	}
	//fmt.Printf("List of compartments: %v", resComp.Items)
	return allCompartments
}

func GetADs(tenancyID string, client identity.IdentityClient) []identity.AvailabilityDomain {
	adReq := identity.ListAvailabilityDomainsRequest{
		CompartmentId: &tenancyID,
	}
	adResp, err := client.ListAvailabilityDomains(context.Background(), adReq)
	helpers.FatalIfError(err)
	return adResp.Items
}
func GetALLADdata(client identity.IdentityClient, tenancyID string, regions []identity.RegionSubscription) []identity.AvailabilityDomain {
	var adsAll []identity.AvailabilityDomain
	for _, region := range regions {
		fmt.Printf("Region: %v\n", *region.RegionName)
		client.SetRegion(*region.RegionName)
		ads := GetADs(tenancyID, client)
		adsAll = append(adsAll, ads...)
		fmt.Printf("ads: %v\n", ads)
	}
	return adsAll
}

func Getregions(err error, client identity.IdentityClient, tenancyID string) ([]identity.RegionSubscription, string) {
	reqReg, err := client.ListRegionSubscriptions(context.Background(), identity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyID,
	})
	helpers.FatalIfError(err)
	//fmt.Printf("List of regions: %v", reqReg.Items)

	return reqReg.Items, getHomeRegion(reqReg.Items)
}

func getHomeRegion(regions []identity.RegionSubscription) string {
	for _, region := range regions {
		if *region.IsHomeRegion {
			fmt.Printf("Home Region: %v\n", *region.RegionName)
			return *region.RegionName
		}
	}
	return ""
}

func Getconfig() (error, Config) {
	data, err := os.ReadFile("toolkit-config.yaml")
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
	CSI         string `yaml:"SUPPORT_CSI_NUMBER"`
}

func Prep(config Config) (common.ConfigurationProvider, identity.IdentityClient, string, error) {
	provider := common.CustomProfileConfigProvider(config.ConfigPath, config.ProfileName)
	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	tenancyID, err := provider.TenancyOCID()
	helpers.FatalIfError(err)
	return provider, client, tenancyID, err

}

func CommonSetup(err error, client identity.IdentityClient, tenancyID string, fetchADs bool) ([]identity.RegionSubscription, []identity.Compartment, []identity.AvailabilityDomain, string) {
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
	var homeregion string
	go func() {
		defer wgDataPrep.Done()
		regions, homeregion = Getregions(err, client, tenancyID)

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
		ads = GetALLADdata(client, tenancyID, regions)
	} else {
		ads = nil
	}

	return regions, compartments, ads, homeregion
}

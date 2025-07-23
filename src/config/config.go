package setup

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"gopkg.in/yaml.v2"
)

// GetCompartmentsHeirarchy returns a list of compartments in a tenancy - including the root compartment and all nested compartments
func GetCompartmentsHeirarchy(err error, client identity.IdentityClient, tenancyID string) {

}

func Getcompartments(err error, client identity.IdentityClient, tenancyID string) []identity.Compartment {
	var allCompartments []identity.Compartment
	allCompartments = append(allCompartments, identity.Compartment{Id: &tenancyID,
		Name: common.String("root")})

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
func FDs(tenancyID string, client identity.IdentityClient, ad identity.AvailabilityDomain) []identity.FaultDomain {
	fdreq := identity.ListFaultDomainsRequest{
		CompartmentId:      &tenancyID,
		AvailabilityDomain: ad.Name,
	}
	fdResp, err := client.ListFaultDomains(context.Background(), fdreq)
	helpers.FatalIfError(err)
	fmt.Printf("Fault Domains: %v\n", fdResp)
	return fdResp.Items
}
func GetALLADdata(client identity.IdentityClient, tenancyID string, regions []identity.RegionSubscription) []identity.AvailabilityDomain {
	//start := time.Now()
	//fmt.Print("Fetching ADs\n")
	var adsAll []identity.AvailabilityDomain
	/*
		for _, region := range regions {
			client.SetRegion(*region.RegionName)
			ads := GetADs(tenancyID, client)
			adsAll = append(adsAll, ads...)
			mu.Unlock()
		}
	*/
	/**     start comment here */
	var wg sync.WaitGroup
	wg.Add(len(regions))

	var regionalSlices = make(chan []identity.AvailabilityDomain, len(regions))

	for _, region := range regions {
		go func(region identity.RegionSubscription) {
			defer wg.Done()
			//fmt.Printf("Region: %v\n", *region.RegionName)
			client.SetRegion(*region.RegionName)
			ads := GetADs(tenancyID, client)
			adsAll = append(adsAll, ads...)
			regionalSlices <- ads
		}(region)
	}
	wg.Wait()

	/*  end comment here  */
	//elapsed := time.Since(start)
	//fmt.Printf("Fetching ADs took %s \n", elapsed)
	/*
		allReg := getALLRegions(nil, client)
			fmt.Printf("size of all regions %v ", len(allReg))
			fmt.Printf(" all regions %v ", allReg)
	*/
	return adsAll
}

func getSubscribedRegions(err error, client identity.IdentityClient, tenancyID string) ([]identity.RegionSubscription, string) {
	reqReg, err := client.ListRegionSubscriptions(context.Background(), identity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyID,
	})
	helpers.FatalIfError(err)
	//fmt.Printf("List of subcribed regions:\n %v", reqReg.Items)
	//getALLRegions(err, client)
	return reqReg.Items, getHomeRegion(reqReg.Items)

}
func getALLRegions(err error, client identity.IdentityClient) []identity.Region {
	allReg, err := client.ListRegions(context.Background()) // this gets all POSSIBLE regions -

	helpers.FatalIfError(err)
	//fmt.Printf("\nList of ALL regions: \n %v", allReg.Items)

	return allReg.Items
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

func Getconfig() (Config, error) {
	data, err := os.ReadFile("toolkit-config.yaml")
	helpers.FatalIfError(err)

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		// handle error
	}
	return config, err
}

type Config struct {
	ConfigPath           string `yaml:"configPath"`
	ProfileName          string `yaml:"profileName"`
	UseInstancePrincipal bool   `yaml:"useinstanceprincipal"`
	CSI                  string `yaml:"SUPPORT_CSI_NUMBER"`
	ORG_ID               string `yaml:"ORG_ID"`
	SUBSCRIPTION_ID      string `yaml:"SUBSCRIPTION_ID"`
}

func Prep(config Config) (common.ConfigurationProvider, identity.IdentityClient, string, error) {
	var _provider common.ConfigurationProvider

	if config.UseInstancePrincipal {
		fmt.Println("Using Instance Principal")
		_provider, _ = auth.InstancePrincipalConfigurationProvider()
	} else {
		fmt.Println("Using Config File")
		fmt.Println("Using profile:", config.ProfileName)
		fmt.Printf("Config: %v\n", config.ConfigPath)
		_provider = common.CustomProfileConfigProvider(config.ConfigPath, config.ProfileName)
	}

	client, err := identity.NewIdentityClientWithConfigurationProvider(_provider)
	helpers.FatalIfError(err)
	tenancyID, err := _provider.TenancyOCID()
	helpers.FatalIfError(err)
	return _provider, client, tenancyID, err

}

func CommonSetup(err error, client identity.IdentityClient, tenancyID string) ([]identity.RegionSubscription, []identity.Compartment, []identity.AvailabilityDomain, string) {
	var wgDataPrep = sync.WaitGroup{}
	wgDataPrep.Add(2)
	/*
		go func() {
			defer wgDataPrep.Done()
			possibleRegions := getALLRegions(err, client)
			fmt.Printf("\nList of ALL regions: num: %v  \ndump: %v\n", len(possibleRegions), possibleRegions)
		}()
	*/

	var compartments []identity.Compartment
	go func() {
		defer wgDataPrep.Done()
		compartments = Getcompartments(err, client, tenancyID)

	}()

	var regions []identity.RegionSubscription
	var homeregion string

	go func() {
		defer wgDataPrep.Done()
		regions, homeregion = getSubscribedRegions(err, client, tenancyID)

		/*
			for _, region := range regions {
				fmt.Printf("Region: %v\n", *region.RegionName)
			}*/
		//util.PrintSpace()
		//slog.Debug("List of regions:", regions)
	}()
	wgDataPrep.Wait()
	var ads []identity.AvailabilityDomain

	ads = GetALLADdata(client, tenancyID, regions)

	//getTenancyObj(client, tenancyID, homeregion)

	return regions, compartments, ads, homeregion
}

func getTenancyObj(client identity.IdentityClient, tenancyID string, homeregion string) {
	client.SetRegion(homeregion)
	req := identity.GetTenancyRequest{
		TenancyId: common.String(tenancyID),
	}
	resp, err := client.GetTenancy(context.Background(), req)
	helpers.FatalIfError(err)
	fmt.Printf("Tenancy: %v\n", resp.Tenancy)

}

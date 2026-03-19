package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"gopkg.in/yaml.v2"
	"oci-collector/util"
)

var profileNameOverride string

// SetProfileOverride sets a runtime profile name that overrides profileName in toolkit-config.yaml.
// Passing an empty string clears the override.
func SetProfileOverride(profile string) {
	profileNameOverride = strings.TrimSpace(profile)
}

// GetCompartmentsHeirarchy returns a list of compartments in a tenancy - including the root compartment and all nested compartments
func GetCompartmentsHeirarchy(err error, client identity.IdentityClient, tenancyID string) {

}

func Getcompartments(client identity.IdentityClient, tenancyID string) []identity.Compartment {
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
		util.FatalIfError(err)
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
	util.FatalIfError(err)
	return adResp.Items
}
func FDs(tenancyID string, client identity.IdentityClient, ad identity.AvailabilityDomain) []identity.FaultDomain {
	fdreq := identity.ListFaultDomainsRequest{
		CompartmentId:      &tenancyID,
		AvailabilityDomain: ad.Name,
	}
	fdResp, err := client.ListFaultDomains(context.Background(), fdreq)
	util.FatalIfError(err)
	fmt.Printf("Fault Domains: %v\n", fdResp)
	return fdResp.Items
}
func GetALLADdata(client identity.IdentityClient, tenancyID string, regions []identity.RegionSubscription) []identity.AvailabilityDomain {
	// IMPORTANT: SetRegion mutates client endpoint state.
	// Do this sequentially on a shared client to avoid endpoint races/corruption.
	var adsAll []identity.AvailabilityDomain
	for _, region := range regions {
		client.SetRegion(*region.RegionName)
		ads := GetADs(tenancyID, client)
		adsAll = append(adsAll, ads...)
	}
	return adsAll
}

func getSubscribedRegions(client identity.IdentityClient, tenancyID string) ([]identity.RegionSubscription, string) {
	reqReg, err := client.ListRegionSubscriptions(context.Background(), identity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyID,
	})
	util.FatalIfError(err)
	//fmt.Printf("List of subcribed regions:\n %v", reqReg.Items)
	//getALLRegions(err, client)
	return reqReg.Items, getHomeRegion(reqReg.Items)

}
func getALLRegions(client identity.IdentityClient) []identity.Region {
	allReg, err := client.ListRegions(context.Background()) // this gets all POSSIBLE regions -

	util.FatalIfError(err)
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

// Getconfig reads and parses toolkit-config.yaml into a Config struct.
func Getconfig() (Config, error) {
	data, err := os.ReadFile("toolkit-config.yaml")
	util.FatalIfError(err)

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		// handle error
	}
	if profileNameOverride != "" {
		config.ProfileName = profileNameOverride
	}
	return config, err
}

// Config holds authentication and configuration settings from YAML.
type Config struct {
	ConfigPath           string `yaml:"configPath"`
	ProfileName          string `yaml:"profileName"`
	UseInstancePrincipal bool   `yaml:"useinstanceprincipal"`
	CSI                  string `yaml:"SUPPORT_CSI_NUMBER"`
}

// Prep creates OCI ConfigurationProvider and IdentityClient based on config.
// Supports Instance Principal or config file auth.
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
	util.FatalIfError(err)
	tenancyID, err := _provider.TenancyOCID()
	util.FatalIfError(err)
	return _provider, client, tenancyID, err

}

// CommonSetup fetches subscribed regions, compartments, and availability domains.
// Uses concurrency for efficiency.
func CommonSetup(client identity.IdentityClient, tenancyID string) ([]identity.RegionSubscription, []identity.Compartment, []identity.AvailabilityDomain, string) {
	// Keep setup deterministic and race-free on a shared Identity client.
	// (SetRegion is used later for AD enumeration.)
	compartments := Getcompartments(client, tenancyID)
	regions, homeregion := getSubscribedRegions(client, tenancyID)
	ads := GetALLADdata(client, tenancyID, regions)

	//getTenancyObj(client, tenancyID, homeregion)

	return regions, compartments, ads, homeregion
}

func getTenancyObj(client identity.IdentityClient, tenancyID string, homeregion string) {
	client.SetRegion(homeregion)
	req := identity.GetTenancyRequest{
		TenancyId: common.String(tenancyID),
	}
	resp, err := client.GetTenancy(context.Background(), req)
	util.FatalIfError(err)
	fmt.Printf("Tenancy: %v\n", resp.Tenancy)

}

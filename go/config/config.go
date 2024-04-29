package setup

import (
	"context"
	"os"

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

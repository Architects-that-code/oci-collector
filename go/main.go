package main

import (
	compute "check-limits/compute"
	setup "check-limits/config"
	"check-limits/limits"
	"check-limits/util"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

// refactor to create CLI using OCI SDK for Go to interact with the OCI API. Use local auth configs but let user specify
// profile to use. The CLI should list the subscribed regions available to the specified profile and identify all the compartments and then loop thru each compartment in each region to query for
// the limits for each service. The CLI should output the limits to a file in the limits directory in the current working directory. The file should be named
func main() {

	var (
		usage = `specity the action you want to take:
		pick 1 of the following
		-limits: fetch limits in all region
		-compute: fetch compute active instances in all regions
		-checkconfig: check config file`
	)

	limitsAction := flag.Bool("limits", false, "fetch limits in all regions")
	computeAction := flag.Bool("compute", false, "fetch compute active instances in all regions")
	checkConfigAction := flag.Bool("checkconfig", false, "check config file")

	err, config := setup.Getconfig()
	if err != nil {
		//fmt.Printf("%+v\n", err)
		slog.Info("%+v\n", err)
		os.Exit(1)
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		fmt.Println("Using profile:", config.ProfileName)
		fmt.Printf("Config: %v\n", config.ConfigPath)
		return
	}

	fmt.Printf("os.Args: %v\n", os.Args)

	fmt.Printf("limit flags: %v\n", *limitsAction)
	fmt.Printf("compute flags: %v\n", *computeAction)
	fmt.Printf("check flags: %v\n", *checkConfigAction)

	fmt.Printf("flag.Args: %v\n", flag.Args())
	command := flag.Args()[0]

	// Parse command line arguments

	fmt.Println("Using profile:", config.ProfileName)
	fmt.Printf("Config: %v\n", config.ConfigPath)
	util.PrintSpace()

	provider := setup.GetProvider(config)
	slog.Debug("provider: %v\n", provider)

	client, err := setup.GetIdentityClient(provider)
	util.PrintSpace()
	helpers.FatalIfError(err)
	slog.Debug("client %v\n", client)

	tenancyID, err := provider.TenancyOCID()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	slog.Debug("TenancyOCID: %v\n", tenancyID)

	// getallregions
	// getall compartments

	//start common setup get all compartemnts get all regions
	regions, compartemnts, ads := setup.CommonSetup(err, client, tenancyID, false)
	if compartemnts == nil {
		fmt.Println("compartments is nil")
	} else {

		fmt.Printf("compartments: %v\n", len(compartemnts))
		//fmt.Printf("compartments: %v\n", compartemnts)
	}
	if regions == nil {
		fmt.Println("regions is nil")
	} else {
		fmt.Printf("regions: %v\n", len(regions))
	}
	if ads == nil {
		fmt.Println("ads is nil")
	} else {
		fmt.Printf("ads: %v\n", len(ads))
	}

	grr := runCmd(command, flag.Args()[1:], provider, regions, tenancyID, compartemnts)
	if grr != nil {
		fmt.Println(grr)
		os.Exit(1)
	}

	//for _, region := range []string{"us-ashburn-1"} {
	//	reg := region

	//create datastructures that will hold all results

	//for _, region := range localReg {
	//reg := region

}

func runCmd(command string, s []string, provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartments []identity.Compartment) error {
	fmt.Printf("commandname: %v\n", command)

	switch command {
	case "limit-fetch":
		fmt.Println("fetching limits")
		limits.RunLimits(provider, regions, tenancyID)
		return nil
	case "compute-fetch":
		fmt.Println("fetching compute")
		compute.RunCompute(provider, regions, tenancyID, compartments)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)

	}
}

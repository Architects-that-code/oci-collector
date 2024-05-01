package main

import (
	compute "check-limits/compute"
	setup "check-limits/config"
	peopleresource "check-limits/iam"
	"check-limits/limits"
	"check-limits/util"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/oracle/oci-go-sdk/v65/example/helpers"
)

// refactor to create CLI using OCI SDK for Go to interact with the OCI API. Use local auth configs but let user specify
// profile to use. The CLI should list the subscribed regions available to the specified profile and identify all the compartments and then loop thru each compartment in each region to query for
// the limits for each service. The CLI should output the limits to a file in the limits directory in the current working directory. The file should be named
func main() {

	var (
		usage = `usage: #check-limits 'action' 'activate'
	example: check-limits limits -run

	specity the action you want to take:

		
	limits: fetch limits in all region
	compute: fetch compute active instances in all regions
	config: check config file
	peeps: fetch user counts (-r to show users)
	policies: fetch policy counts (-run to show policies -verbose to show statements)
		`
	)

	limitCmd := flag.NewFlagSet("limits", flag.ExitOnError)
	limitFetch := limitCmd.Bool("run", false, "fetch limits in all regions")

	computeCmd := flag.NewFlagSet("compute", flag.ExitOnError)
	computeFetch := computeCmd.Bool("run", false, "fetch compute active instances in all regions")

	checkCmd := flag.NewFlagSet("config", flag.ExitOnError)
	checkFetch := checkCmd.Bool("run", true, "check config file")

	peopleCmd := flag.NewFlagSet("peeps", flag.ExitOnError)
	peopleFetch := peopleCmd.Bool("run", false, "fetch users")

	policyCmd := flag.NewFlagSet("policies", flag.ExitOnError)
	policyFetch := policyCmd.Bool("run", false, "fetch policy")
	policyVerbose := policyCmd.Bool("verbose", false, "show policies")

	/*
		limitsAction := flag.Bool("limits", false, "fetch limits in all regions")
		computeAction := flag.Bool("compute", false, "fetch compute active instances in all regions")
		checkConfigAction := flag.Bool("checkconfig", false, "check config file")
	*/

	err, config := setup.Getconfig()
	if err != nil {
		//fmt.Printf("%+v\n", err)
		slog.Info("%+v\n", err)
		os.Exit(1)
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		//fmt.Println("Using profile:", config.ProfileName)
		//fmt.Printf("Config: %v\n", config.ConfigPath)
		return
	}

	fmt.Printf("os.Args: %v\n", os.Args)

	fmt.Printf("flag.Args: %v\n", flag.Args())

	// Parse command line arguments

	fmt.Println("Using profile:", config.ProfileName)
	fmt.Printf("Config: %v\n", config.ConfigPath)
	util.PrintSpace()

	switch os.Args[1] {
	case "limits":
		fmt.Println("fetching limits")
		limitCmd.Parse(os.Args[2:])
		fmt.Printf("limitFetch: %v\n", *limitFetch)
		if *limitFetch {
			provider, client, tenancyID, err := setup.Prep(config)
			regions, _, _ := setup.CommonSetup(err, client, tenancyID, false)
			limits.RunLimits(provider, regions, tenancyID)
		} else {
			fmt.Println("add -run to run")
		}

	case "compute":
		fmt.Println("fetching compute")
		computeCmd.Parse(os.Args[2:])
		fmt.Printf("computeFetch: %v\n", *computeFetch)
		if *computeFetch {
			provider, client, tenancyID, err := setup.Prep(config)
			regions, compartments, _ := setup.CommonSetup(err, client, tenancyID, false)
			compute.RunCompute(provider, regions, tenancyID, compartments)
		} else {
			fmt.Println("add -run to run")
		}
	case "peeps":
		fmt.Println("fetching users")
		peopleCmd.Parse(os.Args[2:])
		fmt.Printf("peopleFetch: %v\n", *peopleFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		helpers.FatalIfError(err)
		peopleresource.GetAllPeople(provider, client, tenancyID, *peopleFetch)

	case "policies":
		fmt.Println("fetching policies")
		policyCmd.Parse(os.Args[2:])
		fmt.Printf("policies: %v\n", *policyFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		helpers.FatalIfError(err)
		_, compartments, _ := setup.CommonSetup(err, client, tenancyID, false)
		peopleresource.GetAllPolicies(provider, client, tenancyID, compartments, *policyFetch, *policyVerbose)

	case "config":
		fmt.Println("checking config")
		checkCmd.Parse(os.Args[2:])
		fmt.Printf("checkRun: %v\n", *checkFetch)
		_, client, tenancyID, err := setup.Prep(config)
		regions, compartments, ads := setup.CommonSetup(err, client, tenancyID, false)
		if compartments == nil {
			fmt.Println("compartments is nil")
		} else {

			fmt.Printf("compartments: %v\n", len(compartments))
			//fmt.Printf("compartments: %v\n", compartments)
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

	default:
		fmt.Printf("Invalid command: %v\n", os.Args[1])
	}

	//for _, region := range []string{"us-ashburn-1"} {
	//	reg := region

	//create datastructures that will hold all results

	//for _, region := range localReg {
	//reg := region

}

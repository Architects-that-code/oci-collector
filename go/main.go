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
)

// refactor to create CLI using OCI SDK for Go to interact with the OCI API. Use local auth configs but let user specify
// profile to use. The CLI should list the subscribed regions available to the specified profile and identify all the compartments and then loop thru each compartment in each region to query for
// the limits for each service. The CLI should output the limits to a file in the limits directory in the current working directory. The file should be named
func main() {

	var (
		usage = `usage: #check-limits 'action' 'activate'
		         example: check-limits limits -run

		specity the action you want to take:
		expected 'limits' or 'compute' or 'config' as the first argument and -run (to actually run)
		limits: fetch limits in all region
		compute: fetch compute active instances in all regions
		config: check config file
		`
	)

	limitCmd := flag.NewFlagSet("limits", flag.ExitOnError)
	limitFetch := limitCmd.Bool("run", false, "fetch limits in all regions")

	computeCmd := flag.NewFlagSet("compute", flag.ExitOnError)
	computeFetch := computeCmd.Bool("run", false, "fetch compute active instances in all regions")

	checkCmd := flag.NewFlagSet("config", flag.ExitOnError)
	checkFetch := checkCmd.Bool("run", false, "check config file")

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
		fmt.Println("Using profile:", config.ProfileName)
		fmt.Printf("Config: %v\n", config.ConfigPath)
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
			regions, compartemnts, _ := setup.CommonSetup(err, client, tenancyID, false)
			compute.RunCompute(provider, regions, tenancyID, compartemnts)
		} else {
			fmt.Println("add -run to run")
		}

	case "checkconfig":
		fmt.Println("checking config")
		checkCmd.Parse(os.Args[2:])
		fmt.Printf("checkRun: %v\n", *checkFetch)
		_, client, tenancyID, err := setup.Prep(config)
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

	default:
		fmt.Printf("Invalid command: %v\n", os.Args[1])
	}

	//for _, region := range []string{"us-ashburn-1"} {
	//	reg := region

	//create datastructures that will hold all results

	//for _, region := range localReg {
	//reg := region

}

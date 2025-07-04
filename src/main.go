package main

import (
	billing "check-limits/billing"
	capcheck "check-limits/capCheck"
	"check-limits/capability"
	compute "check-limits/compute"
	setup "check-limits/config"
	peopleresource "check-limits/iam"
	"check-limits/limits"
	network "check-limits/networks"
	oos "check-limits/objectstorage"
	scheduler "check-limits/schedule"
	resourcesearch "check-limits/search"
	supportresources "check-limits/support"
	"slices"
	"strings"

	children "check-limits/childtenancies"
	utils "check-limits/util"
	"flag"
	"fmt"
	"os"

	"github.com/oracle/oci-go-sdk/v65/example/helpers"
)

// refactor to create CLI using OCI SDK for Go to interact with the OCI API. Use local auth configs but let user specify
// profile to use. The CLI should list the subscribed regions available to the specified profile and identify all the compartments and then loop thru each compartment in each region to query for
// the limits for each service. The CLI should output the limits to a file in the limits directory in the current working directory. The file should be named
func main() {
	utils.PrintBanner()

	var (
		usage = `This is designed to be a loose collection of tools to help manage and monitor your OCI tenancy.		
		
	usage: #check-limits 'action' 'activate'
		example: check-limits limits -run

	specity the action you want to take:

		
	limits: fetch limits in all YOUR region
	compute: fetch compute active instances in all YOUR regions
	config: check config file and print basic info
	peeps: fetch user counts (-r to show users)
	policies: fetch policy counts (-run to show policies -verbose to show statements)
	groups: fetch group counts (-run to show groups )
	support: fetch support tickets (-list to show tickets)
	capacity: check capacity in all YOUR regions (-ocpus to specify ocpus -memory to specify memory -type to specify shape type)
	capability: what types of 'things' are available for a shape type (-type to specify shape type)
	children: dealing with child tenancies
	object: fetch object storage info
	billing: fetch billing info
	network: network related info
	search: search for resources created by a user
		`
	)

	limitCmd := flag.NewFlagSet("limits", flag.ExitOnError)
	limitFetch := limitCmd.Bool("run", false, "fetch limits in all regions")
	limitWrite := limitCmd.Bool("write", false, "write limits to file")

	computeCmd := flag.NewFlagSet("compute", flag.ExitOnError)
	computeFetch := computeCmd.Bool("run", false, "fetch compute active instances in all regions")

	networkCmd := flag.NewFlagSet("vcn", flag.ExitOnError)
	networkFetch := networkCmd.Bool("run", false, "fetch all vcn in all regions")
	networkCIDRFetch := networkCmd.Bool("cidr", false, "also fetch CIDR blocks")
	networkInventoryFetch := networkCmd.Bool("ip", false, "fetch IP inventory")

	checkCmd := flag.NewFlagSet("config", flag.ExitOnError)
	checkFetch := checkCmd.Bool("verbose", false, "get more details")

	peopleCmd := flag.NewFlagSet("peeps", flag.ExitOnError)
	peopleFetch := peopleCmd.Bool("run", false, "fetch users")

	policyCmd := flag.NewFlagSet("policies", flag.ExitOnError)
	policyFetch := policyCmd.Bool("run", false, "fetch policy")
	policyVerbose := policyCmd.Bool("verbose", false, "show policies")

	groupCmd := flag.NewFlagSet("groups", flag.ExitOnError)
	groupFetch := groupCmd.Bool("run", false, "fetch groups")
	//groupVerbose := groupCmd.Bool("verbose", false, "show groups")

	supportCmd := flag.NewFlagSet("support", flag.ExitOnError)
	//supportCSI := supportCmd.String("csi", "", "csi number")
	supportTicketList := supportCmd.Bool("list", false, "list tickets")

	capacityCmd := flag.NewFlagSet("capacity", flag.ExitOnError)
	capacityFetch := capacityCmd.Bool("run", false, "fetch capacity")
	capacityShapeOCPUs := capacityCmd.Int("ocpus", 0, "number of ocpus")
	capacityShapeMemory := capacityCmd.Int("memory", 0, "amount of memory")
	capacityShapeType := capacityCmd.String("type", "E4", "use shape type E3, E4, E5, X9, A1, A2 --- OR the ChipSet family AMD, Intel, ARM")
	capacityAD := capacityCmd.String("ad", "", "availability domain")
	capacityFD := capacityCmd.String("fd", "", "fault domain")

	capabilityCmd := flag.NewFlagSet("capability", flag.ExitOnError)
	capabilityFetch := capabilityCmd.Bool("run", false, "fetch capability")
	capabilityShapeType := capabilityCmd.String("type", "E4", "use shape type E3, E4, E5, X9, A1")

	childCmd := flag.NewFlagSet("children", flag.ExitOnError)
	childFetch := childCmd.Bool("run", false, "fetch child tenancies")
	childWrite := childCmd.Bool("write", false, "write  to file")

	objectCmd := flag.NewFlagSet("object", flag.ExitOnError)
	objectFetch := objectCmd.Bool("run", false, "fetch object storage")

	billingCMD := flag.NewFlagSet("billing", flag.ExitOnError)
	billingPath := billingCMD.String("path", "reports", "path to save billing files - default is ./reports")
	billingDownload := billingCMD.Bool("download", false, "download billing files")

	scheduleCmd := flag.NewFlagSet("schedule", flag.ExitOnError)
	scheduleFetch := scheduleCmd.Bool("run", false, "fetch schedule")

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchFetchString := searchCmd.String("searchstring", "", "search string")

	config, err := setup.Getconfig()
	if err != nil {
		//fmt.Printf("%+v\n", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		//fmt.Println("Using profile:", config.ProfileName)
		//fmt.Printf("Config: %v\n", config.ConfigPath)
		return
	}

	//fmt.Printf("os.Args: %v\n", os.Args)

	//fmt.Printf("flag.Args: %v\n", flag.Args())

	// Parse command line arguments

	//fmt.Println("Using profile:", config.ProfileName)
	//fmt.Printf("Config: %v\n", config.ConfigPath)
	utils.PrintSpace()

	switch os.Args[1] {
	case "limits":
		fmt.Println("fetching limits")
		limitCmd.Parse(os.Args[2:])
		fmt.Printf("limitFetch: %v\n", *limitFetch)
		fmt.Printf("limitWrite: %v\n", *limitWrite)
		if *limitFetch {
			provider, client, tenancyID, err := setup.Prep(config)
			regions, _, _, _ := setup.CommonSetup(err, client, tenancyID)
			limits.RunLimits(provider, regions, tenancyID, *limitWrite)
		} else {
			fmt.Println("add -run to run")
		}

	case "compute":
		fmt.Println("fetching compute")
		computeCmd.Parse(os.Args[2:])
		fmt.Printf("computeFetch: %v\n", *computeFetch)
		if *computeFetch {
			provider, client, tenancyID, err := setup.Prep(config)
			regions, compartments, _, _ := setup.CommonSetup(err, client, tenancyID)
			compute.RunCompute(provider, regions, tenancyID, compartments)
		} else {
			fmt.Println("add -run to run")
		}
	case "peeps":
		fmt.Println("fetching users")
		peopleCmd.Parse(os.Args[2:])
		fmt.Printf("peopleFetch: %v\n", peopleFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		helpers.FatalIfError(err)
		peopleresource.GetAllPeople(provider, client, tenancyID, *peopleFetch)

	case "policies":
		fmt.Println("fetching policies")
		policyCmd.Parse(os.Args[2:])
		fmt.Printf("policies: %v\n", *policyFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		helpers.FatalIfError(err)
		_, compartments, _, _ := setup.CommonSetup(err, client, tenancyID)
		peopleresource.GetAllPolicies(provider, client, tenancyID, compartments, *policyFetch, *policyVerbose)

	case "groups":
		fmt.Println("fetching groups")
		groupCmd.Parse(os.Args[2:])
		fmt.Printf("groups: %v\n", *groupFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		helpers.FatalIfError(err)
		peopleresource.Groups(provider, client, tenancyID, *groupFetch)

	case "support":
		fmt.Println("fetching support")
		supportCmd.Parse(os.Args[2:])
		//fmt.Printf("supportCSI: %v\n", *supportCSI)
		fmt.Printf("supportTicketList: %v\n", *supportTicketList)
		provider, client, tenancyID, err := setup.Prep(config)
		_, _, _, homeregion := setup.CommonSetup(err, client, tenancyID)
		//_, compartments, _ := setup.CommonSetup(err, client, tenancyID, false)

		//supportresources.CreateTicket(provider, tenancyID, homeregion, config.CSI)
		//supportresources.GetCSI(provider, tenancyID, homeregion)
		supportresources.ListTickets(provider, tenancyID, homeregion, config.CSI)
		//helpers.FatalIfError(err)

		//supportresources.ListLimitsTickets(provider, tenancyID, homeregion, config.CSI)
		//supportresources.ListBillingTickets(provider, tenancyID, homeregion, config.CSI)

	case "capacity":
		fmt.Println("checking capacity")
		capacityCmd.Parse(os.Args[2:])
		fmt.Printf("capacityFetch: %v\n", *capacityFetch)
		fmt.Printf("capacityShapeType: %v\n", *capacityShapeType)
		fmt.Printf("capacityShapeOCPUs: %v\n", *capacityShapeOCPUs)
		fmt.Printf("capacityShapeMemory: %v\n", *capacityShapeMemory)
		fmt.Printf("capacityAD: %v\n", *capacityAD)
		fmt.Printf("capacityFD: %v\n", *capacityFD)
		provider, client, tenancyID, err := setup.Prep(config)
		regions, compartments, _, _ := setup.CommonSetup(err, client, tenancyID)
		chipsetSlice := []string{"AMD", "INTEL", "ARM"}
		if *capacityShapeOCPUs > 0 || *capacityShapeMemory > 0 {
			if !slices.Contains(chipsetSlice, strings.ToUpper(*capacityShapeType)) {
				capcheck.Check(provider, regions, tenancyID, compartments, *capacityFetch, *capacityShapeOCPUs, *capacityShapeMemory, *capacityShapeType)
			} else {
				capcheck.CheckFAMILY(provider, regions, tenancyID, compartments, *capacityFetch, *capacityShapeOCPUs, *capacityShapeMemory, *capacityShapeType)
			}

		} else {
			fmt.Println("add -ocpus and -memory -type  to run - NOTE: for -type you can use shape type 'E3', 'E4', 'E5', 'E6', 'X9', 'A1', 'A2' --- OR the ChipSet family 'AMD', 'Intel', 'ARM'")

		}

	case "network":
		fmt.Println("fetching network")
		networkCmd.Parse(os.Args[2:])
		fmt.Printf("networkFetch: %v\n", *networkFetch)
		fmt.Printf("network")
		fmt.Printf("networkCIDRFetch %v\n", *networkCIDRFetch)
		fmt.Printf("networkInventoryFetch%v\n", *networkInventoryFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		regions, compartments, _, _ := setup.CommonSetup(err, client, tenancyID)
		network.GetAllVcn(provider, regions, tenancyID, compartments, *networkFetch, *networkCIDRFetch, *networkInventoryFetch)

	case "capability":
		fmt.Println("checking capabilities")
		capabilityCmd.Parse(os.Args[2:])
		fmt.Printf("capabilityFetch: %v\n", *capabilityFetch)
		fmt.Printf("capabilityShapeType: %v\n", *capabilityShapeType)
		provider, client, tenancyID, err := setup.Prep(config)
		regions, compartments, _, _ := setup.CommonSetup(err, client, tenancyID)
		capability.OSSupport(provider, regions, tenancyID, compartments, *capabilityFetch, *capabilityShapeType)

	case "children":
		fmt.Println("checking child tenancies")
		childCmd.Parse(os.Args[2:])
		fmt.Printf("childFetch: %v\n", *childFetch)
		fmt.Printf("childWrite: %v\n", *childWrite)
		provider, client, tenancyID, err := setup.Prep(config)
		_, _, _, homeregion := setup.CommonSetup(err, client, tenancyID)

		children.Children(provider, client, tenancyID, *childFetch, homeregion, config, *childWrite)
		//children.Deets(provider, tenancyID, homeregion, config)

	case "object":
		fmt.Println("checking object storage")
		objectCmd.Parse(os.Args[2:])
		fmt.Printf("objectFetch: %v\n", *objectFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		regions, compartments, _, homeregion := setup.CommonSetup(err, client, tenancyID)
		//oos.GetObjectStorageInfo(provider, regions, tenancyID, compartments, *objectFetch, homeregion)
		oos.ObjectStorageSize(provider, regions, tenancyID, compartments, *objectFetch, homeregion)

	case "billing":
		fmt.Println("checking billing")
		billingCMD.Parse(os.Args[2:])
		fmt.Printf("billingPath: %v\n", &billingPath)
		fmt.Printf("billingDownload: %v\n", *billingDownload)
		provider, client, tenancyID, err := setup.Prep(config)
		_, _, _, homeregion := setup.CommonSetup(err, client, tenancyID)
		billing.Getfiles(provider, tenancyID, homeregion, config, *billingPath, *billingDownload)

	case "schedule":
		fmt.Println("checking schedule")
		scheduleCmd.Parse(os.Args[2:])
		fmt.Printf("scheduleFetch: %v\n", *scheduleFetch)
		provider, client, tenancyID, err := setup.Prep(config)
		regions, compartments, _, homeregion := setup.CommonSetup(err, client, tenancyID)
		scheduler.RunSchedule(provider, regions, tenancyID, compartments, homeregion)

	case "search":
		fmt.Println("checking search")
		searchCmd.Parse(os.Args[2:])
		provider, _, tenancyID, _ := setup.Prep(config)
		resourcesearch.Search(provider, tenancyID, *searchFetchString)

	case "config":
		fmt.Println("checking config")
		checkCmd.Parse(os.Args[2:])
		fmt.Printf("checkRun: %v\n", *checkFetch)
		_, client, tenancyID, err := setup.Prep(config)
		regions, compartments, ads, _ := setup.CommonSetup(err, client, tenancyID)

		if regions == nil {
			fmt.Println("regions is nil")
		} else {
			fmt.Printf("subscribed regions: %v\n", len(regions))
			if *checkFetch {
				for _, region := range regions {
					fmt.Printf("\tRegionKey: %v, name: %v \n", strings.ToLower(*region.RegionKey), *region.RegionName)
				}
			}
		}
		if ads == nil {
			fmt.Println("ads is nil")
		} else {
			fmt.Printf("ads: %v\n", len(ads))
			if *checkFetch {
				for _, ad := range ads {
					fmt.Printf("\tAD: %v\n", *ad.Name)
				}
			}
		}
		if compartments == nil {
			fmt.Println("compartments is nil")
		} else {

			fmt.Printf("compartments: %v\n", len(compartments))
			if *checkFetch {
				for _, comp := range compartments {
					fmt.Printf("\tCompartment Name: %v \n", *comp.Name)

				}
			}
			//fmt.Printf("compartments: %v\n", compartments)
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

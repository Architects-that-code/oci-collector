package main

import (
	setup "check-limits/config"
	flimit "check-limits/limits"
	"check-limits/util"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/limits"
)

// refactor to create CLI using OCI SDK for Go to interact with the OCI API. Use local auth configs but let user specify
// profile to use. The CLI should list the subscribed regions available to the specified profile and identify all the compartments and then loop thru each compartment in each region to query for
// the limits for each service. The CLI should output the limits to a file in the limits directory in the current working directory. The file should be named
func main() {

	limitCmd := flag.NewFlagSet("limits", flag.ExitOnError)
	runLimits := limitCmd.Bool("fetch", false, "fetch limits true or false")

	if *runLimits {
		fmt.Println("Usage of run:")
	}

	err, config := setup.Getconfig()
	if err != nil {
		//fmt.Printf("%+v\n", err)
		slog.Info("%+v\n", err)
		os.Exit(1)
	}

	// Parse command line arguments
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Usage: 'check-oci-limits  limits | compute ' (if you do not pass in any extra arg it only prints your configs info")
		fmt.Println("Using profile:", config.ProfileName)
		fmt.Printf("Config: %v\n", config.ConfigPath)
		return
	}

	fmt.Println("Using profile:", config.ProfileName)
	fmt.Printf("Config: %v\n", config.ConfigPath)
	util.PrintSpace()
	provider := common.CustomProfileConfigProvider(config.ConfigPath, config.ProfileName)
	slog.Debug("provider: %v\n", provider)

	client, err := identity.NewIdentityClientWithConfigurationProvider(provider)
	util.PrintSpace()
	helpers.FatalIfError(err)
	slog.Debug("client %v\n", client)

	tenancyID, err := provider.TenancyOCID()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	slog.Debug("TenancyOCID: %v\n", tenancyID)

	ads := getADs(tenancyID, err, client)
	util.PrintSpace()
	//fmt.Printf("ADs: %v\n", ads)
	slog.Debug("ads: ", ads)

	// getallregions
	// getall compartments
	var wgDataPrep = sync.WaitGroup{}

	wgDataPrep.Add(2)

	go func() {
		defer wgDataPrep.Done()
		compartments := setup.Getcompartments(err, client, tenancyID)
		for _, comp := range compartments {
			fmt.Printf("Compartment Name: %v CompartmentID: %v\n", *comp.Name, *comp.CompartmentId)
		}
	}()
	var regions []identity.RegionSubscription
	go func() {
		defer wgDataPrep.Done()
		regions = setup.Getregions(err, client, tenancyID)
		for _, region := range regions {
			fmt.Printf("Region: %v\n", *region.RegionName)
		}
		util.PrintSpace()
		slog.Debug("List of regions:", regions)
	}()
	wgDataPrep.Wait()

	limitsClient, err := limits.NewLimitsClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	util.PrintSpace()
	slog.Debug("limitsClient: %v\n", limitsClient)
	//for _, region := range []string{"us-ashburn-1"} {
	//	reg := region

	//create datastructures that will hold all results
	var Datapile []collector

	//localReg := []string{"us-ashburn-1"}
	fmt.Printf("regions: %v\n", regions)
	var wg_regional = sync.WaitGroup{}
	var regionalSlices = make(chan []collector, len(regions))

	counter := 0
	counterLock := &sync.Mutex{}

	//for _, region := range localReg {
	//reg := region
	for _, region := range regions {
		reg := *region.RegionName

		counterLock.Lock()
		counter++
		currentGoroutineID := counter
		counterLock.Unlock()

		wg_regional.Add(1)
		go func(reg string, goroutineID int) {
			defer wg_regional.Done()
			var localDatapile []collector

			services := flimit.GetServices(limitsClient, err, tenancyID, reg)
			for _, s := range services.Items {
				svc := s.Name
				vals := flimit.GetLimitsForService(err, limitsClient, tenancyID, *svc)
				for _, v := range vals.Items {
					limitName := v.Name
					ad := v.AvailabilityDomain
					var avails = limits.GetResourceAvailabilityResponse{}
					if ad == nil {
						avails = flimit.GetLimitsAvailRegionScoped(err, limitsClient, tenancyID, *svc, *v.Name)
					} else {
						avails = flimit.GetLimitAvailADScoped(err, limitsClient, tenancyID, *svc, *v.Name, *ad)
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
					var r = collector{
						region:    reg,
						service:   *svc,
						limitname: *limitName,
						avail:     *avail,
						used:      *used,
					}

					localDatapile = append(localDatapile, r)
					fmt.Printf("goroutineID: %v region: %v service: %v valLimitName: %v avail: %v used: %v\n", goroutineID, reg, *svc, *limitName, *avail, *used)
				}

			}
			regionalSlices <- localDatapile
		}(reg, currentGoroutineID)

	}
	wg_regional.Wait()
	close(regionalSlices)
	for slice := range regionalSlices {
		Datapile = append(Datapile, slice...)
	}
	for _, dp := range Datapile {
		fmt.Println(dp)
	}
	fmt.Println(len(Datapile))

}

func getADs(tenancyID string, err error, client identity.IdentityClient) identity.ListAvailabilityDomainsResponse {
	request := identity.ListAvailabilityDomainsRequest{
		CompartmentId: &tenancyID,
	}
	r, err := client.ListAvailabilityDomains(context.Background(), request)
	helpers.FatalIfError(err)
	return r
}

type collector struct {
	region    string
	service   string
	limitname string
	avail     int64
	used      int64
}

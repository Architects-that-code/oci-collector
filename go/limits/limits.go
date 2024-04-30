package limits

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/limits"

	util "check-limits/util"
)

func GetLimitsAvailRegionScoped(err error, limitsClient limits.LimitsClient, compartment string, svc string, limitName string) limits.GetResourceAvailabilityResponse {
	req := limits.GetResourceAvailabilityRequest{
		ServiceName:     &svc,
		LimitName:       &limitName,
		CompartmentId:   &compartment,
		RequestMetadata: common.RequestMetadata{},
	}
	avail, err := limitsClient.GetResourceAvailability(
		context.Background(), req)
	helpers.FatalIfError(err)
	return avail
}
func GetLimitAvailADScoped(err error, limitsClient limits.LimitsClient, compartment string, svc string, limitName string, ad string) limits.GetResourceAvailabilityResponse {
	var req = limits.GetResourceAvailabilityRequest{}
	req = limits.GetResourceAvailabilityRequest{
		ServiceName:        &svc,
		LimitName:          &limitName,
		CompartmentId:      &compartment,
		AvailabilityDomain: &ad,
		RequestMetadata:    common.RequestMetadata{},
	}

	avail, err := limitsClient.GetResourceAvailability(
		context.Background(), req)

	helpers.FatalIfError(err)
	return avail
}

func GetLimitDefs(err error, limitsClient limits.LimitsClient, tenancyID string, svc string) limits.ListLimitDefinitionsResponse {
	defs, err := limitsClient.ListLimitDefinitions(context.Background(), limits.ListLimitDefinitionsRequest{
		CompartmentId:   &tenancyID,
		ServiceName:     &svc,
		RequestMetadata: common.RequestMetadata{},
	})
	helpers.FatalIfError(err)
	return defs
}

func GetLimitsForService(err error, limitsClient limits.LimitsClient, tenancyID string, svc string) limits.ListLimitValuesResponse {
	vals, err := limitsClient.ListLimitValues(context.Background(), limits.ListLimitValuesRequest{
		CompartmentId:   &tenancyID,
		ServiceName:     &svc,
		RequestMetadata: common.RequestMetadata{},
	})
	helpers.FatalIfError(err)
	return vals
}

func GetServices(limitsClient limits.LimitsClient, err error, tenancyID string, region string) limits.ListServicesResponse {
	limitsClient.SetRegion(region)
	util.PrintSpace()
	slog.Debug("limitsClientUPDATED: \n", limitsClient.Endpoint())
	services, err := limitsClient.ListServices(context.Background(), limits.ListServicesRequest{
		CompartmentId:   &tenancyID,
		SortBy:          "",
		SortOrder:       "",
		Limit:           nil,
		Page:            nil,
		OpcRequestId:    nil,
		RequestMetadata: common.RequestMetadata{},
	})
	helpers.FatalIfError(err)
	return services
}

type LimitsCollector struct {
	Region    string
	Service   string
	Limitname string
	Avail     int64
	Used      int64
}

func RunLimits(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string) {
	limitsClient, err := limits.NewLimitsClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	util.PrintSpace()
	fmt.Printf("limitsClient: %v\n", limitsClient)

	var Datapile []LimitsCollector

	//localReg := []string{"us-ashburn-1"}
	fmt.Printf("regions: %v\n", regions)

	var wg_regional = sync.WaitGroup{}
	var regionalSlices = make(chan []LimitsCollector, len(regions))

	counter := 0
	counterLock := &sync.Mutex{}

	for _, region := range regions {
		reg := *region.RegionName

		counterLock.Lock()
		counter++
		currentGoroutineID := counter
		counterLock.Unlock()

		wg_regional.Add(1)
		go func(reg string, goroutineID int) {
			defer wg_regional.Done()
			var localDatapile []LimitsCollector

			services := GetServices(limitsClient, err, tenancyID, reg)
			for _, s := range services.Items {
				svc := s.Name
				vals := GetLimitsForService(err, limitsClient, tenancyID, *svc)
				for _, v := range vals.Items {
					limitName := v.Name
					ad := v.AvailabilityDomain
					var avails = limits.GetResourceAvailabilityResponse{}
					if ad == nil {
						avails = GetLimitsAvailRegionScoped(err, limitsClient, tenancyID, *svc, *v.Name)
					} else {
						avails = GetLimitAvailADScoped(err, limitsClient, tenancyID, *svc, *v.Name, *ad)
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
					var r = LimitsCollector{
						Region:    reg,
						Service:   *svc,
						Limitname: *limitName,
						Avail:     *avail,
						Used:      *used,
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

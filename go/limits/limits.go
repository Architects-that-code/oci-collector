package limits

import (
	"context"
	"log/slog"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
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

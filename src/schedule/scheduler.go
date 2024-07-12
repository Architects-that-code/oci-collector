package schedule

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/resourcescheduler"
)

func RunSchedule(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, compartment []identity.Compartment, capacityShapeType string) {

	client, err := resourcescheduler.NewScheduleClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	// Create a request and dependent object(s).

	req := resourcescheduler.ListSchedulesRequest{CompartmentId: common.String("ocid1.test.oc1..<unique_ID>EXAMPLE-compartmentId-Value"),
		DisplayName:    common.String("EXAMPLE-displayName-Value"),
		SortBy:         resourcescheduler.ListSchedulesSortByLifecyclestate,
		LifecycleState: resourcescheduler.ScheduleLifecycleStateDeleted,
		Limit:          common.Int(247),
		OpcRequestId:   common.String("WOY0IFHTFZ6PUISKVETD<unique_ID>"),
		Page:           common.String("EXAMPLE-page-Value"),
		ScheduleId:     common.String("ocid1.test.oc1..<unique_ID>EXAMPLE-scheduleId-Value"),
		SortOrder:      resourcescheduler.ListSchedulesSortOrderDesc}

	// Send the request using the service client
	resp, err := client.ListSchedules(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Println(resp)
}

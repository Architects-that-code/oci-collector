package supportresources

import (
	"context"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/v65/cims"
	common "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
)

func CreateTicket() {
	// create ticket
}

/*
oci support incident create -
-compartment-id ocid1.tenancy.oc1..aaaaaaaaubrkzed3mzqxtsxx4qnfgmcmoh5mm7r6xxxxxxxx --description "test"
--csi 2197008 --problem-type "TECH" --severity "MEDIUM" --title "Test Support Request Broken Node blah"
--homeregion us-ashburn-1 --ocid ocid1.user.oc1..aaaaaaaanclxin474nk5w6jtfuo5rhwfu3dycpnjy5jgkroipp36xiq56s4q
*/
func ListTickets(provider common.ConfigurationProvider, tenancyID string, homeRegion string, CSI string) {
	fmt.Println("ListTickets")
	client, err := cims.NewIncidentClientWithConfigurationProvider(provider)
	var user_ocid, _ = provider.UserOCID()
	req := cims.ListIncidentsRequest{
		CompartmentId:  &tenancyID,
		Limit:          common.Int(100),
		Ocid:           common.String(user_ocid),
		Csi:            &CSI,
		LifecycleState: cims.ListIncidentsLifecycleStateActive,
		ProblemType:    common.String("TECH"),
	}

	fmt.Printf("Tech req: %v\n", req)

	resp, err := client.ListIncidents(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	//fmt.Println(resp)

	for _, incident := range resp.Items {
		fmt.Println("Incident: ", *incident.Key, *incident.Ticket.Title, *&incident.Ticket.Severity, *&incident.Ticket.LifecycleDetails, time.Unix(int64(*incident.Ticket.TimeUpdated), 0))
		//fmt.Printf("Incident: %v\n", incident)

	}
	fmt.Printf("Incidents: %v\n", len(resp.Items))
	fmt.Println("ListTickets -end")
}

func ListLimitsTickets(provider common.ConfigurationProvider, tenancyID string, homeRegion string, CSI string) {
	fmt.Println("ListLimitsTickets")
	client, err := cims.NewIncidentClientWithConfigurationProvider(provider)
	var user_ocid, _ = provider.UserOCID()
	req := cims.ListIncidentsRequest{
		CompartmentId:  &tenancyID,
		Limit:          common.Int(100),
		Ocid:           common.String(user_ocid),
		Csi:            &CSI,
		LifecycleState: cims.ListIncidentsLifecycleStateActive,
		ProblemType:    common.String("LIMIT"),
		Homeregion:     &homeRegion,
	}
	fmt.Printf("Limit req: %v\n", req)

	resp, err := client.ListIncidents(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	//fmt.Println(resp)

	for _, incident := range resp.Items {
		fmt.Println("LIMITS Incident: ", *incident.Key, *incident.Ticket.TicketNumber, *incident.Ticket.Title, time.Unix(int64(*incident.Ticket.TimeUpdated), 0), *&incident.Ticket.LifecycleDetails)
		//fmt.Printf("Incident: %v\n", incident)
	}

	fmt.Printf("LIMITS Incidents: %v\n", len(resp.Items))
	fmt.Println("ListLimitsTickets -end")
}

func ListBillingTickets(provider common.ConfigurationProvider, tenancyID string, homeRegion string, CSI string) {
	fmt.Println("ListBILLINGTickets")
	client, err := cims.NewIncidentClientWithConfigurationProvider(provider)
	var user_ocid, _ = provider.UserOCID()
	req := cims.ListIncidentsRequest{
		CompartmentId:  &tenancyID,
		Limit:          common.Int(100),
		Ocid:           common.String(user_ocid),
		Csi:            &CSI,
		LifecycleState: cims.ListIncidentsLifecycleStateActive,
		ProblemType:    common.String("ACCOUNT")}

	resp, err := client.ListIncidents(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	//fmt.Println(resp)

	for _, incident := range resp.Items {
		fmt.Println("ACCOUNT Incident: ", *incident.Key)
	}

	fmt.Printf("ACCOUNT Incidents: %v\n", len(resp.Items))
	fmt.Println("ListACCOUNTickets -end")
}
func CloseTicket() {
	// close ticket
}

func GetCSI(provider common.ConfigurationProvider, tenancyID string, homeRegion string) {
	fmt.Println("GetCSI")
	client, err := cims.NewIncidentClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	// Create a request and dependent object(s).

	req := cims.GetCsiNumberRequest{

		Homeregion: &homeRegion,
		TenantId:   &tenancyID,
		Region:     &homeRegion}

	// Send the request using the service client
	resp, err := client.GetCsiNumber(context.Background(), req)
	fmt.Printf("response: %v\n", resp)
	helpers.FatalIfError(err)

	// Retrieve value from the response.

	fmt.Println(resp)
	fmt.Println("GetCSI -end")
}

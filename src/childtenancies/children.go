package children

import (
	config "check-limits/config"
	utils "check-limits/util"
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/governancerulescontrolplane"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/tenantmanagercontrolplane"
	"gopkg.in/yaml.v2"
)

func getOrgID(tenancyOCID string, provider common.ConfigurationProvider) (string, error) {
	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	if err != nil {
		return "", err
	}

	request := tenantmanagercontrolplane.ListOrganizationsRequest{
		CompartmentId: common.String(tenancyOCID),
	}

	response, err := client.ListOrganizations(context.Background(), request)
	if err != nil {
		return "", err
	}

	if len(response.Items) == 0 {
		return "", fmt.Errorf("no organization found for this tenancy")
	}

	// Returns the ID of the first associated organization
	fmt.Printf("orgid: %v\n", *response.Items[0].Id)
	return *response.Items[0].Id, nil

}

func Children(provider common.ConfigurationProvider, passThruClient identity.IdentityClient, tenancyID string, childFetch bool, homeregion string, config config.Config, write bool) {
	fmt.Println("checking child tenancies Children")
	orgID, err := getOrgID(tenancyID, provider)

	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	req := tenantmanagercontrolplane.GetOrganizationRequest{

		OrganizationId: common.String(orgID),
	}
	helpers.FatalIfError(err)
	// Send the request using the service client
	resp, err := client.GetOrganization(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	Organization := resp.Organization
	fmt.Printf("\tOrganization: %v\n", Organization)
	Deets(provider, tenancyID, homeregion, config)
	getOutstandingInvites(provider, tenancyID, homeregion, config)

	//getOutstandingInvites(provider, tenancyID, homeregion, config)

	// does this tenancy have children
	children := GetChildTenancies(provider, passThruClient, tenancyID, config, write)
	fmt.Printf("\tcount ChildTenancies: %v\n", len(children))

}

func GetChildTenancies(provider common.ConfigurationProvider, passThruClient identity.IdentityClient, tenancyID string, config config.Config, write bool) []tenantmanagercontrolplane.OrganizationTenancySummary {
	fmt.Println("checking child tenancies GetChildTenancies")
	var allTenancies []tenantmanagercontrolplane.OrganizationTenancySummary
	var tenancies []TenancyCollector

	orgID, err := getOrgID(tenancyID, provider)

	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)
	req := tenantmanagercontrolplane.ListOrganizationTenanciesRequest{
		Limit:          common.Int(1000),
		OrganizationId: common.String(orgID),
	}

	// Send the request using the service client
	resp, err := client.ListOrganizationTenancies(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Printf("count child OrganizationTenancies: %v\n", len(resp.Items))
	fmt.Printf(" checking for policy compliance at child tenancy level s=good e=error \n")
	for _, tenancy := range resp.Items {
		//fmt.Printf("child OrganizationTenancy: %v\n", tenancy)
		//fmt.Printf("child OrganizationTenancy: %v\n", *tenancy.TenancyId)
		if tenancy.TenancyId != nil && tenancy.Name != nil {
			//fmt.Printf("child OrganizationTenancy: %v\n", *&tenancy.LifecycleState)
			//fmt.Printf("has tags %v\n", getchildTAGS(passThruClient, *tenancy.TenancyId))
			var tc = TenancyCollector{
				TenancyId:         *tenancy.TenancyId,
				TenancyName:       *tenancy.Name,
				TenancyConfigured: getchildcompartments(passThruClient, *tenancy.TenancyId),
				GovernanceStatus:  string(tenancy.GovernanceStatus),
				LifecycleState:    string(tenancy.LifecycleState),
			}
			tenancies = append(tenancies, tc)
			allTenancies = append(allTenancies, tenancy)

		}

		// cross tenancy admission does n
		// ot apply to Identity services
		//getAllPeople(provider, passThruClient, *tenancy.TenancyId, true)
		//getchildcompartments(passThruClient, *tenancy.TenancyId)
	}

	//jsonData, _ := utils.ToJSON(tenancies)
	//fmt.Println(string(jsonData))

	// Print the list of child tenancies
	utils.WriteToFile("collector-childTenancy_counts_lifecycleStates.txt", []byte(lifeCycleStats(tenancies)))

	if write {
		fmt.Println("-")
		// // Write to CSV
		writetenanciestoFile(tenancies)

		yamlData, err := yaml.Marshal(tenancies)
		if err != nil {
			fmt.Println("Error marshaling to YAML:", err)
		}
		utils.WriteToFile("collector-childtenancies.yaml", []byte(yamlData))

		//fmt.Println(string(yamlData))
	}
	return allTenancies
}

func lifeCycleStats(tenancies []TenancyCollector) string {
	// Create a map to store the count of each lifecycle state
	lifecycleStats := make(map[string]int)

	// Iterate over the tenancies and count the lifecycle states
	for _, tenancy := range tenancies {
		lifecycleStats[tenancy.LifecycleState]++
	}

	var result string
	result += "Lifecycle State Stats:\n"
	for state, count := range lifecycleStats {
		result += fmt.Sprintf("%4d %s\n", count, state)
	}
	return result
}

func writetenanciestoFile(tenancies []TenancyCollector) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	file, err := os.Create(homedir + "/collector-actualChildren.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Comma = ','
	writer.UseCRLF = true
	defer writer.Flush()
	writer.Write([]string{"tenancy_name", "tenancy_ocid"})
	for _, tenancy := range tenancies {
		//fmt.Print("\"" + tenancy.TenancyId + "\",")
		writer.Write([]string{tenancy.TenancyName, tenancy.TenancyId})
	}
	fmt.Println("\nwrote to file")
}

func Deets(provider common.ConfigurationProvider, tenancyID string, homeregion string, config config.Config) {
	fmt.Println("checking child tenancies Getting Governance rules")
	client, err := tenantmanagercontrolplane.NewOrganizationClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	orgID, err := getOrgID(tenancyID, provider)

	req := tenantmanagercontrolplane.GetOrganizationTenancyRequest{
		OrganizationId: common.String(orgID),
		TenancyId:      common.String(tenancyID),
	}

	resp, err := client.GetOrganizationTenancy(context.Background(), req)
	helpers.FatalIfError(err)
	OrganizationTenancy := resp.OrganizationTenancy

	gclient, err := governancerulescontrolplane.NewGovernanceRuleClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	listReq := governancerulescontrolplane.ListGovernanceRulesRequest{
		CompartmentId:  common.String(tenancyID),
		LifecycleState: governancerulescontrolplane.ListGovernanceRulesLifecycleStateActive,
	}

	listResp, err := gclient.ListGovernanceRules(context.Background(), listReq)

	helpers.FatalIfError(err)
	//fmt.Printf("list response Governance: %v\n", listResp)
	for _, rule := range listResp.Items {

		fmt.Printf("\t rule: name %v, id %v,  type %v \n", *rule.DisplayName, *rule.Id, *&rule.Type)

		/*
			delReq := governancerulescontrolplane.DeleteGovernanceRuleRequest{
				GovernanceRuleId: rule.Id,
			}
			delResp, err := gclient.DeleteGovernanceRule(context.Background(), delReq)
			helpers.FatalIfError(err)
			fmt.Printf("delete response Governance: %v\n", delResp)
		*/

	}

	fmt.Printf("DEETS: Organization Tenancy: %v\n", OrganizationTenancy)

}

// this
func getchildcompartments(client identity.IdentityClient, tenancyID string) bool {

	var allCompartments []identity.Compartment
	allCompartments = append(allCompartments, identity.Compartment{Id: &tenancyID,
		Name: common.String("root")})

	req := identity.ListCompartmentsRequest{
		AccessLevel:            identity.ListCompartmentsAccessLevelAny,
		CompartmentId:          &tenancyID,
		CompartmentIdInSubtree: common.Bool(true),
		SortBy:                 identity.ListCompartmentsSortByName,
		SortOrder:              identity.ListCompartmentsSortOrderAsc,
		LifecycleState:         identity.CompartmentLifecycleStateActive,
		Limit:                  common.Int(208),
	}
	for {
		respComp, err := client.ListCompartments(context.Background(), req)
		if err != nil {
			//fmt.Printf("error %v\n", tenancyID)
			fmt.Print("e")
			return false
		} else {
			//fmt.Printf("success %v\n", tenancyID)
			fmt.Print("s")
			allCompartments = append(allCompartments, respComp.Items...)
			if respComp.OpcNextPage != nil {
				req.Page = respComp.OpcNextPage
			} else {
				break
			}

		}

	}
	//fmt.Printf("List of compartments: %v", allCompartments)

	return len(allCompartments) > 0

}
func getchildTAGS(client identity.IdentityClient, tenancyID string) bool {

	var allTags []identity.TagNamespaceSummary

	req := identity.ListTagNamespacesRequest{
		CompartmentId:          &tenancyID,
		IncludeSubcompartments: common.Bool(false),
		LifecycleState:         identity.TagNamespaceLifecycleStateActive,
		Limit:                  common.Int(1000),
	}

	// Send the request using the service client
	resp, err := client.ListTagNamespaces(context.Background(), req)
	if err != nil {
		//fmt.Printf("error %v\n", tenancyID)
		return false
	} else {
		fmt.Printf("has tags %v\n", tenancyID)
	}

	allTags = append(allTags, resp.Items...)

	// Retrieve value from the response.
	//fmt.Println(resp)

	return len(allTags) > 0
}

// this may never work as IDENTITY service does not FULLY  support cross-tenancy ACCESS at this time
func getAllPeople(provider common.ConfigurationProvider, passThruClient identity.IdentityClient, tenancyID string, showUsers bool) []identity.User {
	fmt.Printf("Showusers: %v \n", showUsers)

	var allUsers []identity.User

	req := identity.ListUsersRequest{
		CompartmentId: &tenancyID,
		Limit:         common.Int(10000),
	}

	for {
		resp, err := passThruClient.ListUsers(context.Background(), req)
		if err != nil {
			fmt.Printf("error %v\n", tenancyID)
			break
		} else {
			fmt.Printf("success %v\n", tenancyID)
			allUsers = append(allUsers, resp.Items...)
			if resp.OpcNextPage != nil {
				req.Page = resp.OpcNextPage
			} else {
				break
			}

		}
		fmt.Printf("users returned %v\n", len(allUsers))
		if showUsers {
			for _, user := range allUsers {
				n := *user.Name
				tc := *user.TimeCreated

				fmt.Printf("User: %s\t Created: %s \t \n", n, tc)
			}
		}
	}
	fmt.Printf("Showusers: end \n")
	return allUsers
}

func getOutstandingInvites(provider common.ConfigurationProvider, tenancyID string, homeregion string, config config.Config) {
	fmt.Printf("\nchecking child tenancies Getting Outstanding Invites\n")
	client, err := tenantmanagercontrolplane.NewSenderInvitationClientWithConfigurationProvider(provider)

	helpers.FatalIfError(err)

	req := tenantmanagercontrolplane.ListSenderInvitationsRequest{
		CompartmentId:  common.String(tenancyID),
		Limit:          common.Int(1000),
		Status:         tenantmanagercontrolplane.ListSenderInvitationsStatusAccepted,
		LifecycleState: tenantmanagercontrolplane.ListSenderInvitationsLifecycleStateInactive,
	}
	fmt.Printf("request is: %v\n", req)

	resp, err := client.ListSenderInvitations(context.Background(), req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	fmt.Println("number of outstanding invites: ", len(resp.Items))

	/*
			for _, invite := range resp.Items {


				fmt.Printf("Invite: %v\n", invite)


					fmt.Printf("InviteID: %v\n", *invite.Id)
					fmt.Printf("LifecycleState: %v\n", *&invite.LifecycleState)
					fmt.Printf("Status: %v\n", *&invite.Status)

					fmt.Printf("recip email: %v\n", *&invite.RecipientEmailAddress)
					fmt.Printf("timecreated: %v\n", *invite.TimeCreated)

		}
	*/

}
